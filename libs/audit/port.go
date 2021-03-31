package audit

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ullaakut/nmap/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/sync/errgroup"
)

const portAuditName string = "PORT"

type ScanType func(ctx context.Context, target *PortTarget) ([]nmap.Port, error)

// comparePorts compares if user defined ports match open/closed nmap ports
//
// comparePorts(
// 	[]string{"tcp/22", "tcp/80", "tcp/443"},
// 	[]string{"tcp/22", "tcp/80", "tcp/443"},
// ) == true
func comparePorts(userPorts, nmapPorts []string) bool {
	// convert nmapPorts to a map
	nmapPortsSet := make(map[string]bool)
	for _, p := range nmapPorts {
		nmapPortsSet[p] = true
	}

	// check if user defined ports are present in nmap discovered ports
	for _, port := range userPorts {
		// check if port has "-" to denote a range
		//
		// 0 - n ports need to be open/closed in the range.
		// n denotes number of ports in the range
		//
		// services such as Mosh needs ports 60000 - 61000 to be open
		// so that it can support up to 1000 simultaneous connections.
		// However, port 60000 will only start listening when a new mosh
		// process is created to handle a user's connection.
		if strings.Contains(port, "-") {
			a := strings.FieldsFunc(port, func(r rune) bool {
				return r == '/' || r == '-'
			})
			protocol := a[0]
			start, _ := strconv.Atoi(a[1])
			end, _ := strconv.Atoi(a[2])

			for i := start; i < end+1; i++ {
				portToLookup := fmt.Sprintf("%s/%d", protocol, i)

				// remove port from nmapPortsSet as it's expected to be open/closed from
				// the range definition
				delete(nmapPortsSet, portToLookup)
			}
		} else {
			_, found := nmapPortsSet[port]
			if found {
				// delete port from nmapPortsSet. port exists in the nmapPorts
				delete(nmapPortsSet, port)
			} else {
				return false
			}
		}
	}

	if len(nmapPortsSet) > 0 {
		return false
	}

	return true
}

// readCommonPorts parses comma delimited ports from a file
func readCommonPorts(commonPortsPath string) ([]string, error) {
	commonPortsFilePath := filepath.Join(
		filepath.Dir(auditFilePath),
		commonPortsPath,
	)

	file, err := os.Open(commonPortsFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ports := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		port := strings.TrimSpace(scanner.Text())
		if port != "" {
			ports = append(ports, port)
		}
	}

	err = scanner.Err()

	return ports, nil
}

// tcpScan runs a TCP scan on the target
func tcpScan(ctx context.Context, t *PortTarget) ([]nmap.Port, error) {
	commonPorts, err := readCommonPorts(t.Group.CommonPortsPath.TCP)
	if err != nil {
		return nil, err
	}

	ports, err := runScanner(
		nmap.WithContext(ctx),
		nmap.WithTargets(t.Host),
		nmap.WithSkipHostDiscovery(),
		nmap.WithPorts(commonPorts...),
		nmap.WithSYNScan(), // requires root privileges
	)
	if err != nil {
		return nil, err
	}

	return ports, nil
}

// udpScan runs a UDP scan on the target
func udpScan(ctx context.Context, t *PortTarget) ([]nmap.Port, error) {
	commonPorts, err := readCommonPorts(t.Group.CommonPortsPath.UDP)
	if err != nil {
		return nil, err
	}

	ports, err := runScanner(
		nmap.WithContext(ctx),
		nmap.WithTargets(t.Host),
		nmap.WithSkipHostDiscovery(),
		nmap.WithPorts(commonPorts...),
		nmap.WithUDPScan(), // requires root privileges
		nmap.WithVersionIntensity(0),
	)
	if err != nil {
		return nil, err
	}

	return ports, nil
}

func runScanner(options ...func(*nmap.Scanner)) ([]nmap.Port, error) {
	var (
		resultBytes []byte
		errorBytes  []byte
		ports       []nmap.Port
	)

	scanner, err := nmap.NewScanner(options...)
	if err != nil {
		return nil, err
	}

	// Executes asynchronously, allowing results to be streamed in real time.
	if err := scanner.RunAsync(); err != nil {
		return nil, err
	}

	// Connect to stdout and stderr of scanner.
	stdout := scanner.GetStdout()
	stderr := scanner.GetStderr()

	// Goroutine to watch for stdout and print to screen. Additionally it stores
	// the bytes intoa variable for processiing later.
	go func() {
		for stdout.Scan() {
			resultBytes = append(resultBytes, stdout.Bytes()...)
		}
	}()

	// Goroutine to watch for stderr and print to screen. Additionally it stores
	// the bytes intoa variable for processiing later.
	go func() {
		for stderr.Scan() {
			errorBytes = append(errorBytes, stderr.Bytes()...)
		}
	}()

	// Blocks main until the scan has completed.
	if err := scanner.Wait(); err != nil {
		return nil, err
	}

	// Parsing the results into corresponding structs.
	result, err := nmap.Parse(resultBytes)
	if err != nil {
		return nil, err
	}

	// Parsing the results into the NmapError slice of our nmap Struct.
	result.NmapErrors = strings.Split(string(errorBytes), "\n")

	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		ports = append(ports, host.Ports...)
	}

	return ports, nil
}

// PortTarget holds information about a host to be scanned
type PortTarget struct {
	Host          string
	Group         *PortTargetGroup
	ScanInfo      []nmap.Port
	ScanInfoError error
}

// Scan performs a port scan on a host
func (t *PortTarget) Scan() {
	scans := func(ctx context.Context) ([]nmap.Port, error) {
		var (
			mutex sync.Mutex
			ports []nmap.Port
		)

		g, ctx := errgroup.WithContext(ctx)

		scanTypes := []ScanType{tcpScan, udpScan}

		for _, scanType := range scanTypes {
			scanType := scanType // https://golang.org/doc/faq#closures_and_goroutines
			g.Go(func() error {
				scanPorts, err := scanType(ctx, t)
				if err == nil {
					mutex.Lock()
					defer mutex.Unlock()

					ports = append(ports, scanPorts...)
				}

				return err
			})
		}

		if err := g.Wait(); err != nil {
			return nil, err
		}

		return ports, nil
	}

	timeout, err := time.ParseDuration(t.Group.Timeout)
	if err != nil {
		t.ScanInfoError = err
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ports, err := scans(ctx)
	if err != nil {
		t.ScanInfoError = err
	}

	t.ScanInfo = ports
}

// Result constructs results output for a given port scan
func (t *PortTarget) Result() []*AuditResult {
	var results []*AuditResult
	var openPorts []string
	var closedPorts []string

	if t.ScanInfoError != nil {
		res := &AuditResult{
			Type:          portAuditName,
			Status:        Error,
			StatusMessage: fmt.Sprintf("%s: %s", t.Host, t.ScanInfoError.Error()),
		}
		results = append(results, res)
		return results
	}

	for _, port := range t.ScanInfo {
		protocolPort := fmt.Sprintf("%s/%d", port.Protocol, port.ID)
		switch port.State.State {
		case string(nmap.Open):
			openPorts = append(openPorts, protocolPort)
		case string(nmap.Closed):
			closedPorts = append(closedPorts, protocolPort)
		}
	}

	if len(t.Group.AllowList) > 0 {
		if comparePorts(t.Group.AllowList, openPorts) {
			statusMsg := fmt.Sprintf(
				"%s has all allowed ports %v open", t.Host, t.Group.AllowList,
			)
			res := &AuditResult{
				Type:          portAuditName,
				Status:        Pass,
				StatusMessage: statusMsg,
			}
			results = append(results, res)
		} else {
			statusMsg := fmt.Sprintf(
				"%s has %v open ports but expected %v ports to be open",
				t.Host, openPorts, t.Group.AllowList,
			)
			res := &AuditResult{
				Type:          portAuditName,
				Status:        Fail,
				StatusMessage: statusMsg,
			}
			results = append(results, res)
		}
	} else if len(t.Group.BlockList) > 0 {
		if comparePorts(t.Group.BlockList, closedPorts) {
			statusMsg := fmt.Sprintf(
				"%s has all blocked ports %v closed", t.Host, t.Group.BlockList,
			)
			res := &AuditResult{
				Type:          portAuditName,
				Status:        Pass,
				StatusMessage: statusMsg,
			}
			results = append(results, res)
		} else {
			statusMsg := fmt.Sprintf(
				"%s has %v closed ports but expected %v ports to be closed",
				t.Host, closedPorts, t.Group.BlockList,
			)
			res := &AuditResult{
				Type:          portAuditName,
				Status:        Fail,
				StatusMessage: statusMsg,
			}
			results = append(results, res)
		}
	} else {
		statusMsg := fmt.Sprintf(
			"%s does not have any specified allowlist or blocklist", t.Host,
		)
		res := &AuditResult{
			Type:          portAuditName,
			Status:        Fail,
			StatusMessage: statusMsg,
		}
		results = append(results, res)
	}

	return results
}

type PortTargetGroup struct {
	Timeout         string     `mapstructure:"timeout"`
	AllowList       []string   `mapstructure:"allowlist"`
	BlockList       []string   `mapstructure:"blocklist"`
	Discovery       *Discovery `mapstructure:"discovery"`
	CommonPortsPath struct {
		TCP string `mapstructure:"tcp"`
		UDP string `mapstructure:"udp"`
	} `mapstructure:"common_ports_path"`
}

func (tg *PortTargetGroup) Scan() ([]*AuditResult, error) {
	var portScanResults []*AuditResult
	var wg sync.WaitGroup
	var mutex sync.Mutex

	hosts, err := tg.Discovery.GetHosts()
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		target := &PortTarget{
			Host:  host,
			Group: tg,
		}

		handler := func(results []*AuditResult, err error) {
			mutex.Lock()
			defer mutex.Unlock()

			portScanResults = append(portScanResults, results...)
		}

		wg.Add(1)

		go func(target *PortTarget, hander AuditScanHandler) {
			defer wg.Done()

			target.Scan()
			results := target.Result()
			handler(results, nil)
		}(target, handler)
	}

	wg.Wait()

	return portScanResults, nil
}

type PortScan struct {
	PortTargetGroups []*PortTargetGroup
}

// Load decodes yaml into struct
func (ps *PortScan) Load(input interface{}) error {
	err := mapstructure.Decode(input, &ps.PortTargetGroups)
	return err
}

// Scan scans hosts
func (ps *PortScan) Scan() ([]*AuditResult, error) {
	var portScanResults []*AuditResult
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var finalErr error

	for _, targetGroup := range ps.PortTargetGroups {
		handler := func(results []*AuditResult, err error) {
			mutex.Lock()
			defer mutex.Unlock()

			portScanResults = append(portScanResults, results...)
			finalErr = err
		}

		wg.Add(1)

		go func(tg *PortTargetGroup, handler AuditScanHandler) {
			defer wg.Done()

			results, err := tg.Scan()
			handler(results, err)
		}(targetGroup, handler)
	}

	wg.Wait()

	return portScanResults, finalErr
}

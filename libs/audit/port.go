package audit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ullaakut/nmap/v2"
	"github.com/mitchellh/mapstructure"
)

const portAuditName string = "PORT"

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
		if strings.Contains(port, "-") {
			a := strings.FieldsFunc(port, func(r rune) bool {
				return r == '/' || r == '-'
			})
			protocol := a[0]
			start, _ := strconv.Atoi(a[1])
			end, _ := strconv.Atoi(a[2])

			for i := start; i < end+1; i++ {
				portToLookup := fmt.Sprintf("%s/%d", protocol, i)
				_, found := nmapPortsSet[portToLookup]
				if found {
					// delete port from nmapPortsSet. At the end of the range, if there are
					// ports left in nmapPortsSet then the two ports don't match
					delete(nmapPortsSet, portToLookup)
				} else {
					return false
				}
			}
		} else {
			_, found := nmapPortsSet[port]
			if found {
				// delete port from nmapPortsSet. At the end of the range, if there are
				// ports left in nmapPortsSet then the two ports don't match
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

// PortTarget holds information about a host to be scanned
type PortTarget struct {
	Host          string
	Group         *PortTargetGroup
	ScanInfo      *nmap.Run
	ScanInfoError error
}

// Scan performs a port scan on a host
func (t *PortTarget) Scan() {
	timeout, err := time.ParseDuration(t.Group.Timeout)
	if err != nil {
		t.ScanInfoError = err
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	scanner, err := nmap.NewScanner(
		nmap.WithContext(ctx),
		nmap.WithTargets(t.Host),
		nmap.WithSkipHostDiscovery(),
		// SYNScan and UDPScan require root privileges
		nmap.WithSYNScan(),
		nmap.WithUDPScan(),
	)
	if err != nil {
		t.ScanInfoError = err
		return
	}

	result, warnings, err := scanner.Run()
	if warnings != nil {
		fmt.Printf("[%s] Warnings: \n %v\n", t.Host, warnings)
	}

	if err != nil {
		t.ScanInfoError = err
		return
	}

	t.ScanInfo = result
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

	for _, host := range t.ScanInfo.Hosts {
		for _, port := range host.Ports {
			protocolPort := fmt.Sprintf("%s/%d", port.Protocol, port.ID)
			switch port.State.State {
			case string(nmap.Open):
				openPorts = append(openPorts, protocolPort)
			case string(nmap.Closed):
				closedPorts = append(closedPorts, protocolPort)
			}
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
	Timeout   string     `mapstructure:"timeout"`
	AllowList []string   `mapstructure:"allowlist"`
	BlockList []string   `mapstructure:"blocklist"`
	Discovery *Discovery `mapstructure:"discovery"`
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

package audit

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Ullaakut/nmap/v2"
	"github.com/mitchellh/mapstructure"
)

const portAuditName string = "PORT"

// portsEqual tells whether a and b contain the same elements
func portsEqual(a, b []uint16) bool {
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })

	if len(a) != len(b) {
		return false
	}

	for idx, value := range a {
		if value != b[idx] {
			return false
		}
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
	var openPorts []uint16
	var closedPorts []uint16

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
			switch port.State.State {
			case string(nmap.Open):
				openPorts = append(openPorts, port.ID)
			case string(nmap.Closed):
				closedPorts = append(closedPorts, port.ID)
			}
		}
	}

	if len(t.Group.AllowList) > 0 {
		if portsEqual(t.Group.AllowList, openPorts) {
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
		if portsEqual(t.Group.BlockList, closedPorts) {
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
	AllowList []uint16   `mapstructure:"allowlist"`
	BlockList []uint16   `mapstructure:"blocklist"`
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

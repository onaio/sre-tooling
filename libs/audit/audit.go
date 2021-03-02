package audit

import (
	"io/ioutil"
	"strings"
	"sync"

	"github.com/onaio/sre-tooling/libs/infra"
	"github.com/onaio/sre-tooling/libs/types"
	"gopkg.in/yaml.v2"
)

type Status int

const (
	Pass Status = iota
	Fail
	Error
)

type Audit interface {
	Load(input interface{}) error
	Scan() ([]*AuditResult, error)
}

type AuditResult struct {
	Type          string // type of audit e.g. "SSL", "SSH"
	Status        Status // status of the audit e.g "PASS", "ERROR", "FAIL"
	StatusMessage string // message of the audit
}

type AuditScanHandler func(results []*AuditResult, err error)

func (s Status) String() string {
	return [...]string{"PASS", "FAIL", "ERROR"}[s]
}

// CompareGrades checks whether grade1 is less than "<" grade2. Returns
// true if grade1 < grade2 else return false.
//
// Example:
//
// 		CompareGrades("B", "A+") => true since Grade B < Grade A+
// 		CompareGrades("A", "C") => false since Grade A > Grade C
// 		CompareGrades("A", "A") => false since Grade B == Grade B
// 		CompareGrades("B+", "A-") => true since Grade B+ < Grade A-
func CompareGrades(grade1, grade2 string) bool {
	grade1Val := gradeValue(grade1)
	grade2Val := gradeValue(grade2)

	return grade1Val > grade2Val
}

func gradeValue(grade string) float64 {
	value := 0.0

	for _, r := range []rune(strings.ToLower(grade)) {
		if string(r) == "+" {
			// plus grades rank higher, subtract
			value -= 0.3
		} else if string(r) == "-" {
			// minus grades rank lower, add
			value += 0.3
		} else {
			// rune contains a letter, no "-" or "+"
			value += float64(r)
		}
	}

	return value
}

type Discovery struct {
	Type          string            `mapstructure:"type"`
	Targets       []string          `mapstructure:"targets"`
	ResourceTypes []string          `mapstructure:"resource_types"`
	Regions       []string          `mapstructure:"regions"`
	Tags          map[string]string `mapstructure:"tags"`
}

func (d *Discovery) GetHosts() ([]string, error) {
	if d.Type == "host" {
		return d.Targets, nil
	}

	// discover hosts that have not been specified directly by their IP address of domain
	targets, err := d.hostDiscovery()
	return targets, err
}

func (d *Discovery) hostDiscovery() ([]string, error) {
	var targets []string

	infraFilter := &types.InfraFilter{
		Providers:     []string{d.Type},
		ResourceTypes: d.ResourceTypes,
		Regions:       d.Regions,
		Tags:          d.Tags,
	}

	resources, err := infra.GetResources(infraFilter)
	if err != nil {
		return nil, err
	}

	for _, resource := range resources {
		if resource.Properties["state"] != "running" {
			continue
		}

		targets = append(targets, resource.Properties["public-ip"])
	}

	return targets, nil
}

func EnabledAudits() map[string]Audit {
	m := make(map[string]Audit)
	m["ssl"] = &SSLAudit{}
	m["ssh"] = &SSHAudit{}
	m["port"] = &PortScan{}

	return m
}

func Run(inputFile string) ([]*AuditResult, error) {
	var auditResults []*AuditResult
	var auditWG sync.WaitGroup
	var finalErr error
	var mutex sync.Mutex

	auditFile, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, err
	}

	auditMap := make(map[string]interface{})
	err = yaml.Unmarshal(auditFile, &auditMap)
	if err != nil {
		return nil, err
	}

	enabledAudits := EnabledAudits()
	for name, audit := range enabledAudits {
		if _, prs := auditMap[name]; prs {
			err = audit.Load(auditMap[name])
			if err != nil {
				return nil, err
			}

			handler := func(results []*AuditResult, err error) {
				mutex.Lock()
				defer mutex.Unlock()

				auditResults = append(auditResults, results...)

				if err != nil {
					finalErr = err
				}
			}

			auditWG.Add(1)

			go func(audit Audit, handler AuditScanHandler) {
				defer auditWG.Done()

				res, err := audit.Scan()
				handler(res, err)

			}(audit, handler)
		}
	}

	auditWG.Wait()

	return auditResults, finalErr
}

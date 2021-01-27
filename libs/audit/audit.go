package audit

import (
	"io/ioutil"

	"github.com/mitchellh/mapstructure"

	"gopkg.in/yaml.v2"
)

type Status int

const (
	Pass Status = iota
	Fail
	Error
)

type Audit interface {
	Scan() error
	Results() []*AuditResult
}

type AuditResult struct {
	Status        Status
	StatusMessage string
}

func (s Status) String() string {
	return [...]string{"PASS", "FAIL", "ERROR"}[s]
}

func EnabledAudits() map[string]Audit {
	m := make(map[string]Audit)

	m["ssl_audit"] = &SSLAudit{}
	m["ssh_audit"] = &SSHAudit{}

	return m
}

func Run(inputFile string) error {
	auditFile, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}

	auditMap := make(map[string]interface{})
	err = yaml.Unmarshal(auditFile, &auditMap)
	if err != nil {
		return err
	}

	enabledAudits := EnabledAudits()
	for name, audit := range enabledAudits {
		if _, prs := auditMap[name]; prs {
			err = mapstructure.Decode(auditMap[name], audit)
			if err != nil {
				return err
			}

			audit.Scan()
		}
	}

	return nil
}

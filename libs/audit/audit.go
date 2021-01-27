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
	Scan() ([]*AuditResult, error)
}

type AuditResult struct {
	Status        Status
	StatusMessage string
}

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

	for _, r := range []rune(grade) {
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

func EnabledAudits() map[string]Audit {
	m := make(map[string]Audit)

	m["ssl_audit"] = &SSLAudit{}
	m["ssh_audit"] = &SSHAudit{}

	return m
}

func Run(inputFile string) ([]*AuditResult, error) {
	auditFile, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, err
	}

	auditMap := make(map[string]interface{})
	err = yaml.Unmarshal(auditFile, &auditMap)
	if err != nil {
		return nil, err
	}

	var auditResults []*AuditResult

	enabledAudits := EnabledAudits()
	for name, audit := range enabledAudits {
		if _, prs := auditMap[name]; prs {
			err = mapstructure.Decode(auditMap[name], audit)
			if err != nil {
				return nil, err
			}

			res, err := audit.Scan()
			if err != nil {
				return nil, err
			}

			auditResults = append(auditResults, res...)
		}
	}

	return auditResults, nil
}

package audit

import (
	"fmt"
	"sync"

	"github.com/mitchellh/mapstructure"
	sshaudit "github.com/onaio/sre-tooling/libs/ssh_audit"
)

const sshAuditName string = "SSH"

type AuditType int

const (
	Standard AuditType = iota
	Policy
)

func (a AuditType) String() string {
	return [...]string{"standard", "policy"}[a]
}

type Target struct {
	Host             string
	Group            *TargetGroup
	StandardScanInfo *sshaudit.StandardAuditResp
	PolicyScanInfo   *sshaudit.PolicyAuditResp
	ScanInfoError    error
}

func (target *Target) Scan(api *sshaudit.API) {
	if target.Group.AuditType == "policy" {
		scanInfo, err := api.PolicyAudit(target.Host, target.Group.Port, target.Group.PolicyName)
		target.PolicyScanInfo = scanInfo
		target.ScanInfoError = err
	} else {
		// default to "standard" audit
		scanInfo, err := api.StandardAudit(target.Host, target.Group.Port)
		target.StandardScanInfo = scanInfo
		target.ScanInfoError = err
	}

	return
}

func (target *Target) Result() []*AuditResult {
	var results []*AuditResult

	if target.ScanInfoError != nil {
		res := &AuditResult{
			Type:          sshAuditName,
			Status:        Error,
			StatusMessage: target.ScanInfoError.Error(),
		}
		results = append(results, res)
	} else {
		if target.Group.AuditType == "standard" {
			// Standard audit assigns a grade to a server after a scan
			info := target.StandardScanInfo
			if CompareGrades(info.Grade, target.Group.Threshold) {
				statusMsg := fmt.Sprintf(
					"%s (%s) with Grade %s is below threshold Grade %s",
					info.TargetServer, info.TargetServerIP, info.Grade, target.Group.Threshold,
				)
				res := &AuditResult{
					Type:          sshAuditName,
					Status:        Fail,
					StatusMessage: statusMsg,
				}
				results = append(results, res)
			} else {
				statusMsg := fmt.Sprintf(
					"%s (%s) has Grade %s",
					info.TargetServer, info.TargetServerIP, info.Grade,
				)
				res := &AuditResult{
					Type:          sshAuditName,
					Status:        Pass,
					StatusMessage: statusMsg,
				}
				results = append(results, res)
			}
		} else {
			// Policy audit only says if a server passed a test against a policy
			info := target.PolicyScanInfo
			if info.Passed {
				statusMsg := fmt.Sprintf(
					"%s (%s) policy passed", info.TargetServer, info.TargetServerIP,
				)
				res := &AuditResult{
					Type:          sshAuditName,
					Status:        Pass,
					StatusMessage: statusMsg,
				}
				results = append(results, res)
			} else {
				statusMsg := fmt.Sprintf(
					"%s (%s) policy failed", info.TargetServer, info.TargetServerIP,
				)
				res := &AuditResult{
					Type:          sshAuditName,
					Status:        Fail,
					StatusMessage: statusMsg,
				}
				results = append(results, res)
			}
		}
	}

	return results
}

type TargetGroup struct {
	Port       int        `mapstructure:"port"`
	AuditType  string     `mapstructure:"audit_type"`
	PolicyName string     `mapstructure:"policy_name"`
	Threshold  string     `mapstructure:"threshold"`
	Discovery  *Discovery `mapstructure:"discovery"`
}

func (tg *TargetGroup) Scan(api *sshaudit.API) ([]*AuditResult, error) {
	var tgAuditResults []*AuditResult
	var tgWG sync.WaitGroup
	var mutex sync.Mutex

	hosts, err := tg.Discovery.GetHosts()
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		target := &Target{
			Host:  host,
			Group: tg,
		}

		handler := func(results []*AuditResult, err error) {
			mutex.Lock()
			defer mutex.Unlock()

			tgAuditResults = append(tgAuditResults, results...)
		}

		tgWG.Add(1)

		go func(target *Target, handler AuditScanHandler) {
			defer tgWG.Done()

			target.Scan(api)
			results := target.Result()
			handler(results, nil)
		}(target, handler)
	}

	tgWG.Wait()

	return tgAuditResults, nil
}

type SSHAudit struct {
	TargetGroups []*TargetGroup
}

func (ssh *SSHAudit) Load(input interface{}) error {
	err := mapstructure.Decode(input, &ssh.TargetGroups)
	return err
}

func (ssh *SSHAudit) Scan() ([]*AuditResult, error) {
	var sshAuditResults []*AuditResult
	var sshWG sync.WaitGroup
	var mutex sync.Mutex
	var finalErr error

	api, err := sshaudit.NewAPI()
	if err != nil {
		return nil, err
	}

	for _, targetGroup := range ssh.TargetGroups {
		handler := func(results []*AuditResult, err error) {
			mutex.Lock()
			defer mutex.Unlock()

			sshAuditResults = append(sshAuditResults, results...)
			finalErr = err
		}

		sshWG.Add(1)

		go func(tg *TargetGroup, handler AuditScanHandler) {
			defer sshWG.Done()

			results, err := tg.Scan(api)
			handler(results, err)
		}(targetGroup, handler)
	}

	sshWG.Wait()

	return sshAuditResults, finalErr
}

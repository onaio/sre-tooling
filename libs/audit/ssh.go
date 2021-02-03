package audit

import (
	"fmt"
	"sync"

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

type Server struct {
	TargetServer     string                      `mapstructure:"target_server"`
	Port             int                         `mapstructure:"port"`
	AuditType        string                      `mapstructure:"audit_type"`
	PolicyName       string                      `mapstructure:"policy_name"`
	Threshold        string                      `mapstructure:"threshold"`
	StandardScanInfo *sshaudit.StandardAuditResp // scan information for standard audit
	PolicyScanInfo   *sshaudit.PolicyAuditResp   // scan information for policy audit
	ScanInfoError    error                       // contains an error if an error occured while scanning host
}

func (server *Server) Scan(api *sshaudit.API) {
	if server.AuditType == "policy" {
		scanInfo, err := api.PolicyAudit(server.TargetServer, server.Port, server.PolicyName)
		server.PolicyScanInfo = scanInfo
		server.ScanInfoError = err
	} else {
		// default to "standard" audit
		scanInfo, err := api.StandardAudit(server.TargetServer, server.Port)
		server.StandardScanInfo = scanInfo
		server.ScanInfoError = err
	}

	return
}

func (server *Server) Result() []*AuditResult {
	var results []*AuditResult

	if server.ScanInfoError != nil {
		res := &AuditResult{
			Type:          sshAuditName,
			Status:        Error,
			StatusMessage: server.ScanInfoError.Error(),
		}
		results = append(results, res)
	} else {
		if server.AuditType == "standard" {
			// Standard audit assigns a grade to a server after a scan
			info := server.StandardScanInfo
			if CompareGrades(info.Grade, server.Threshold) {
				statusMsg := fmt.Sprintf(
					"%s (%s) with Grade %s is below threshold Grade %s",
					info.TargetServer, info.TargetServerIP, info.Grade, server.Threshold,
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
			info := server.PolicyScanInfo
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

type SSHAudit struct {
	Servers []*Server `mapstructure:"servers"`
}

func (ssh *SSHAudit) Scan() ([]*AuditResult, error) {
	var sshAuditResults []*AuditResult
	var sshWG sync.WaitGroup
	var mutex sync.Mutex

	api, err := sshaudit.NewAPI()
	if err != nil {
		return nil, err
	}

	for _, server := range ssh.Servers {
		handler := func(results []*AuditResult, err error) {
			mutex.Lock()
			defer mutex.Unlock()

			sshAuditResults = append(sshAuditResults, results...)
		}

		sshWG.Add(1)

		go func(server *Server, handler AuditScanHandler) {
			defer sshWG.Done()

			server.Scan(api)
			results := server.Result()
			handler(results, nil)
		}(server, handler)
	}

	sshWG.Wait()

	return sshAuditResults, nil
}

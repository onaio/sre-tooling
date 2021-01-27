package audit

import (
	sshaudit "github.com/onaio/sre-tooling/libs/ssh_audit"
)

type Server struct {
	TargetServer string `mapstructure:"target_server"`
	Port         int    `mapstructure:"port"`
	AuditType    string `mapstructure:"audit_type"`
	PolicyName   string `mapstructure:"policy_name"`
	Threshold    string `mapstructure:"threshold"`
}

type SSHAudit struct {
	Servers []*Server `mapstructure:"servers"`
}

func (ssh *SSHAudit) Scan() error {
	api, err := sshaudit.NewAPI()
	if err != nil {
		return err
	}

	for _, server := range ssh.Servers {
		scanServer(api, server)
	}

	return nil
}

func (ssh *SSHAudit) Results() []*AuditResult {
	return nil
}

func scanServer(api *sshaudit.API, server *Server) error {
	if server.AuditType == "policy" {
		api.PolicyAudit(server.TargetServer, server.Port, server.PolicyName)
	} else {
		// default to "standard" audit
		api.StandardAudit(server.TargetServer, server.Port)
	}

	return nil
}

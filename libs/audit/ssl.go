package audit

import (
	"fmt"
	"time"

	"pkg.re/essentialkaos/sslscan.v13"
)

const sreToolingName string = "onaio/sre-tooling"
const sreToolingVersion string = ""

// DOCS: https://github.com/ssllabs/ssllabs-scan/blob/master/ssllabs-api-docs-v3.md

type Host struct {
	Host           string `mapstructure:"host"`
	Public         bool   `mapstructure:"public"`
	StartNew       bool   `mapstructure:"start_new"`
	FromCache      bool   `mapstructure:"from_cache"`
	MaxAge         int    `mapstructure:"max_age"`
	IgnoreMismatch bool   `mapstructure:"ignore_mismatch"`
}

type SSLAudit struct {
	Hosts []*Host `mapstructure:"hosts"`
}

func (ssl *SSLAudit) Scan() error {
	api, err := sslscan.NewAPI(sreToolingName, sreToolingVersion)
	if err != nil {
		return err
	}

	api.RequestTimeout = 5 * time.Second

	for _, host := range ssl.Hosts {
		_, err := scanHost(api, host)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ssl *SSLAudit) Results() []*AuditResult {
	return nil
}

func scanHost(api *sslscan.API, host *Host) (*sslscan.AnalyzeInfo, error) {
	params := sslscan.AnalyzeParams{
		Public:         host.Public,
		StartNew:       host.StartNew,
		FromCache:      host.FromCache,
		MaxAge:         host.MaxAge,
		IgnoreMismatch: host.IgnoreMismatch,
	}

	progress, err := api.Analyze(host.Host, params)
	if err != nil {
		return nil, err
	}

	var info *sslscan.AnalyzeInfo
	lastSuccess := time.Now()

	for range time.NewTicker(5 * time.Second).C {
		info, err = progress.Info(false, false)

		if info != nil && err == nil {
			// Update lastSuccess if we get result from API and there is no error
			lastSuccess = time.Now()
		}

		if info.Status == sslscan.STATUS_ERROR {
			return info, fmt.Errorf(info.StatusMessage)
		}

		if info.Status == sslscan.STATUS_READY {
			break
		}

		if time.Since(lastSuccess) > 30*time.Second {
			return info, fmt.Errorf("Can't get result from API more than 30 sec")
		}
	}

	return info, nil
}

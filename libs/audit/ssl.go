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
	Host           string               `mapstructure:"host"`
	Public         bool                 `mapstructure:"public"`
	StartNew       bool                 `mapstructure:"start_new"`
	FromCache      bool                 `mapstructure:"from_cache"`
	MaxAge         int                  `mapstructure:"max_age"`
	IgnoreMismatch bool                 `mapstructure:"ignore_mismatch"`
	Threshold      string               `mapstructure:"threshold"`
	ScanInfo       *sslscan.AnalyzeInfo // scan information
	ScanInfoError  error                // contains an error if an error occured while scanning host
}

func (host *Host) Scan(api *sslscan.API) {
	params := sslscan.AnalyzeParams{
		Public:         host.Public,
		StartNew:       host.StartNew,
		FromCache:      host.FromCache,
		MaxAge:         host.MaxAge,
		IgnoreMismatch: host.IgnoreMismatch,
	}

	progress, err := api.Analyze(host.Host, params)
	if err != nil {
		host.ScanInfoError = err
		return
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
			host.ScanInfoError = fmt.Errorf(info.StatusMessage)
			break
		}

		if info.Status == sslscan.STATUS_READY {
			break
		}

		if time.Since(lastSuccess) > 30*time.Second {
			host.ScanInfoError = fmt.Errorf("Can't get result from API more than 30 sec")
			break
		}
	}

	host.ScanInfo = info
	return
}

func (host *Host) Result() []*AuditResult {
	var results []*AuditResult

	if host.ScanInfoError != nil {
		res := &AuditResult{
			Status:        Error,
			StatusMessage: host.ScanInfoError.Error(),
		}
		results = append(results, res)
	} else {
		for _, endpoint := range host.ScanInfo.Endpoints {
			if CompareGrades(endpoint.Grade, host.Threshold) {
				statusMsg := fmt.Sprintf(
					"%s (%s) with Grade %s is below threshold Grade %s",
					host.ScanInfo.Host, endpoint.IPAdress, endpoint.Grade, host.Threshold,
				)
				res := &AuditResult{
					Status:        Fail,
					StatusMessage: statusMsg,
				}
				results = append(results, res)
			} else {
				statusMsg := fmt.Sprintf(
					"%s (%s) has Grade %s",
					host.ScanInfo.Host, endpoint.IPAdress, endpoint.Grade,
				)
				res := &AuditResult{
					Status:        Pass,
					StatusMessage: statusMsg,
				}
				results = append(results, res)
			}
		}
	}

	return results
}

type SSLAudit struct {
	Hosts []*Host `mapstructure:"hosts"`
}

func (ssl *SSLAudit) Scan() ([]*AuditResult, error) {
	api, err := sslscan.NewAPI(sreToolingName, sreToolingVersion)
	if err != nil {
		return nil, err
	}

	api.RequestTimeout = 5 * time.Second

	var sslAuditResults []*AuditResult

	for _, host := range ssl.Hosts {
		host.Scan(api)

		results := host.Result()

		sslAuditResults = append(sslAuditResults, results...)
	}

	return sslAuditResults, nil
}
package ssl

import (
	"fmt"
	"time"

	"pkg.re/essentialkaos/sslscan.v13"
)

const sreToolingName string = "sre-tooling"
const sreToolingVersion string = ""

// DOCS: https://github.com/ssllabs/ssllabs-scan/blob/master/ssllabs-api-docs-v3.md

type Host struct {
	Host           string               `yaml:"host"`
	Public         bool                 `yaml:"public"`
	StartNew       bool                 `yaml:"start_new"`
	FromCache      bool                 `yaml:"from_cache"`
	MaxAge         int                  `yaml:"max_age"`
	IgnoreMismatch bool                 `yaml:"ignore_mismatch"`
	ScanInfo       *sslscan.AnalyzeInfo // scan information
	ScanInfoError  string               // contains an error message if an error occured while scanning host
}

type SSLHosts struct {
	Hosts []*Host `yaml:"hosts"`
}

func (hosts *SSLHosts) Scan() error {
	api, err := sslscan.NewAPI(sreToolingName, sreToolingVersion)
	api.RequestTimeout = 5 * time.Second

	if err != nil {
		return err
	}

	for _, host := range hosts.Hosts {
		params := sslscan.AnalyzeParams{
			Public:         host.Public,
			StartNew:       host.StartNew,
			FromCache:      host.FromCache,
			MaxAge:         host.MaxAge,
			IgnoreMismatch: host.IgnoreMismatch,
		}
		progress, err := api.Analyze(host.Host, params)

		if err != nil {
			host.ScanInfoError = err.Error()
			break
		}

		var info *sslscan.AnalyzeInfo
		fmt.Printf("Progress (%s): ∙", host.Host)
		lastSuccess := time.Now()

		for range time.NewTicker(5 * time.Second).C {
			info, err = progress.Info(false, false)

			if info != nil && err == nil {
				// Update lastSuccess if we get result from API and there is no error
				lastSuccess = time.Now()
			}

			if info.Status == sslscan.STATUS_ERROR {
				host.ScanInfoError = info.StatusMessage
				break
			}

			if info.Status == sslscan.STATUS_READY {
				break
			}

			if time.Since(lastSuccess) > 30*time.Second {
				host.ScanInfoError = "Can't get result from API more than 30 sec"
				break
			}

			fmt.Printf("∙")
		}

		fmt.Printf("\n")

		host.ScanInfo = info
	}

	return nil
}

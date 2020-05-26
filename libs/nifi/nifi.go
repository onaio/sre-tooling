package nifi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const nifiSystemDiagnosticsURLEnvVar string = "SRE_NIFI_SYSTEM_DIAGNOSTICS_URL"

// SystemDiagnosticsResp is the root JSON object from the NiFi diagnostics endpoint
// response
type SystemDiagnosticsResp struct {
	SystemDiagnostics SystemDiagnostics `json:"systemDiagnostics"`
}

// SystemDiagnostics is part of the response from the NiFi diagnostics endpoint response
// and is a child element of SystemDiagnosticsResp
type SystemDiagnostics struct {
	AggregateSnapshot AggregateSnapshot `json:"aggregateSnapshot"`
}

// AggregateSnapshot is part of the response from the NiFi diagnostics endpoint response
// and contains the version information for NiFi. The JSON response object corresponding
// to this object contains other elements apart from VersionInfo
type AggregateSnapshot struct {
	VersionInfo VersionInfo `json:"versionInfo`
}

// VersionInfo is part of the response from the NiFi diagnostics endpoint response
// and contains version informantion for the NiFi instance
type VersionInfo struct {
	NiFiVersion    string `json:"niFiVersion"`
	JavaVendor     string `json:"javaVendor"`
	JavaVersion    string `json:"javaVersion"`
	OSName         string `json:"osName"`
	OSVersion      string `json:"osVersion"`
	OSArchitecture string `json:"osArchitecture"`
	BuildTag       string `json:"buildTag"`
	BuildRevision  string `json:"buildRevision"`
	BuildBranch    string `json:"buildBranch"`
	BuildTimestamp string `json:"buildTimestamp"`
}

// GetSystemDiagnostics returns system diagnostics information from NiFi.
// It's expected that the environment value whose name is in
// the nifiSystemDiagnosticsURLEnvVar variable
func GetSystemDiagnostics() (SystemDiagnosticsResp, error) {
	var systemDiagnosticsResp SystemDiagnosticsResp

	client := &http.Client{}
	diagnosticsURL := os.Getenv(nifiSystemDiagnosticsURLEnvVar)
	req, requestErr := http.NewRequest("GET", diagnosticsURL, nil)
	req.Header.Add("Accept", "application/json")

	if requestErr != nil {
		return systemDiagnosticsResp, requestErr
	}

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()
	resp, respErr := client.Do(req)
	if respErr != nil {
		return systemDiagnosticsResp, respErr
	}

	defer resp.Body.Close()
	respBody, respBodyErr := ioutil.ReadAll(resp.Body)

	if respBodyErr != nil {
		return systemDiagnosticsResp, respBodyErr
	}

	if resp.StatusCode != 200 {
		return systemDiagnosticsResp, fmt.Errorf("Status from NiFi diagnostics endpoint is %d. Expecting 200", resp.StatusCode)
	}

	marshallErr := json.Unmarshal(respBody, &systemDiagnosticsResp)

	return systemDiagnosticsResp, marshallErr
}

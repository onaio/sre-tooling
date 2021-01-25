package sshaudit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
)

const baseURL = "https://www.sshaudit.com"

// StandardAuditResp holds response of a standard SSH audit
//
// A standard audit evaluates each of the individual cryptographic algorithms
// supported by the target. An overall score is given based on how many strong,
// acceptable, and weak options are available.
type StandardAuditResp struct {
	AuditType        string `json:"audit_type"`
	Banner           string `json:"banner"`
	Score            int    `json:"score"`
	Grade            string `json:"grade"`
	Version          string `json:"version"`
	TargetServer     string `json:"target_server"`
	TargetServerPort int    `json:"target_server_port"`
	TargetServerIP   string `json:"target_server_ip"`
}

// PolicyAuditResp holds response of a SSH audit against a policy
//
// A policy audit determines if the target adheres to a specific set of
// expected options. The resulting score is either pass or fail. Policy
// audits are useful for ensuring a server has been successfully
// (and remains) hardened.
type PolicyAuditResp struct {
	AuditType      string `json:"audit_type"`
	TargetServer   string `json:"target_server"`
	TargetServerIP string `json:"target_server_ip"`
	PolicyName     string `json:"policy_name"`
	Passed         bool   `json:"passed"`
}

// API performs requests to www.sshaudit.com
type API struct {
	Cookies   []*http.Cookie
	CSRFToken string
	Client    *http.Client
}

func (api *API) makeRequest(req *http.Request, result interface{}) (*http.Response, error) {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", getUserAgent())

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBody, result)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request to URL %s returned HTTP code %d", req.URL, resp.StatusCode)
	}

	return resp, nil
}

func (api *API) ping() error {
	var pingResult map[string]interface{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ping", baseURL), nil)
	if err != nil {
		return err
	}

	resp, err := api.makeRequest(req, &pingResult)
	if err != nil {
		return err
	}

	api.CSRFToken = fmt.Sprintf("%v", pingResult["csrf_token"])
	api.Cookies = resp.Cookies()

	return nil
}

func (api *API) serverAudit(payload *url.Values, result interface{}) error {
	payload.Add("csrf_token", api.CSRFToken)

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/server_audit", baseURL),
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return err
	}

	// set cookies
	for _, cookie := range api.Cookies {
		req.AddCookie(cookie)
	}

	_, err = api.makeRequest(req, result)

	return err
}

// StandardAudit runs a standard SSH audit on a server
func (api *API) StandardAudit(server string, port int) (*StandardAuditResp, error) {
	payload := &url.Values{}
	payload.Add("s", server)
	payload.Add("p", strconv.Itoa(port))
	payload.Add("audit_type", "standard")

	standardAuditResp := &StandardAuditResp{}
	err := api.serverAudit(payload, standardAuditResp)

	if err != nil {
		return nil, err
	}

	return standardAuditResp, nil
}

// PolicyAudit runs a SSH audit on a server using a specified policy
func (api *API) PolicyAudit(server string, port int, policyName string) (*PolicyAuditResp, error) {
	payload := &url.Values{}
	payload.Add("s", server)
	payload.Add("p", strconv.Itoa(port))
	payload.Add("audit_type", "policy")
	payload.Add("policy_name", policyName)

	policyAuditResp := &PolicyAuditResp{}
	err := api.serverAudit(payload, policyAuditResp)

	if err != nil {
		return nil, err
	}

	return policyAuditResp, nil
}

// NewAPI creates an API struct
func NewAPI() (*API, error) {
	api := &API{
		Client: &http.Client{},
	}

	err := api.ping()
	if err != nil {
		return nil, err
	}

	return api, nil
}

// getUserAgent generate user-agent string for client
func getUserAgent() string {
	return fmt.Sprintf(
		"onaio/sre-tooling (go; %s; %s-%s)",
		runtime.Version(), runtime.GOARCH, runtime.GOOS,
	)
}

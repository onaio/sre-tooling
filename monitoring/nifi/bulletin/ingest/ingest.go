package ingest

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "ingest"
const nifiBulletinBoardAPIURLEnvVar string = "SRE_MONITORING_NIFI_BULLETINGBOARD_API_URL"

// Ingest does...
type Ingest struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
	next        *int
	limit       *int
}

// APIResponse does..
type APIResponse struct {
	BulletinBoard BulletinBoard `json:"bulletinBoard"`
}

// BulletinBoard does..
type BulletinBoard struct {
	Bulletins []BulletinProcessor `json:"bulletins"`
	Generated string              `json:"generated"`
}

// BulletinProcessor does...
type BulletinProcessor struct {
	GroupID  string   `json:"groupId"`
	SourceID string   `json:"sourceId"`
	CanRead  bool     `json:"canRead"`
	Bulletin Bulletin `json:"bulletin"`
}

// Bulletin does...
type Bulletin struct {
	ID         int    `json:"id"`
	Category   string `json:"category"`
	SourceName string `json:"sourceName"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
}

// Init does...
func (ingest *Ingest) Init(helpFlagName string, helpFlagDescription string) {
	ingest.flagSet = flag.NewFlagSet(ingest.GetName(), flag.ExitOnError)
	ingest.helpFlag = ingest.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	ingest.next = ingest.flagSet.Int("next", 0, "includes bulletins with an id after this value")
	ingest.limit = ingest.flagSet.Int("limit", 0, "the number of bulletins to limit the request to")
}

// GetName does...
func (ingest *Ingest) GetName() string {
	return name
}

// GetDescription does...
func (ingest *Ingest) GetDescription() string {
	return "Ingests NiFi bulletins and sends to the configured notification channel(s)"
}

// GetFlagSet does..
func (ingest *Ingest) GetFlagSet() *flag.FlagSet {
	return ingest.flagSet
}

// GetSubCommands does..
func (ingest *Ingest) GetSubCommands() []cli.Command {
	return ingest.subCommands
}

// GetHelpFlag does..
func (ingest *Ingest) GetHelpFlag() *bool {
	return ingest.helpFlag
}

// Process does...
func (ingest *Ingest) Process() {
	client := &http.Client{}
	apiURL := os.Getenv(nifiBulletinBoardAPIURLEnvVar)
	req, requestErr := http.NewRequest("GET", apiURL, nil)
	req.Header.Add("Accept", "application/json")

	if requestErr != nil {
		notification.SendMessage(requestErr.Error())
		cli.ExitCommandExecutionError()
	}

	q := req.URL.Query()
	q.Add("next", strconv.Itoa(*ingest.next))

	if *ingest.limit != 0 {
		q.Add("limit", strconv.Itoa(*ingest.limit))
	}

	req.URL.RawQuery = q.Encode()

	resp, respErr := client.Do(req)

	if respErr != nil {
		notification.SendMessage(respErr.Error())
		cli.ExitCommandExecutionError()
	}

	defer resp.Body.Close()
	respBody, respBodyErr := ioutil.ReadAll(resp.Body)

	if respBodyErr != nil {
		notification.SendMessage(respBodyErr.Error())
		cli.ExitCommandExecutionError()
	}

	if resp.StatusCode != 200 {
		notification.SendMessage(fmt.Sprintf("Status from NiFi bulletin API is %d", resp.StatusCode))
		cli.ExitCommandExecutionError()
	}

	var apiResponse APIResponse
	marshallErr := json.Unmarshal(respBody, &apiResponse)

	if marshallErr != nil {
		notification.SendMessage(marshallErr.Error())
		cli.ExitCommandExecutionError()
	}

	val, _ := json.Marshal(apiResponse)

	fmt.Printf("%s", string(val))
}

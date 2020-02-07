package ingest

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
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

type NiFiEvent interface {
	GetCategory() string
	GetId() string
	GetSourceId() string
	GetGroupId() string
	GetSourceName() string
	GetTimestamp() string
	GetRuntime() string
	GetRuntimeName() string
}

type NiFiApi interface {
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
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "dsn",
	})

	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

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

	for _, currBulletin := range apiResponse.BulletinBoard.Bulletins {
		event := sentry.NewEvent()
		event.Message = currBulletin.Bulletin.Message
		event.EventID = sentry.EventID(currBulletin.Bulletin.ID)
		event.Level = sentry.Level(strings.ToLower(currBulletin.Bulletin.Level))
		event.Tags["category"] = currBulletin.Bulletin.Category
		event.Tags["ID"] = string(currBulletin.Bulletin.ID)
		event.Tags["sourceID"] = currBulletin.SourceID
		event.Tags["groupID"] = currBulletin.GroupID
		event.Tags["sourceName"] = currBulletin.Bulletin.SourceName
		event.Tags["timestamp"] = currBulletin.Bulletin.Timestamp
		event.Tags["runtime"] = "NiFi-1.8.0"
		event.Tags["runtime.name"] = "NiFi"
		event.Fingerprint = []string{currBulletin.GroupID}
		sentry.CaptureEvent(event)
	}
	sentry.Flush(time.Second * 30)
}

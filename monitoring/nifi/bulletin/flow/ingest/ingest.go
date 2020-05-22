package ingest

import (
	"bufio"
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
const nifiBulletinBoardAPIURLEnvVar string = "SRE_MONITORING_NIFI_FLOW_BULLETIN_API_URL"
const nifiSentryDsnEnvVar string = "SRE_MONITORING_NIFI_FLOW_BULLETIN_SENTRY_DSN"

// Ingest does...
type Ingest struct {
	helpFlag            *bool
	persistencePathFlag *string
	flagSet             *flag.FlagSet
	subCommands         []cli.Command
	nextFlag            *int
	limitFlag           *int
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
	ID       int64    `json:"id"`
	GroupID  string   `json:"groupId"`
	SourceID string   `json:"sourceId"`
	CanRead  bool     `json:"canRead"`
	Bulletin Bulletin `json:"bulletin"`
}

// Bulletin does...
type Bulletin struct {
	ID         int64  `json:"id"`
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
	ingest.persistencePathFlag = ingest.flagSet.String("persistence-path", ".sre-tooling-monitoring-nifi-bulletin.last", "Where to store persistent data for this command")
	ingest.nextFlag = ingest.flagSet.Int("next", 0, "Includes bulletins with an id after this value")
	ingest.limitFlag = ingest.flagSet.Int("limit", 0, "The number of bulletins to limit the request to")
}

// GetName does...
func (ingest *Ingest) GetName() string {
	return name
}

// GetDescription does...
func (ingest *Ingest) GetDescription() string {
	return "Ingests the NiFi flow bulletins and sends the data to the configured Sentry DSN"
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
	apiURL := os.Getenv(nifiBulletinBoardAPIURLEnvVar)
	sentryDSN := os.Getenv(nifiSentryDsnEnvVar)

	err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
	})

	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	client := &http.Client{}
	req, requestErr := http.NewRequest("GET", apiURL, nil)
	req.Header.Add("Accept", "application/json")

	if requestErr != nil {
		notification.SendMessage(requestErr.Error())
		cli.ExitCommandExecutionError()
	}

	q := req.URL.Query()
	q.Add("next", strconv.Itoa(*ingest.nextFlag))

	if *ingest.limitFlag != 0 {
		q.Add("limit", strconv.Itoa(*ingest.limitFlag))
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

	lastId, lastIdErr := getLastId(ingest.persistencePathFlag)
	if lastIdErr != nil {
		notification.SendMessage(lastIdErr.Error())
		cli.ExitCommandExecutionError()
	}

	for _, currBulletin := range apiResponse.BulletinBoard.Bulletins {
		if lastId >= currBulletin.ID {
			fmt.Printf("Event in bulletin with ID '%d' already sent to Sentry. Not sending it again.\n", currBulletin.ID)
			continue
		}

		event := sentry.NewEvent()
		event.Message = currBulletin.Bulletin.SourceName + " [" + currBulletin.GroupID + "]"
		event.EventID = sentry.EventID(currBulletin.ID)
		event.Exception = buildSentryException(currBulletin.Bulletin.SourceName+" ["+currBulletin.GroupID+"]", currBulletin.Bulletin.Message)
		event.Level = sentry.Level(strings.ToLower(currBulletin.Bulletin.Level))
		event.Tags["category"] = currBulletin.Bulletin.Category
		event.Tags["ID"] = fmt.Sprintf("%d", currBulletin.ID)
		event.Tags["sourceID"] = currBulletin.SourceID
		event.Tags["groupID"] = currBulletin.GroupID
		event.Tags["sourceName"] = currBulletin.Bulletin.SourceName
		event.Tags["timestamp"] = currBulletin.Bulletin.Timestamp
		event.Tags["runtime"] = "NiFi-1.8.0" // TODO: Get version of NiFi
		event.Tags["runtime.name"] = "NiFi"
		event.Fingerprint = []string{currBulletin.GroupID, currBulletin.SourceID}
		sentry.CaptureEvent(event)
		lastId = currBulletin.ID
	}
	sentry.Flush(time.Second * 30)

	// Save lastId to file after loop is done. Benefit of checkpointing within
	// the loop (and potentially have a super accurate lastId--in case the command
	// errors) not considered more beneficial than less frequent I/O operations.
	saveErr := saveLastId(ingest.persistencePathFlag, lastId)
	if saveErr != nil {
		notification.SendMessage(saveErr.Error())
		cli.ExitCommandExecutionError()
	}
}

func buildSentryException(messageType string, message string) []sentry.Exception {
	return []sentry.Exception{sentry.Exception{
		Value:      message,
		Type:       messageType,
		Stacktrace: sentry.ExtractStacktrace(fmt.Errorf(message)),
	}}
}

func saveLastId(path *string, lastId int64) error {
	return ioutil.WriteFile(*path, []byte(fmt.Sprintf("%d", lastId)), 0600)
}

func getLastId(path *string) (int64, error) {
	// Check if file exists. If it doesn't return 0, nil
	if _, statErr := os.Stat(*path); os.IsNotExist(statErr) {
		return 0, nil
	}

	file, fileErr := os.Open(*path)
	if fileErr != nil {
		return 0, fileErr
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// We expect the last ID to be the first line in the file
	scanner.Scan()
	lastIdString := scanner.Text()
	if scannerErr := scanner.Err(); scannerErr != nil {
		return 0, scannerErr
	}

	lastId, lastIdErr := parseNiFiIdString(lastIdString)

	return lastId, lastIdErr
}

func parseNiFiIdString(idString string) (int64, error) {
	return strconv.ParseInt(idString, 10, 64)
}

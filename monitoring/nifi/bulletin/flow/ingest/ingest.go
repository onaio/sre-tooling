package ingest

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/onaio/sre-tooling/libs/nifi"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "ingest"
const nifiBulletinBoardAPIURLEnvVar string = "SRE_MONITORING_NIFI_FLOW_BULLETIN_URL"
const nifiSentryDsnEnvVar string = "SRE_MONITORING_NIFI_FLOW_BULLETIN_SENTRY_DSN"
const lastIdExtension string = ".last"

// The format for the date portion to be added to the timestamp
// The timestamp string from the bulletin board doesn't come with a date portion
// This constant dictates how the date should be formatted before prepending it to the time
const nifiMissingDateFormat string = "Mon, 02 Jan 2006"
const nifiFormattedTimestampFormat string = time.RFC1123 // Matches Mon, 02 Jan 2006 15:04:05 MST

// Ingest does...
type Ingest struct {
	helpFlag           *bool
	persistenceDirFlag *string
	flagSet            *flag.FlagSet
	subCommands        []cli.Command
	nextFlag           *int
	limitFlag          *int
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
	ingest.persistenceDirFlag = ingest.flagSet.String("persistence-dir", ".", "The directory to store persistent data for this command")
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
	release := "n/a"
	sysDiagnosticsResp, sysDiagnosticsErr := nifi.GetSystemDiagnostics()
	if sysDiagnosticsErr != nil {
		notification.SendMessage(sysDiagnosticsErr.Error())
	} else {
		release = sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.NiFiVersion
	}

	apiURL := os.Getenv(nifiBulletinBoardAPIURLEnvVar)
	sentryDSN := os.Getenv(nifiSentryDsnEnvVar)

	err := sentry.Init(sentry.ClientOptions{
		Dsn:     sentryDSN,
		Release: release,
		Dist:    "NiFi",
	})

	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	if sysDiagnosticsErr == nil {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("os_name", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.OSName)
			scope.SetTag("os_version", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.OSVersion)
			scope.SetTag("os_architecture", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.OSArchitecture)
			scope.SetTag("java_version", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.JavaVersion)
			scope.SetTag("java_vendor", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.JavaVendor)
			scope.SetTag("build_branch", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.BuildBranch)
			scope.SetTag("build_revision", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.BuildRevision)
			scope.SetTag("build_tag", sysDiagnosticsResp.SystemDiagnostics.AggregateSnapshot.VersionInfo.BuildTag)
		})
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

	lastId, lastIdErr := getLastId(ingest.persistenceDirFlag, &apiURL, &sentryDSN)
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
		event.Exception = buildSentryException(currBulletin.Bulletin.SourceName+" ["+currBulletin.GroupID+"]", currBulletin.Bulletin.Message)
		event.Level = sentry.Level(strings.ToLower(currBulletin.Bulletin.Level))

		timestampString := currBulletin.Bulletin.Timestamp
		timestamp, timestampErr := parseNiFiTimestampString(timestampString)
		if timestampErr != nil {
			notification.SendMessage(timestampErr.Error())
		} else {
			timestampString = timestamp.Format(nifiFormattedTimestampFormat)
			event.Timestamp = timestamp.Unix()
		}

		event.Tags["category"] = currBulletin.Bulletin.Category
		event.Tags["id"] = fmt.Sprintf("%d", currBulletin.ID)
		event.Tags["source_id"] = currBulletin.SourceID
		event.Tags["group_id"] = currBulletin.GroupID
		event.Tags["source_name"] = currBulletin.Bulletin.SourceName
		event.Tags["timestamp_string"] = timestampString
		event.Fingerprint = []string{currBulletin.GroupID, currBulletin.SourceID}
		sentry.CaptureEvent(event)
		lastId = currBulletin.ID
	}
	sentry.Flush(time.Second * 30)

	// Save lastId to file after loop is done. Benefit of checkpointing within
	// the loop (and potentially have a super accurate lastId--in case the command
	// errors) not considered more beneficial than less frequent I/O operations.
	saveErr := saveLastId(ingest.persistenceDirFlag, &apiURL, &sentryDSN, lastId)
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

func getLastIdStoragePath(storageDir *string, flowBulletinURL *string, sentryDSN *string) (string, error) {
	if len(*storageDir) == 0 || len(*flowBulletinURL) == 0 || len(*sentryDSN) == 0 {
		return "", fmt.Errorf("Make sure the storage directory, flow bulletin URL and Sentry DSN are not blank")
	}

	hasher := sha256.New()
	hasher.Write([]byte(*flowBulletinURL + *sentryDSN))
	sha := hex.EncodeToString(hasher.Sum(nil))

	return *storageDir + string(os.PathSeparator) + sha + lastIdExtension, nil
}

func saveLastId(storageDir *string, flowBulletinURL *string, sentryDSN *string, lastId int64) error {
	path, pathErr := getLastIdStoragePath(storageDir, flowBulletinURL, sentryDSN)
	if pathErr != nil {
		return pathErr
	}

	return ioutil.WriteFile(path, []byte(fmt.Sprintf("%d", lastId)), 0600)
}

func getLastId(storageDir *string, flowBulletinURL *string, sentryDSN *string) (int64, error) {
	path, pathErr := getLastIdStoragePath(storageDir, flowBulletinURL, sentryDSN)
	if pathErr != nil {
		return 0, pathErr
	}

	// Check if file exists. If it doesn't return 0, nil
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return 0, nil
	}

	file, fileErr := os.Open(path)
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

func parseNiFiTimestampString(timestamp string) (time.Time, error) {
	if len(timestamp) == 0 {
		return time.Now(), fmt.Errorf("The provided unformatted timestamp is blank")
	}

	dateString := time.Now().Format(nifiMissingDateFormat)
	timestamp = dateString + " " + timestamp

	return time.Parse(nifiFormattedTimestampFormat, timestamp)
}

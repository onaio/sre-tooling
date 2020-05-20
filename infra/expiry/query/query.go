package query

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/onaio/sre-tooling/libs/notification"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/cloud"
)

const name string = "query"
const defaultTimeFormat string = "2006-01-02"
const defaultExpiryTagNAValue string = "-"
const dataFieldExpiryTime = "expiry-time"
const outputFormatPlain = "plain"
const outputFormatMarkdown = "markdown"

type ExpiredResourceHandler func(resource *cloud.Resource, hasExpired bool, expiryTime time.Time, err error)

// Query queries then notifies (using configured notification channels) infrastructure that has expired
type Query struct {
	helpFlag              *bool
	flagSet               *flag.FlagSet
	providerFlag          *flags.StringArray
	regionFlag            *flags.StringArray
	typeFlag              *flags.StringArray
	tagFlag               *flags.StringArray
	maxAgeFlag            *string
	expiryTagFlag         *string
	expiryTagNAValueFlag  *string
	expiryTagFormatFlag   *string
	showFlag              *flags.StringArray
	hideHeadersFlag       *bool
	csvFlag               *bool
	fieldSeparatorFlag    *string
	resourceSeparatorFlag *string
	listFieldsFlag        *bool
	defaultFieldValueFlag *string
	outputFormatFlag      *string
	subCommands           []cli.Command
}

// Init initializes the command object
func (query *Query) Init(helpFlagName string, helpFlagDescription string) {
	query.flagSet = flag.NewFlagSet(query.GetName(), flag.ExitOnError)
	query.helpFlag = query.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	query.outputFormatFlag = query.flagSet.String(
		"output-format",
		outputFormatPlain,
		fmt.Sprintf(
			"How to format the full output text. Possible values are '%s' and '%s'.",
			outputFormatPlain,
			outputFormatMarkdown))

	query.providerFlag,
		query.regionFlag,
		query.typeFlag,
		query.tagFlag,
		query.maxAgeFlag,
		query.expiryTagFlag,
		query.expiryTagNAValueFlag,
		query.expiryTagFormatFlag = AddQueryFlags(query.flagSet)

	query.showFlag,
		query.hideHeadersFlag,
		query.csvFlag,
		query.fieldSeparatorFlag,
		query.resourceSeparatorFlag,
		query.listFieldsFlag,
		query.defaultFieldValueFlag = cloud.AddResourceTableFlags(query.flagSet)

}

// AddQueryFlags returns the flags required to query expired infrastructure in the order:
// 	- Infrastructure filter tags, in the order needed by cloud.AddFilterFlags
//  - Maximum age flag
//	- Expiry tag flag
//	- Expiry tag not applicable value
//	- Expiry tag format flag
func AddQueryFlags(flagSet *flag.FlagSet) (*flags.StringArray, *flags.StringArray, *flags.StringArray, *flags.StringArray, *string, *string, *string, *string) {
	providerFlag,
		regionFlag,
		typeFlag,
		tagFlag := cloud.AddFilterFlags(flagSet)

	maxAgeFlag := flagSet.String("max-age", "", "Maximum age of a resource e.g '1h' to mean one hour. Valid time units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', and 'h'.")
	expiryTagFlag := flagSet.String("expiry-tag", "", "Name of the tag storing the time when the resource will expire")
	expiryTagNAValue := flagSet.String("expiry-tag-na-value", defaultExpiryTagNAValue, "Value for the expiry tag that symbolizes that resource doesn't expire")
	expiryTagFormatFlag := flagSet.String("expiry-tag-format", defaultTimeFormat, "The format of the time in the tag specified in -expiry-tag. Check the Golang time documentation on example formats here -> https://golang.org/pkg/time/#pkg-constants.")

	return providerFlag,
		regionFlag,
		typeFlag,
		tagFlag,
		maxAgeFlag,
		expiryTagFlag,
		expiryTagNAValue,
		expiryTagFormatFlag
}

// GetName returns the value of the name constant
func (query *Query) GetName() string {
	return name
}

// GetDescription returns the description for the query command
func (query *Query) GetDescription() string {
	return "Notifies, using configured notification channels, infrastructure that has expired"
}

// GetFlagSet returns a pointer to the flag.FlagSet associated to the command
func (query *Query) GetFlagSet() *flag.FlagSet {
	return query.flagSet
}

// GetSubCommands returns a slice of subcommands under the query command
// (expect empty slice if none)
func (query *Query) GetSubCommands() []cli.Command {
	return query.subCommands
}

// GetHelpFlag returns a pointer to the initialized help flag for the command
func (query *Query) GetHelpFlag() *bool {
	return query.helpFlag
}

// Process fetches the list of infrastructure that matches the criteria provided by the user
// and that has expired and sends notifications to the configured notification channels
func (query *Query) Process() {
	hasResourceErr := false
	var expiredResources []*cloud.Resource
	resourceErr := GetExpiredResources(
		query.providerFlag,
		query.regionFlag,
		query.typeFlag,
		query.tagFlag,
		query.maxAgeFlag,
		query.expiryTagFlag,
		query.expiryTagNAValueFlag,
		query.expiryTagFormatFlag,
		func(resource *cloud.Resource, hasExpired bool, expiryTime time.Time, err error) {
			if err != nil {
				notification.SendMessage(fmt.Errorf("Could not figure out which resources have expired: %w", err).Error())
				hasResourceErr = true
				return
			}

			if !hasExpired {
				return
			}

			if resource.Data == nil {
				resource.Data = make(map[string]string)
			}

			resource.Data[dataFieldExpiryTime] = expiryTime.Format(time.RFC1123)
			expiredResources = append(expiredResources, resource)
		},
	)

	if resourceErr != nil {
		notification.SendMessage(resourceErr.Error())
		hasResourceErr = true
	}

	if hasResourceErr {
		cli.ExitCommandExecutionError()
	}

	if len(expiredResources) == 0 {
		return
	}

	rt := new(cloud.ResourceTable)
	rt.Init(
		query.showFlag,
		query.hideHeadersFlag,
		query.csvFlag,
		query.fieldSeparatorFlag,
		query.resourceSeparatorFlag,
		query.listFieldsFlag,
		query.defaultFieldValueFlag)
	table, tableErr := rt.Render(expiredResources)
	if tableErr != nil {
		notification.SendMessage(tableErr.Error())
	}

	formattedOutput := ""
	errorMessage := "Some cloud resources have expired:"
	switch *query.outputFormatFlag {
	case outputFormatMarkdown:
		formattedOutput = fmt.Sprintf("%s\n```\n%s```", errorMessage, table)
	case outputFormatPlain:
		formattedOutput = fmt.Sprintf("%s\n%s", errorMessage, table)
	default:
		notification.SendMessage(fmt.Sprintf("Unrecognized output format '%s'", *query.outputFormatFlag))
		cli.ExitCommandInterpretationError()
	}

	notification.SendMessage(formattedOutput)
	cli.ExitCommandExecutionError()
}

// GetExpiredResources returns a list of expired resources
func GetExpiredResources(
	providerFlag *flags.StringArray,
	regionFlag *flags.StringArray,
	typeFlag *flags.StringArray,
	tagFlag *flags.StringArray,
	maxAgeFlag *string,
	expiryTagFlag *string,
	expiryTagNAValueFlag *string,
	expiryTagFormatFlag *string,
	expiredResourceHandler ExpiredResourceHandler) error {

	if len(*expiryTagFlag) == 0 && len(*maxAgeFlag) == 0 {
		return fmt.Errorf("Either maximum age or expiry tag need to be provided")
	}

	if len(*expiryTagFlag) > 0 && len(*expiryTagFormatFlag) == 0 {
		return fmt.Errorf("If the expiry tag is provided, then the expiry tag format also needs to be provided")
	}

	allResources, resourcesErr := cloud.GetAllCloudResources(
		cloud.GetFiltersFromCommandFlags(
			providerFlag,
			regionFlag,
			typeFlag,
			tagFlag),
		true)

	if resourcesErr != nil {
		return fmt.Errorf("Could not get the list of cloud resources: %w", resourcesErr)
	}

	for _, curResource := range allResources {
		hasExpired, expiryTime, hasExpiredErr := hasResourceExpired(curResource, maxAgeFlag, expiryTagFlag, expiryTagNAValueFlag, expiryTagFormatFlag)
		expiredResourceHandler(curResource, hasExpired, expiryTime, hasExpiredErr)
	}

	return nil
}

// hasResourceExpired checks whether a resource has expired using the provided maximum age and expiry time flag
func hasResourceExpired(resource *cloud.Resource, maxAge *string, expiryTag *string, expiryTagNAValue, expiryTagFormat *string) (bool, time.Time, error) {
	maxAgeReached, maxAgeExpiryTime, maxAgeReachedErr := hasMaxAgeReached(resource, maxAge)
	if maxAgeReached {
		return true, maxAgeExpiryTime, maxAgeReachedErr
	}

	expiryTimeMature, matureExpiryTime, expiryTimeMatureErr := hasExpiryTimeMatured(resource, expiryTag, expiryTagNAValue, expiryTagFormat)
	if expiryTimeMature {
		return true, matureExpiryTime, expiryTimeMatureErr
	}

	// Only return the errors if all methods of evaluating expiry have failed
	if maxAgeReachedErr != nil {
		return false, maxAgeExpiryTime, maxAgeReachedErr
	} else if expiryTimeMatureErr != nil {
		return false, matureExpiryTime, expiryTimeMatureErr
	}

	return false, time.Now(), nil
}

// hasExpiryTimeMatured checks whether the expiry time in the provided resource tag has matured. Returns true if time
// in the past (has matured)
func hasExpiryTimeMatured(resource *cloud.Resource, expiryTag *string, expiryTagNAValue, expiryTagFormat *string) (bool, time.Time, error) {
	expiryDateDiff := int64(0)
	expiryTime := time.Now()

	if len(*expiryTag) > 0 {
		expiryTimeString, expiryTagDefined := resource.Tags[*expiryTag]
		expiryTimeString = strings.TrimSpace(expiryTimeString)

		if expiryTagDefined && expiryTimeString != *expiryTagNAValue {
			gotExpiryTime, expiryTimeErr := time.Parse(*expiryTagFormat, expiryTimeString)
			expiryTime = gotExpiryTime

			if expiryTimeErr != nil {
				return false, expiryTime, fmt.Errorf("Could not parse the expiry time '%s' of resource '%s': %w", expiryTimeString, resource.ID, expiryTimeErr)
			}

			// substract now from the expiry time
			expiryDateDiff = expiryTime.UnixNano() - time.Now().UnixNano()
		}
	}

	return expiryDateDiff < 0, expiryTime, nil
}

// hasMaxAgeReached checks whether the provided resource has surpursed its maximum age.
// Returns true if it has
func hasMaxAgeReached(resource *cloud.Resource, maxAge *string) (bool, time.Time, error) {
	maxAgeDiff := int64(0)
	expiryTime := time.Now()

	if len(*maxAge) > 0 {
		maxAgeDuration, maxAgeDurationErr := time.ParseDuration(*maxAge)

		if maxAgeDurationErr != nil {
			return false, expiryTime, fmt.Errorf("Unable to process the value of maximum age string %s: %w", *maxAge, maxAgeDurationErr)
		}

		// add max age to start time of the resource
		expiryTime = resource.LaunchTime.Add(maxAgeDuration)

		// subtract now from (launchtime + max age)
		maxAgeDiff = expiryTime.UnixNano() - time.Now().UnixNano()
	}

	return maxAgeDiff < 0, expiryTime, nil
}

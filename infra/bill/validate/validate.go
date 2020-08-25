package validate

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/infra"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "validate"
const outputFormatPlain = "plain"
const outputFormatMarkdown = "markdown"
const requiredTagsEnvVar = "SRE_INFRA_BILL_REQUIRED_TAGS"
const dataFieldMissingTags = "missing-tags"

type Validate struct {
	helpFlag              *bool
	flagSet               *flag.FlagSet
	providerFlag          *flags.StringArray
	regionFlag            *flags.StringArray
	typeFlag              *flags.StringArray
	tagFlag               *flags.StringArray
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

func (validate *Validate) Init(helpFlagName string, helpFlagDescription string) {
	validate.flagSet = flag.NewFlagSet(validate.GetName(), flag.ExitOnError)
	validate.helpFlag = validate.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	validate.providerFlag, validate.regionFlag, validate.typeFlag, validate.tagFlag = infra.AddFilterFlags(validate.flagSet)
	validate.outputFormatFlag = validate.flagSet.String(
		"output-format",
		outputFormatPlain,
		fmt.Sprintf(
			"How to format the full output text. Possible values are '%s' and '%s'.",
			outputFormatPlain,
			outputFormatMarkdown))
	validate.showFlag,
		validate.hideHeadersFlag,
		validate.csvFlag,
		validate.fieldSeparatorFlag,
		validate.resourceSeparatorFlag,
		validate.listFieldsFlag,
		validate.defaultFieldValueFlag = infra.AddResourceTableFlags(validate.flagSet)
	validate.subCommands = []cli.Command{}
}

func (validate *Validate) GetName() string {
	return name
}

func (validate *Validate) GetDescription() string {
	return "Validates whether billing tags for resources are okay"
}

func (validate *Validate) GetFlagSet() *flag.FlagSet {
	return validate.flagSet
}

func (validate *Validate) GetSubCommands() []cli.Command {
	return validate.subCommands
}

func (validate *Validate) GetHelpFlag() *bool {
	return validate.helpFlag
}

func (validate *Validate) Process() {
	requiredTagsString := os.Getenv(requiredTagsEnvVar)
	if len(requiredTagsString) == 0 {
		notification.SendMessage(fmt.Sprintf("%s not set", requiredTagsEnvVar))
		cli.ExitCommandInterpretationError()
	}
	requiredTags := strings.Split(requiredTagsString, ",")

	allResources, resourcesErr := infra.GetAllCloudResources(infra.GetFiltersFromCommandFlags(validate.providerFlag, validate.regionFlag, validate.typeFlag, validate.tagFlag), true)
	if resourcesErr != nil {
		notification.SendMessage(resourcesErr.Error())
		cli.ExitCommandExecutionError()
	}

	var untaggedResources []*infra.Resource
	for _, curResource := range allResources {
		curTagKeys := infra.GetTagKeys(curResource)
		missingTags := getItemsInANotB(&requiredTags, &curTagKeys)
		if len(missingTags) > 0 {
			if curResource.Data == nil {
				curResource.Data = make(map[string]string)
			}

			curResource.Data[dataFieldMissingTags] = fmt.Sprintf("%v", missingTags)
			untaggedResources = append(untaggedResources, curResource)
		}
	}

	if len(untaggedResources) == 0 {
		return
	}

	rt := new(infra.ResourceTable)
	rt.Init(
		validate.showFlag,
		validate.hideHeadersFlag,
		validate.csvFlag,
		validate.fieldSeparatorFlag,
		validate.resourceSeparatorFlag,
		validate.listFieldsFlag,
		validate.defaultFieldValueFlag)
	table, tableErr := rt.Render(untaggedResources)
	if tableErr != nil {
		notification.SendMessage(tableErr.Error())
	}

	formattedOutput := ""
	errorMessage := "Cloud resources violating billing requirements:"
	switch *validate.outputFormatFlag {
	case outputFormatMarkdown:
		formattedOutput = fmt.Sprintf("%s\n```\n%s```", errorMessage, table)
	case outputFormatPlain:
		formattedOutput = fmt.Sprintf("%s\n%s", errorMessage, table)
	default:
		notification.SendMessage(fmt.Sprintf("Unrecognized output format '%s'", *validate.outputFormatFlag))
		cli.ExitCommandInterpretationError()
	}

	notification.SendMessage(formattedOutput)
	cli.ExitCommandExecutionError()
}

func getItemsInANotB(a *[]string, b *[]string) []string {
	itemsInANotB := []string{}

	for _, curAItem := range *a {
		if !stringInSlice(curAItem, b) {
			itemsInANotB = append(itemsInANotB, curAItem)
		}
	}

	return itemsInANotB
}

func stringInSlice(x string, slice *[]string) bool {
	for _, curItem := range *slice {
		if curItem == x {
			return true
		}
	}

	return false
}

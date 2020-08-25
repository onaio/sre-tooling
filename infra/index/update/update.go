package update

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/onaio/sre-tooling/infra/index/calculate"
	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/infra"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "update"
const updateTagsSeparator = ":"
const updateTagsFormatDescription = "\"TagName" + updateTagsSeparator + "<prefix to prepend before index>" + updateTagsSeparator + "<suffix to append after index>\""

type Update struct {
	helpFlag        *bool
	updateTagsFlag  *flags.StringArray
	providerFlag    *flags.StringArray
	regionFlag      *flags.StringArray
	typeFlag        *flags.StringArray
	tagFlag         *flags.StringArray
	idFlag          *string
	indexTagFlag    *string
	randomSleepFlag *int
	flagSet         *flag.FlagSet
	subCommands     []cli.Command
}

func (update *Update) Init(helpFlagName string, helpFlagDescription string) {
	update.flagSet = flag.NewFlagSet(update.GetName(), flag.ExitOnError)
	update.helpFlag = update.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	update.updateTagsFlag = new(flags.StringArray)
	update.flagSet.Var(update.updateTagsFlag, "update-tag", "Tag to update with index in the form "+updateTagsFormatDescription+". Multiple values can be provided by specifying multiple -update-tag")

	update.providerFlag,
		update.regionFlag,
		update.typeFlag,
		update.tagFlag,
		update.idFlag,
		update.indexTagFlag,
		update.randomSleepFlag = calculate.AddCalculateFlags(update.flagSet)

	update.subCommands = []cli.Command{}
}

func (update *Update) GetName() string {
	return name
}

func (update *Update) GetDescription() string {
	return "Updates a resource's tags based on the value of a newly calculated resource index in a group"
}

func (update *Update) GetFlagSet() *flag.FlagSet {
	return update.flagSet
}

func (update *Update) GetSubCommands() []cli.Command {
	return update.subCommands
}

func (update *Update) GetHelpFlag() *bool {
	return update.helpFlag
}

func (update *Update) Process() {
	if len(*update.providerFlag) != 1 {
		notification.SendMessage("Exactly one provider needs to be specified for the index update command to work")
		cli.ExitCommandInterpretationError()
	}

	if len(*update.typeFlag) != 1 {
		notification.SendMessage("Exactly one resource type needs to be specified for the index update command to work")
		cli.ExitCommandInterpretationError()
	}

	if len(*update.regionFlag) != 1 {
		notification.SendMessage("Exactly one region type needs to be specified for the index update command to work")
		cli.ExitCommandInterpretationError()
	}

	newIndex, newIndexErr := calculate.FetchAndCalculateResourceIndex(
		update.randomSleepFlag,
		update.providerFlag,
		update.regionFlag,
		update.typeFlag,
		update.tagFlag,
		update.idFlag,
		update.indexTagFlag,
	)

	if newIndexErr != nil {
		notification.SendMessage(newIndexErr.Error())
		cli.ExitCommandExecutionError()
	}

	newIndexStr := strconv.Itoa(newIndex)
	provider := (*update.providerFlag)[0]
	resourceType := (*update.typeFlag)[0]
	region := (*update.regionFlag)[0]
	resource := infra.Resource{
		Provider:     provider,
		ResourceType: resourceType,
		ID:           *update.idFlag,
	}

	// Update other tags
	for _, curTagDetails := range *update.updateTagsFlag {
		tagDetails := strings.Split(curTagDetails, updateTagsSeparator)
		if len(tagDetails) != 3 {
			notification.SendMessage(fmt.Sprintf("Tags to be updated should be provided in the format %s", updateTagsFormatDescription))
			cli.ExitCommandExecutionError()
		}
		tagValue := fmt.Sprintf("%s%s%s", tagDetails[1], newIndexStr, tagDetails[2])
		curUpdateErr := infra.UpdateResourceTag(&region, &resource, &tagDetails[0], &tagValue)

		if curUpdateErr != nil {
			notification.SendMessage(curUpdateErr.Error())
			cli.ExitCommandExecutionError()
		}
	}

	// Update the resource's index tag last
	updateErr := infra.UpdateResourceTag(&region, &resource, update.indexTagFlag, &newIndexStr)

	if updateErr != nil {
		notification.SendMessage(updateErr.Error())
		cli.ExitCommandExecutionError()
	}

	notification.SendMessage(newIndexStr)
}

package update

import (
	"flag"
	"time"

	"github.com/onaio/sre-tooling/infra/index/calculate"
	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/cloud"
	"github.com/onaio/sre-tooling/libs/notification"
	"github.com/onaio/sre-tooling/libs/numbers"
)

const name string = "update"

type Update struct {
	helpFlag           *bool
	updateHostnameFlag *bool
	hostnamePrefixFlag *string
	updateTagsFlag     *flags.StringArray
	providerFlag       *flags.StringArray
	regionFlag         *flags.StringArray
	typeFlag           *flags.StringArray
	tagFlag            *flags.StringArray
	idFlag             *string
	indexTagFlag       *string
	randomSleepFlag    *int
	flagSet            *flag.FlagSet
	subCommands        []cli.Command
}

func (update *Update) Init(helpFlagName string, helpFlagDescription string) {
	update.flagSet = flag.NewFlagSet(update.GetName(), flag.ExitOnError)
	update.helpFlag = flag.Bool(helpFlagName, false, helpFlagDescription)
	update.updateHostnameFlag = flag.Bool("update-hostname", false, "Whether to also update the hostname")
	update.idFlag = update.flagSet.String("id", "", "The ID of the resource to check the index")
	update.indexTagFlag = update.flagSet.String("index-tag", "", "The name of the tag containing the indexes of the resources")
	update.hostnamePrefixFlag = update.flagSet.String("hostname-prefix", "", "The prefix to append to the index when setting the hostname")
	update.randomSleepFlag = update.flagSet.Int("random-sleep", 0, "Sleep for a random number of seconds between 0 and what is defined before trying to calculate")
	update.flagSet.Var(update.updateTagsFlag, "update-tag", "Tag to update with index in the form \"TagName:prefix-to-prepend\". Multiple values can be provided by specifying multiple -update-tag")

	update.providerFlag,
		update.regionFlag,
		update.typeFlag,
		update.tagFlag = cloud.AddFilterFlags(update.flagSet)
	flag.Parse()
}

func (update *Update) GetName() string {
	return name
}

func (update *Update) GetDescription() string {
	return "Updates the index as well as resource names tied to the index"
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
	if len(*update.idFlag) == 0 {
		notification.SendMessage("You need to provide the ID of the resource you want to check its index")
		cli.ExitCommandInterpretationError()
	}
	if len(*update.indexTagFlag) == 0 {
		notification.SendMessage("You need to provide the name of the tag containing resource indexes")
		cli.ExitCommandInterpretationError()
	}

	// Sleep for some random amount of time
	sleepTime := numbers.GetRandomInt(*update.randomSleepFlag)
	if sleepTime > 0 {
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	// Get the new index
	newIndex, newIndexErr := calculate.GetNewResourceIndex(
		calculate.idFlag,
		calculate.indexTagFlag,
		allResources)
	if newIndexErr != nil {
		notification.SendMessage(newIndexErr.Error())
		cli.ExitCommandExecutionError()
	}

	// Update the resource's index
}

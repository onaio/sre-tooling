package query

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/infra"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "query"

type Query struct {
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
	subCommands           []cli.Command
}

func (query *Query) Init(helpFlagName string, helpFlagDescription string) {
	query.flagSet = flag.NewFlagSet(query.GetName(), flag.ExitOnError)
	query.helpFlag = query.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	query.providerFlag, query.regionFlag, query.typeFlag, query.tagFlag = infra.AddFilterFlags(query.flagSet)
	query.showFlag,
		query.hideHeadersFlag,
		query.csvFlag,
		query.fieldSeparatorFlag,
		query.resourceSeparatorFlag,
		query.listFieldsFlag,
		query.defaultFieldValueFlag = infra.AddResourceTableFlags(query.flagSet)

	query.subCommands = []cli.Command{}
}

func (query *Query) GetName() string {
	return name
}

func (query *Query) GetDescription() string {
	return "Queries infrastructure resources and prints out requested fields"
}

func (query *Query) GetFlagSet() *flag.FlagSet {
	return query.flagSet
}

func (query *Query) GetSubCommands() []cli.Command {
	return query.subCommands
}

func (query *Query) GetHelpFlag() *bool {
	return query.helpFlag
}

func (query *Query) Process() {
	if len(*query.regionFlag) == 0 && len(*query.typeFlag) == 0 && len(*query.tagFlag) == 0 {
		notification.SendMessage("You need to filter resources using at least one region, type, or tag")
		cli.ExitCommandInterpretationError()
	}

	allResources, resourcesErr := infra.GetAllCloudResources(infra.GetFiltersFromCommandFlags(query.providerFlag, query.regionFlag, query.typeFlag, query.tagFlag), true)
	if resourcesErr != nil {
		notification.SendMessage(resourcesErr.Error())
		cli.ExitCommandExecutionError()
	}

	rt := new(infra.ResourceTable)
	rt.Init(
		query.showFlag,
		query.hideHeadersFlag,
		query.csvFlag,
		query.fieldSeparatorFlag,
		query.resourceSeparatorFlag,
		query.listFieldsFlag,
		query.defaultFieldValueFlag)
	output, outputErr := rt.Render(allResources)
	if outputErr != nil {
		notification.SendMessage(outputErr.Error())
		cli.ExitCommandExecutionError()
	}

	notification.SendMessage(output)
}

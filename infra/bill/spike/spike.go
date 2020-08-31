package spike

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cloud"

	"github.com/onaio/sre-tooling/libs/cli"

	"github.com/onaio/sre-tooling/libs/cli/flags"
)

const name string = "spike"

type Spike struct {
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

func (spike *Spike) Init(helpFlagName string, helpFlagDescription string) {
	spike.flagSet = flag.NewFlagSet(spike.GetName(), flag.ExitOnError)
	spike.helpFlag = spike.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	spike.providerFlag, spike.regionFlag, spike.typeFlag, spike.tagFlag = cloud.AddFilterFlags(spike.flagSet)
	spike.showFlag,
		spike.hideHeadersFlag,
		spike.csvFlag,
		spike.fieldSeparatorFlag,
		spike.resourceSeparatorFlag,
		spike.listFieldsFlag,
		spike.defaultFieldValueFlag = cloud.AddResourceTableFlags(spike.flagSet)

	spike.subCommands = []cli.Command{}
}

func (spike *Spike) GetName() string {
	return name
}

func (spike *Spike) GetDescription() string {
	return "Calculate cost usage spikes for resources against a threshold"
}

func (spike *Spike) GetFlagSet() *flag.FlagSet {
	return spike.flagSet
}

func (spike *Spike) GetSubCommands() []cli.Command {
	return spike.subCommands
}

func (spike *Spike) GetHelpFlag() *bool {
	return spike.helpFlag
}

func (spike *Spike) Process() {}

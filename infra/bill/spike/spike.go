package spike

import (
	"flag"
	"fmt"

	"github.com/onaio/sre-tooling/libs/infra"

	"github.com/onaio/sre-tooling/libs/cli"

	"github.com/onaio/sre-tooling/libs/cli/flags"
)

const name string = "spike"
const outputFormatPlain = "plain"
const outputFormatMarkdown = "markdown"

type Spike struct {
	helpFlag              *bool
	flagSet               *flag.FlagSet
	granularityFlag       *string
	startDateFlag         *string
	endDateFlag           *string
	showFlag              *flags.StringArray
	providerFlag          *flags.StringArray
	hideHeadersFlag       *bool
	csvFlag               *bool
	fieldSeparatorFlag    *string
	resourceSeparatorFlag *string
	listFieldsFlag        *bool
	defaultFieldValueFlag *string
	outputFormatFlag      *string
	subCommands           []cli.Command
}

func (spike *Spike) Init(helpFlagName string, helpFlagDescription string) {
	spike.flagSet = flag.NewFlagSet(spike.GetName(), flag.ExitOnError)
	spike.helpFlag = spike.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	providerFlag := new(flags.StringArray)
	spike.flagSet.Var(
		providerFlag,
		"filter-provider",
		"Name of provider to filter using. Multiple values can be provided by specifying multiple -filter-provider",
	)
	spike.providerFlag = providerFlag
	spike.granularityFlag = spike.flagSet.String(
		"granularity",
		"DAILY",
		"Cost granularity to use. Can be MONTHLY, DAILY or HOURLY.",
	)
	spike.startDateFlag = spike.flagSet.String(
		"start-date",
		"",
		"Start date for retrieving costs. Start date is inclusive. Should be in the format YYYY-MM-DD.",
	)
	spike.endDateFlag = spike.flagSet.String(
		"end-date",
		"",
		"End date for retrieving costs. End date is exclusive. Should be in the format YYYY-MM-DD.",
	)
	spike.outputFormatFlag = spike.flagSet.String(
		"output-format",
		outputFormatPlain,
		fmt.Sprintf(
			"How to format the full output text. Possible values are '%s' and '%s'.",
			outputFormatPlain,
			outputFormatMarkdown))

	spike.showFlag,
		spike.hideHeadersFlag,
		spike.csvFlag,
		spike.fieldSeparatorFlag,
		spike.resourceSeparatorFlag,
		spike.listFieldsFlag,
		spike.defaultFieldValueFlag = infra.AddResourceTableFlags(spike.flagSet)

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

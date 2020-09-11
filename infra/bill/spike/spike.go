package spike

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onaio/sre-tooling/libs/types"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/infra"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "spike"
const outputFormatPlain = "plain"
const outputFormatMarkdown = "markdown"
const requiredThresholdEnvVar = "SRE_INFRA_COST_SPIKE_THRESHOLD"
const layoutISO = "2006-01-02"

type Spike struct {
	helpFlag              *bool
	flagSet               *flag.FlagSet
	providerFlag          *flags.StringArray
	regionFlag            *flags.StringArray
	typeFlag              *flags.StringArray
	tagFlag               *flags.StringArray
	granularityFlag       *string
	startDateFlag         *string
	endDateFlag           *string
	groupByFlag           *flags.StringArray
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

func (spike *Spike) Init(helpFlagName string, helpFlagDescription string) {
	spike.flagSet = flag.NewFlagSet(spike.GetName(), flag.ExitOnError)
	spike.helpFlag = spike.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	spike.AddFilterFlags()
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

func (spike *Spike) Process() {
	thresholdString := os.Getenv(requiredThresholdEnvVar)
	if len(thresholdString) == 0 {
		notification.SendMessage(fmt.Sprintf("%s not set", requiredThresholdEnvVar))
		cli.ExitCommandInterpretationError()
	}
	threshold, parseErr := strconv.ParseFloat(thresholdString, 64)
	if parseErr != nil {
		notification.SendMessage(fmt.Sprintf("Unable to parse %s %s", requiredThresholdEnvVar, thresholdString))
		cli.ExitCommandExecutionError()
	}

	// parse startDate and endDate
	startDate, startDateParseErr := time.Parse(layoutISO, *spike.startDateFlag)
	if startDateParseErr != nil {
		notification.SendMessage(fmt.Sprintf("Unable to parse -start-date %s", *spike.startDateFlag))
		cli.ExitCommandInterpretationError()
	}
	endDate, endDateParseErr := time.Parse(layoutISO, *spike.endDateFlag)
	if endDateParseErr != nil {
		notification.SendMessage(fmt.Sprintf("Unable to parse -end-date %s", *spike.endDateFlag))
		cli.ExitCommandInterpretationError()
	}
	daysDiff := endDate.Sub(startDate).Hours() / -24

	// Calculate current period's costs and usages
	costAndUsageFilter := spike.GetFiltersFromFlags()
	curProviderCosts, err := infra.GetCostsAndUsages(costAndUsageFilter)
	if err != nil {
		notification.SendMessage(err.Error())
		cli.ExitCommandExecutionError()
	}

	// Calculate previous period's costs and usages
	prevStartDate := startDate.AddDate(0, 0, int(daysDiff))
	costAndUsageFilter.StartDate = prevStartDate.Format(layoutISO)
	costAndUsageFilter.EndDate = startDate.Format(layoutISO)
	prevProviderCosts, err := infra.GetCostsAndUsages(costAndUsageFilter)
	if err != nil {
		notification.SendMessage(err.Error())
		cli.ExitCommandExecutionError()
	}

	// Calculate spike
	spikedCosts := []*types.CostSpikeOutput{}
	for providerName, curCosts := range curProviderCosts {
		prevCosts := prevProviderCosts[providerName]

		for name, curAmount := range curCosts.Groups {
			prevAmount := prevCosts.Groups[name]
			increaseRate := ((curAmount - prevAmount) / curAmount) * 100
			if increaseRate > threshold {
				spikedCosts = append(spikedCosts, &types.CostSpikeOutput{
					Provider: providerName,
					GroupKey: name,
					CurPeriod: &types.CostAndUsagePeriod{
						StartDate: curCosts.Period.StartDate,
						EndDate:   curCosts.Period.EndDate,
					},
					PrevPeriod: &types.CostAndUsagePeriod{
						StartDate: prevCosts.Period.StartDate,
						EndDate:   prevCosts.Period.EndDate,
					},
					CurPeriodAmount:  curAmount,
					PrevPeriodAmount: prevAmount,
					IncreaseRate:     increaseRate,
				})
			}
		}
	}

	if len(spikedCosts) == 0 {
		return
	}

	rt := new(infra.ResourceTable)
	rt.Init(
		spike.showFlag,
		spike.hideHeadersFlag,
		spike.csvFlag,
		spike.fieldSeparatorFlag,
		spike.resourceSeparatorFlag,
		spike.listFieldsFlag,
		spike.defaultFieldValueFlag)
	table, tableErr := rt.RenderCostSpikes(spikedCosts)
	if tableErr != nil {
		notification.SendMessage(tableErr.Error())
	}

	formattedOutput := ""
	message := "Cost spikes"
	switch *spike.outputFormatFlag {
	case outputFormatMarkdown:
		formattedOutput = fmt.Sprintf("%s\n```\n%s```", message, table)
	case outputFormatPlain:
		formattedOutput = fmt.Sprintf("%s\n%s", message, table)
	default:
		notification.SendMessage(fmt.Sprintf("Unrecognized output format '%s'", *spike.outputFormatFlag))
		cli.ExitCommandInterpretationError()
	}

	notification.SendMessage(formattedOutput)
	cli.ExitCommandExecutionError()
}

func (spike *Spike) GetFiltersFromFlags() *types.CostAndUsageFilter {
	filter := &types.CostAndUsageFilter{}
	if len(*spike.providerFlag) > 0 {
		filter.Providers = *spike.providerFlag
	}
	if len(*spike.regionFlag) > 0 {
		filter.Regions = *spike.regionFlag
	}
	if len(*spike.typeFlag) > 0 {
		filter.ResourceTypes = *spike.typeFlag
	}
	if len(*spike.tagFlag) > 0 {
		for _, tagPair := range *spike.tagFlag {
			tagKeyValue := strings.Split(tagPair, ":")
			if len(tagKeyValue) == 2 {
				if filter.Tags == nil {
					filter.Tags = make(map[string]string)
				}
				filter.Tags[tagKeyValue[0]] = tagKeyValue[1]
			}
		}
	}
	if len(*spike.groupByFlag) > 0 {
		for _, groupByPair := range *spike.groupByFlag {
			groupByValue := strings.Split(groupByPair, ":")
			if len(groupByValue) == 2 {
				if filter.GroupBy == nil {
					filter.GroupBy = make(map[string]string)
				}
				filter.GroupBy[groupByValue[0]] = groupByValue[1]
			}
		}
	}
	filter.Granularity = *spike.granularityFlag
	filter.StartDate = *spike.startDateFlag
	filter.EndDate = *spike.endDateFlag

	return filter
}

func (spike *Spike) AddFilterFlags() {
	spike.providerFlag, spike.regionFlag, spike.typeFlag, spike.tagFlag = infra.AddFilterFlags(spike.flagSet)

	// add cost spike flags
	spike.granularityFlag = spike.flagSet.String(
		"granularity",
		"DAILY",
		"Cost granularity to use. Can be MONTHLY, DAILY or HOURLY.",
	)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)
	spike.startDateFlag = spike.flagSet.String(
		"start-date",
		startDate.Format(layoutISO),
		"Start date for retrieving costs. Start date is inclusive. Should be in the format yyyy-MM-dd.",
	)
	spike.endDateFlag = spike.flagSet.String(
		"end-date",
		endDate.Format(layoutISO),
		"End date for retrieving costs. End date is exclusive. Should be in the format yyyy-MM-dd.",
	)
	groupFlag := new(flags.StringArray)
	spike.flagSet.Var(groupFlag, "group-by", "Field to group costs by. Use the format \"groupType:groupValue\". Multiple values can be provided by specifying multiple -group-by")
	spike.groupByFlag = groupFlag
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
}

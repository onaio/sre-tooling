package query

import (
	"bytes"
	"flag"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/cloud"
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
	csv                   *bool
	fieldSeparatorFlag    *string
	resourceSeparatorFlag *string
	listFieldsFlag        *bool
	defaultFieldValueFlag *string
	subCommands           []cli.Command
}

func (query *Query) Init(helpFlagName string, helpFlagDescription string) {
	query.flagSet = flag.NewFlagSet(query.GetName(), flag.ExitOnError)
	query.helpFlag = query.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	query.providerFlag, query.regionFlag, query.typeFlag, query.tagFlag = cloud.AddFilterFlags(query.flagSet)
	query.showFlag = new(flags.StringArray)
	query.flagSet.Var(query.showFlag, "show", "Resource's field to be shown e.g tag:Name. Multiple values can be provided by specifying multiple -show")
	query.hideHeadersFlag = query.flagSet.Bool("hide-headers", false, "Whether to hide the names of the fields shown")
	query.csv = query.flagSet.Bool("csv", false, "Whether output should follow the CSV standard")
	query.fieldSeparatorFlag = query.flagSet.String("field-separator", "\t", "What to use to separate displayed fields. Only applies if -csv is enabled")
	query.resourceSeparatorFlag = query.flagSet.String("resource-separator", "\n", "What to use to separate displayed resources. Only applies of -list-fields or -csv is enabled")
	query.listFieldsFlag = query.flagSet.Bool("list-fields", false, "Whether to just list the fields available to be displayed, instead of the resources")
	query.defaultFieldValueFlag = query.flagSet.String("default-field-value", "", "Text that should be set if a resource's queried field doesn't have a value")
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
		cli.ExitCommandExecutionError()
	}

	allResources, resourcesErr := cloud.GetAllCloudResources(cloud.GetFiltersFromCommandFlags(query.providerFlag, query.regionFlag, query.typeFlag, query.tagFlag), true)
	if resourcesErr != nil {
		notification.SendMessage(resourcesErr.Error())
		cli.ExitCommandExecutionError()
	}

	rows := make([]map[string]string, len(allResources))
	headers := make(map[string]bool)
	for rowIndex, curResource := range allResources {
		headers, rows = query.addFields(headers, rows, rowIndex, curResource.Tags, "tag")
		headers, rows = query.addFields(headers, rows, rowIndex, curResource.Properties, "property")
	}

	output := ""
	displayedHeaders := []string{}
	for curHeader, includeHeader := range headers {
		if includeHeader {
			displayedHeaders = append(displayedHeaders, curHeader)
		}
	}
	sort.Strings(displayedHeaders)
	if *query.listFieldsFlag {
		output = strings.Join(displayedHeaders, *query.resourceSeparatorFlag)
	} else {
		buf := new(bytes.Buffer)
		table := tablewriter.NewWriter(buf)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoFormatHeaders(false)
		if *query.csv {
			table.SetTablePadding("")
			table.SetAutoWrapText(false)
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetCenterSeparator("")
			table.SetRowLine(false)
			table.SetNewLine(*query.resourceSeparatorFlag)
			table.SetColumnSeparator(*query.fieldSeparatorFlag)
		}

		if !*query.hideHeadersFlag {
			table.SetHeader(displayedHeaders)
		}

		for _, curRow := range rows {
			rowSlice := []string{}
			for _, curHeader := range displayedHeaders {
				fieldVal := *query.defaultFieldValueFlag
				if curField, ok := curRow[curHeader]; ok {
					fieldVal = curField
				}

				rowSlice = append(rowSlice, fieldVal)
			}
			table.Append(rowSlice)
		}

		table.Render()
		output = buf.String()
	}

	notification.SendMessage(output)
}

func (query *Query) addFields(headers map[string]bool, rows []map[string]string, rowIndex int, newFields map[string]string, fieldType string) (map[string]bool, []map[string]string) {
	if rows[rowIndex] == nil {
		rows[rowIndex] = make(map[string]string)
	}
	for curKey, curVal := range newFields {
		fieldName := fieldType + ":" + curKey
		rows[rowIndex][fieldName] = curVal
		headers[fieldName] = query.considerField(fieldName)
	}

	return headers, rows
}

func (query *Query) considerField(field string) bool {
	if len(*query.showFlag) == 0 {
		return true
	}

	for _, curField := range *query.showFlag {
		if curField == field {
			return true
		}
	}

	return false
}

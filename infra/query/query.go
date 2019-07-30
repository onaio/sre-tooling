package query

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

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
	fieldSeparatorFlag    *string
	resourceSeparatorFlag *string
	listFieldsFlag        *bool
}

func (query *Query) Init(helpFlagName string, helpFlagDescription string) {
	query.flagSet = flag.NewFlagSet(query.GetName(), flag.ExitOnError)
	query.helpFlag = query.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	query.providerFlag, query.regionFlag, query.typeFlag, query.tagFlag = cloud.AddFilterFlags(query.flagSet)
	query.showFlag = new(flags.StringArray)
	query.flagSet.Var(query.showFlag, "show", "Resource's field to be shown e.g tag:Name. Multiple values can be provided by specifying multiple -show")
	query.hideHeadersFlag = query.flagSet.Bool("hide-headers", false, "Whether to hide the names of the fields shown")
	query.fieldSeparatorFlag = query.flagSet.String("field-separator", "\t", "What to use to separate displayed fields")
	query.resourceSeparatorFlag = query.flagSet.String("resource-separator", "\n", "What to use to separate displayed resources")
	query.listFieldsFlag = query.flagSet.Bool("list-fields", false, "Whether to just list the fields available to be displayed, instead of the resources")
}

func (query *Query) GetName() string {
	return name
}

func (query *Query) GetDescription() string {
	return "Queries infrastructure resources and prints out requested fields"
}

func (query *Query) ParseArgs(args []string) {
	query.flagSet.Parse(args)
	if *query.helpFlag {
		query.printHelp()
	} else {
		query.query()
	}
}

func (query *Query) query() {
	if len(*query.regionFlag) == 0 && len(*query.typeFlag) == 0 && len(*query.tagFlag) == 0 {
		notification.SendMessage("You need to filter resources using at least one region, type, or tag")
		os.Exit(1)
	}

	allResources, resourcesErr := cloud.GetAllCloudResources(cloud.GetFiltersFromCommandFlags(query.providerFlag, query.regionFlag, query.typeFlag, query.tagFlag), true)
	if resourcesErr != nil {
		notification.SendMessage(resourcesErr.Error())
		os.Exit(1)
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
		if !*query.hideHeadersFlag {
			output = strings.Join(displayedHeaders, *query.fieldSeparatorFlag) + *query.resourceSeparatorFlag
		}

		for rowIndex, curRow := range rows {
			rowOutput := ""
			for _, curHeader := range displayedHeaders {
				if len(rowOutput) != 0 {
					rowOutput = rowOutput + *query.fieldSeparatorFlag
				}
				if curField, ok := curRow[curHeader]; ok {
					rowOutput = rowOutput + curField
				} else {
					rowOutput = rowOutput + " "
				}
			}
			output = output + rowOutput
			if rowIndex < len(rows)-1 {
				output = output + *query.resourceSeparatorFlag
			}
		}
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

func (query *Query) printHelp() {
	fmt.Println(query.GetDescription())
	query.flagSet.PrintDefaults()
}

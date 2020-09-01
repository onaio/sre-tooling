package infra

// This file contains all the CLI helpers related to the cloud package

import (
	"bytes"
	"flag"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/types"
)

// ResourceTable is responsible for rendering cloud resources in a table
type ResourceTable struct {
	showFlag              *flags.StringArray
	hideHeadersFlag       *bool
	csvFlag               *bool
	fieldSeparatorFlag    *string
	resourceSeparatorFlag *string
	listFieldsFlag        *bool
	defaultFieldValueFlag *string
}

func (rt *ResourceTable) Init(
	showFlag *flags.StringArray,
	hideHeadersFlag *bool,
	csvFlag *bool,
	fieldSeparatorFlag *string,
	resourceSeparatorFlag *string,
	listFieldsFlag *bool,
	defaultFieldValueFlag *string) {
	rt.showFlag = showFlag
	rt.hideHeadersFlag = hideHeadersFlag
	rt.csvFlag = csvFlag
	rt.fieldSeparatorFlag = fieldSeparatorFlag
	rt.resourceSeparatorFlag = resourceSeparatorFlag
	rt.listFieldsFlag = listFieldsFlag
	rt.defaultFieldValueFlag = defaultFieldValueFlag
}

func AddResourceTableFlags(flagSet *flag.FlagSet) (*flags.StringArray, *bool, *bool, *string, *string, *bool, *string) {
	showFlag := new(flags.StringArray)
	flagSet.Var(showFlag, "show", "Resource's field to be shown e.g tag:Name. Multiple values can be provided by specifying multiple -show")
	hideHeadersFlag := flagSet.Bool("hide-headers", false, "Whether to hide the names of the fields shown")
	csvFlag := flagSet.Bool("csv", false, "Whether output should follow the CSV standard")
	fieldSeparatorFlag := flagSet.String("field-separator", "\t", "What to use to separate displayed fields. Only applies if -csv is enabled")
	resourceSeparatorFlag := flagSet.String("resource-separator", "\n", "What to use to separate displayed resources. Only applies of -list-fields or -csv is enabled")
	listFieldsFlag := flagSet.Bool("list-fields", false, "Whether to just list the fields available to be displayed, instead of the resources")
	defaultFieldValueFlag := flagSet.String("default-field-value", "", "Text that should be set if a resource's queried field doesn't have a value")

	return showFlag, hideHeadersFlag, csvFlag, fieldSeparatorFlag, resourceSeparatorFlag, listFieldsFlag, defaultFieldValueFlag
}

func (rt *ResourceTable) Render(allResources []*types.InfraResource) (string, error) {
	rows := make([]map[string]string, len(allResources))
	headers := make(map[string]bool)
	for rowIndex, curResource := range allResources {
		headers, rows = rt.addResourceTableFields(headers, rows, rowIndex, curResource.Tags, "tag")
		headers, rows = rt.addResourceTableFields(headers, rows, rowIndex, curResource.Properties, "property")
		headers, rows = rt.addResourceTableFields(headers, rows, rowIndex, curResource.Data, "data")
	}

	output := ""
	displayedHeaders := []string{}
	if len(*rt.showFlag) > 0 {
		// User provided a list of columns they want to see.
		// Show these properties in the order user specified.
		for _, curHeaderName := range *rt.showFlag {
			showHeader, headerExists := headers[curHeaderName]
			if headerExists && showHeader {
				displayedHeaders = append(displayedHeaders, curHeaderName)
			}
		}
	} else {
		// User didn't specify which columns they want to see.
		// Go through all the returned columns, see if they are
		// flagged as viewable, and sort them alphabetically.
		for curHeader, includeHeader := range headers {
			if includeHeader {
				displayedHeaders = append(displayedHeaders, curHeader)
			}
		}
		sort.Strings(displayedHeaders)
	}

	if *rt.listFieldsFlag {
		output = strings.Join(displayedHeaders, *rt.resourceSeparatorFlag)
	} else {
		buf := new(bytes.Buffer)
		table := tablewriter.NewWriter(buf)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoFormatHeaders(false)
		if *rt.csvFlag {
			table.SetTablePadding("")
			table.SetAutoWrapText(false)
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetCenterSeparator("")
			table.SetRowLine(false)
			table.SetNewLine(*rt.resourceSeparatorFlag)
			table.SetColumnSeparator(*rt.fieldSeparatorFlag)
		}

		if !*rt.hideHeadersFlag {
			table.SetHeader(displayedHeaders)
		}

		for _, curRow := range rows {
			rowSlice := []string{}
			for _, curHeader := range displayedHeaders {
				fieldVal := *rt.defaultFieldValueFlag
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

	return output, nil
}

func (rt *ResourceTable) addResourceTableFields(headers map[string]bool, rows []map[string]string, rowIndex int, newFields map[string]string, fieldType string) (map[string]bool, []map[string]string) {
	if rows[rowIndex] == nil {
		rows[rowIndex] = make(map[string]string)
	}
	for curKey, curVal := range newFields {
		fieldName := fieldType + ":" + curKey
		rows[rowIndex][fieldName] = curVal
		headers[fieldName] = rt.considerResourceTableField(fieldName)
	}

	return headers, rows
}

func (rt *ResourceTable) considerResourceTableField(field string) bool {
	if len(*rt.showFlag) == 0 {
		return true
	}

	for _, curField := range *rt.showFlag {
		if curField == field {
			return true
		}
	}

	return false
}

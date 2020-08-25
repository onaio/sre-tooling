package infra

import (
	"strings"
	"testing"
	"time"

	"github.com/onaio/sre-tooling/libs/cli/flags"
)

// Test whether setting the "hide headers" argument in the resource table actually works
func TestResourceTableRenderHeader(t *testing.T) {
	showFlag := new(flags.StringArray)
	showFlag.Set("tag:Name")
	csvFlag := false
	fieldSeparatorFlag := "\t"
	resourceSeparatorFlag := "\n"
	listFieldsFlag := false
	defaultFieldValueFlag := ""

	resource1 := Resource{
		Provider:   "testProvider",
		ID:         "blah",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{"Name": "testResourceNameTag"},
		Properties: map[string]string{}}
	var resources []*Resource
	resources = append(resources, &resource1)

	t.Run("no-headers", func(t *testing.T) {
		hideHeadersFlag := true

		rt := new(ResourceTable)
		rt.Init(
			showFlag,
			&hideHeadersFlag,
			&csvFlag,
			&fieldSeparatorFlag,
			&resourceSeparatorFlag,
			&listFieldsFlag,
			&defaultFieldValueFlag)
		table, tableErr := rt.Render(resources)
		if tableErr != nil {
			t.Errorf("Not expecting an error to be returned: %w", tableErr)
		}
		if strings.Contains(table, "tag:Name") {
			t.Errorf("Not expecting the tag:Name header")
		}
	})

	t.Run("with-headers", func(t *testing.T) {
		hideHeadersFlag := false

		rt := new(ResourceTable)
		rt.Init(
			showFlag,
			&hideHeadersFlag,
			&csvFlag,
			&fieldSeparatorFlag,
			&resourceSeparatorFlag,
			&listFieldsFlag,
			&defaultFieldValueFlag)
		table, tableErr := rt.Render(resources)
		if tableErr != nil {
			t.Errorf("Not expecting an error to be returned: %w", tableErr)
		}
		if !strings.Contains(table, "tag:Name") {
			t.Errorf("Expecting the tag:Name header to be present")
		}
	})
}

// Test whether the the resource table is able to render the different field types
func TestResourceTableShowColumns(t *testing.T) {
	nameTag := "Name"
	publicIPProperty := "public-ip"
	missingTagsData := "missing-tags"
	csvFlag := false
	fieldSeparatorFlag := "\t"
	resourceSeparatorFlag := "\n"
	listFieldsFlag := false
	defaultFieldValueFlag := ""
	hideHeadersFlag := false

	nameTagValue := "testResourceNameTag"
	ipPropertyValue := "0.0.0.0"
	missingTagsDataValue := "EndDate"

	resource1 := Resource{
		Provider:   "testProvider",
		ID:         "blah",
		Location:   "eu-central-1a",
		LaunchTime: time.Now(),
		Tags:       map[string]string{nameTag: nameTagValue},
		Properties: map[string]string{publicIPProperty: ipPropertyValue},
		Data:       map[string]string{missingTagsData: missingTagsDataValue}}
	var resources []*Resource
	resources = append(resources, &resource1)

	t.Run("show-flags-set", func(t *testing.T) {
		showFlag := new(flags.StringArray)
		showFlag.Set("tag:" + nameTag)
		showFlag.Set("property:" + publicIPProperty)
		showFlag.Set("data:" + missingTagsData)

		rt := new(ResourceTable)
		rt.Init(
			showFlag,
			&hideHeadersFlag,
			&csvFlag,
			&fieldSeparatorFlag,
			&resourceSeparatorFlag,
			&listFieldsFlag,
			&defaultFieldValueFlag)
		table, tableErr := rt.Render(resources)
		if tableErr != nil {
			t.Errorf("Not expecting an error to be returned: %w", tableErr)
		}
		if !strings.Contains(table, nameTagValue) {
			t.Errorf("Expecting the '%s' tag value '%s' to be present", nameTag, nameTagValue)
		}
		if !strings.Contains(table, ipPropertyValue) {
			t.Errorf("Expecting the '%s' property value '%s' to be present", publicIPProperty, ipPropertyValue)
		}
		if !strings.Contains(table, missingTagsDataValue) {
			t.Errorf("Expecting the '%s' data value '%s' to be present", missingTagsData, missingTagsDataValue)
		}
	})

	t.Run("show-flags-partially-set", func(t *testing.T) {
		showFlag := new(flags.StringArray)
		showFlag.Set("tag:" + nameTag)

		rt := new(ResourceTable)
		rt.Init(
			showFlag,
			&hideHeadersFlag,
			&csvFlag,
			&fieldSeparatorFlag,
			&resourceSeparatorFlag,
			&listFieldsFlag,
			&defaultFieldValueFlag)
		table, tableErr := rt.Render(resources)
		if tableErr != nil {
			t.Errorf("Not expecting an error to be returned: %w", tableErr)
		}
		if !strings.Contains(table, nameTagValue) {
			t.Errorf("Expecting the '%s' tag value '%s' to be present", nameTag, nameTagValue)
		}
		if strings.Contains(table, ipPropertyValue) {
			t.Errorf("Not expecting the '%s' property value '%s' to be present", publicIPProperty, ipPropertyValue)
		}
		if strings.Contains(table, missingTagsDataValue) {
			t.Errorf("Not expecting the '%s' data value '%s' to be present", missingTagsData, missingTagsDataValue)
		}
	})

	t.Run("show-flags-not-set", func(t *testing.T) {
		showFlag := new(flags.StringArray)

		rt := new(ResourceTable)
		rt.Init(
			showFlag,
			&hideHeadersFlag,
			&csvFlag,
			&fieldSeparatorFlag,
			&resourceSeparatorFlag,
			&listFieldsFlag,
			&defaultFieldValueFlag)
		table, tableErr := rt.Render(resources)
		if tableErr != nil {
			t.Errorf("Not expecting an error to be returned: %w", tableErr)
		}
		if !strings.Contains(table, nameTagValue) {
			t.Errorf("Expecting the '%s' tag value '%s' to be present", nameTag, nameTagValue)
		}
		if !strings.Contains(table, ipPropertyValue) {
			t.Errorf("Expecting the '%s' property value '%s' to be present", publicIPProperty, ipPropertyValue)
		}
		if !strings.Contains(table, missingTagsDataValue) {
			t.Errorf("Expecting the '%s' data value '%s' to be present", missingTagsData, missingTagsDataValue)
		}
	})
}

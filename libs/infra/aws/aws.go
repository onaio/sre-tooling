package aws

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/onaio/sre-tooling/libs/types"
)

const awsProviderName string = "AWS"

type AWS struct {
	resourceTypes []resourceType
	dataMutex     sync.Mutex
}

type resourceType interface {
	init(*session.Session) error
	getName() string
	getResources(filter *types.InfraFilter) ([]*types.InfraResource, error)
	updateResourceTag(resource *types.InfraResource, tagKey *string, tagValue *string) error
	updateResourceState(resource *types.InfraResource, safe bool, state string) error
}

type awsResourceHandler func(resourceType resourceType, resources []*types.InfraResource, err error)

func (aws *AWS) GetName() string {
	return awsProviderName
}

func (a *AWS) Init() error {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2 := new(EC2)
	ec2Err := ec2.init(session)
	if ec2Err != nil {
		return ec2Err
	}

	a.resourceTypes = []resourceType{
		ec2,
	}

	return nil
}

func (a *AWS) GetCostsAndUsages(filter *types.CostAndUsageFilter) (*types.CostAndUsageOutput, error) {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ceService := costexplorer.New(session)

	groupDefinitions := []*costexplorer.GroupDefinition{}
	for groupType, groupKey := range filter.GroupBy {
		groupDefinitions = append(groupDefinitions, &costexplorer.GroupDefinition{
			Type: aws.String(groupType),
			Key:  aws.String(groupKey),
		})
	}
	filterExpression := constructFilterExpression(filter)
	costAndUsageInput := &costexplorer.GetCostAndUsageInput{
		Filter:      filterExpression,
		Granularity: aws.String(filter.Granularity),
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(filter.StartDate),
			End:   aws.String(filter.EndDate),
		},
		Metrics: []*string{
			aws.String("UNBLENDED_COST"),
		},
		GroupBy: groupDefinitions,
	}
	costAndUsageOutput, ceErr := ceService.GetCostAndUsage(costAndUsageInput)
	if ceErr != nil {
		return nil, ceErr
	}

	groupAmounts := make(map[string]float64)
	for _, resultsByTime := range costAndUsageOutput.ResultsByTime {
		for _, groups := range resultsByTime.Groups {
			for _, metrics := range groups.Metrics {
				if amount, err := strconv.ParseFloat(*metrics.Amount, 64); err == nil {
					groupAmounts[*groups.Keys[0]] += amount
				}
			}
		}
	}

	costsAndUsages := &types.CostAndUsageOutput{
		Provider:  a.GetName(),
		Groups:    groupAmounts,
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
	}

	return costsAndUsages, nil
}

func (a *AWS) GetResources(filter *types.InfraFilter) ([]*types.InfraResource, error) {
	allResources := []*types.InfraResource{}
	dataWG := new(sync.WaitGroup)

	var finalErr error
	for _, curType := range a.resourceTypes {
		if considerResourceType(resourceTypeEc2, filter) {
			handler := func(resourceType resourceType, resources []*types.InfraResource, err error) {
				a.dataMutex.Lock()
				allResources = append(allResources, resources...)
				if err != nil {
					finalErr = err
				}
				a.dataMutex.Unlock()
			}
			dataWG.Add(1)
			go a.getResourcesOfType(dataWG, curType, filter, handler)
		}
	}

	dataWG.Wait()

	return allResources, finalErr
}

func (a *AWS) getResourcesOfType(wg *sync.WaitGroup, resourceType resourceType, filter *types.InfraFilter, handler awsResourceHandler) {
	defer wg.Done()

	resources, resourceErr := resourceType.getResources(filter)
	handler(resourceType, resources, resourceErr)
}

func (a *AWS) UpdateResourceTag(resource *types.InfraResource, tagKey *string, tagValue *string) error {
	if resource.Provider != awsProviderName {
		return fmt.Errorf("Resource's provider is %s instead of EC2. Cannot update the tag", resource.Provider)
	}

	for _, curType := range a.resourceTypes {
		if curType.getName() == resource.ResourceType {
			return curType.updateResourceTag(resource, tagKey, tagValue)
		}
	}

	return fmt.Errorf("Cannot update tag for resource of type '%s'", resource.ResourceType)
}

func (a *AWS) UpdateResourceState(resource *types.InfraResource, safe bool, state string) error {
	if resource.Provider != awsProviderName {
		return fmt.Errorf("Resource's provider is %s instead of EC2. Cannot update the tag", resource.Provider)
	}

	for _, curType := range a.resourceTypes {
		if curType.getName() == resource.ResourceType {
			return curType.updateResourceState(resource, safe, state)
		}
	}

	return fmt.Errorf("Cannot update resource state for type '%s'", resource.ResourceType)
}

func addStringProperty(propName string, propValue *string, properties *map[string]string) {
	if propValue != nil {
		(*properties)[propName] = *propValue
	}
}

func addTimeProperty(propName string, propValue *time.Time, properties *map[string]string) {
	if propValue != nil {
		(*properties)[propName] = time.Time.String(*propValue)
	}
}

func considerRegion(region string, filter *types.InfraFilter) bool {
	if len(filter.Regions) == 0 {
		return true
	}

	for _, curRegion := range filter.Regions {
		if strings.ToLower(curRegion) == strings.ToLower(region) {
			return true
		}
	}

	return false
}

func considerResourceType(resourceType string, filter *types.InfraFilter) bool {
	if len(filter.ResourceTypes) == 0 {
		return true
	}

	for _, curType := range filter.ResourceTypes {
		if strings.ToLower(curType) == strings.ToLower(resourceType) {
			return true
		}
	}

	return false
}

func considerTags(tags map[string]string, filter *types.InfraFilter) bool {
	if len(filter.Tags) == 0 {
		return true
	}

	allOk := true
	for tagName, tagValue := range filter.Tags {
		if tagValueToCheck, ok := tags[tagName]; ok {
			if strings.ToLower(tagValue) != strings.ToLower(tagValueToCheck) {
				allOk = false
				break
			}
		} else {
			allOk = false
			break
		}
	}
	return allOk
}

func constructFilterExpression(filter *types.CostAndUsageFilter) *costexplorer.Expression {
	filters := []*costexplorer.Expression{}

	if len(filter.ResourceTypes) > 0 {
		resourceTypeExpression := &costexplorer.Expression{
			Dimensions: &costexplorer.DimensionValues{
				Key:    aws.String("SERVICE"),
				Values: aws.StringSlice(filter.ResourceTypes),
			},
		}
		filters = append(filters, resourceTypeExpression)
	}

	if len(filter.Regions) > 0 {
		regionExpression := &costexplorer.Expression{
			Dimensions: &costexplorer.DimensionValues{
				Key:    aws.String("REGION"),
				Values: aws.StringSlice(filter.Regions),
			},
		}
		filters = append(filters, regionExpression)
	}

	if len(filter.Tags) > 0 {
		tagsExpressions := []*costexplorer.Expression{}
		for tagName, tagValue := range filter.Tags {
			tagsExpressions = append(tagsExpressions, &costexplorer.Expression{
				Tags: &costexplorer.TagValues{
					Key:    aws.String(tagName),
					Values: aws.StringSlice([]string{tagValue}),
				},
			})
		}

		if len(tagsExpressions) == 1 {
			filters = append(filters, tagsExpressions[0])
		} else {
			filters = append(filters, &costexplorer.Expression{
				And: tagsExpressions,
			})
		}
	}

	filtersLen := len(filters)
	if filtersLen == 0 {
		return nil
	} else if filtersLen == 1 {
		return filters[0]
	} else {
		return &costexplorer.Expression{
			And: filters,
		}
	}
}

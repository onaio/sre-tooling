package aws

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
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

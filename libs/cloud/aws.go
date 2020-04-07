package cloud

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWS struct {
	dataMutex sync.Mutex
}
type awsResourceHandler func(region *string, resources []*Resource, err error)

const awsProviderName string = "AWS"
const resourceTypeEc2 string = "EC2"

func (aws *AWS) getName() string {
	return awsProviderName
}

func (a *AWS) getAllResources(filter *Filter, quiet bool) ([]*Resource, error) {
	allResources := []*Resource{}

	regions, regionErr := a.getRegions()
	if regionErr != nil {
		return allResources, regionErr
	}

	var finalErr error
	dataWG := new(sync.WaitGroup)
	for _, curRegion := range regions {
		if considerRegion(curRegion, filter) {
			if considerResourceType(resourceTypeEc2, filter) {
				handler := func(region *string, resources []*Resource, err error) {
					a.dataMutex.Lock()
					allResources = append(allResources, resources...)
					a.dataMutex.Unlock()
				}
				dataWG.Add(1)
				go a.getEC2InstancesInRegion(dataWG, curRegion, filter, quiet, handler)
			}
		}
	}

	dataWG.Wait()

	return allResources, finalErr
}

func (a *AWS) getRegions() ([]string, error) {
	regions := []string{}
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2Service := ec2.New(session)
	awsRegions, regionErr := ec2Service.DescribeRegions(nil)
	for _, curRegion := range awsRegions.Regions {
		regions = append(regions, *curRegion.RegionName)
	}
	if regionErr != nil {
		return nil, regionErr
	}

	return regions, nil
}

func (a *AWS) getEC2InstancesInRegion(wg *sync.WaitGroup, region string, filter *Filter, quiet bool, handler awsResourceHandler) {
	defer wg.Done()

	virtualMachines := []*Resource{}
	var finalErr error

	if !quiet {
		fmt.Printf("Getting AWS EC2 instances in %s\n", region)
	}
	awsConfig := aws.Config{
		Region: &region}
	// Load session from shared config
	session := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2Service := ec2.New(session)
	ec2Instances, ec2InstancesErr := ec2Service.DescribeInstances(a.constructAWSDescribeInstancesInput(filter))

	if ec2InstancesErr != nil {
		finalErr = ec2InstancesErr
	}

	if finalErr == nil {
		for _, curReservation := range ec2Instances.Reservations {
			for _, curInstance := range curReservation.Instances {
				instanceTags := make(map[string]string)
				for _, curInstanceTag := range curInstance.Tags {
					instanceTags[*curInstanceTag.Key] = *curInstanceTag.Value
				}

				instanceProperties := make(map[string]string)
				a.addStringProperty("availability-zone", curInstance.Placement.AvailabilityZone, &instanceProperties)
				a.addStringProperty("id", curInstance.InstanceId, &instanceProperties)
				a.addTimeProperty("launch-time", curInstance.LaunchTime, &instanceProperties)
				a.addStringProperty("private-ip", curInstance.PrivateIpAddress, &instanceProperties)
				a.addStringProperty("private-dns-name", curInstance.PrivateDnsName, &instanceProperties)
				a.addStringProperty("public-ip", curInstance.PublicIpAddress, &instanceProperties)
				a.addStringProperty("public-dns-name", curInstance.PublicDnsName, &instanceProperties)
				a.addStringProperty("image-id", curInstance.ImageId, &instanceProperties)
				a.addStringProperty("vpc-id", curInstance.VpcId, &instanceProperties)
				a.addStringProperty("instance-type", curInstance.InstanceType, &instanceProperties)
				a.addStringProperty("key-name", curInstance.KeyName, &instanceProperties)
				a.addStringProperty("state", curInstance.State.Name, &instanceProperties)
				a.addStringProperty("architecture", curInstance.Architecture, &instanceProperties)
				a.addStringProperty("platform", curInstance.Platform, &instanceProperties)

				resource := Resource{
					Provider:     awsProviderName,
					ID:           *curInstance.InstanceId,
					Location:     *curInstance.Placement.AvailabilityZone,
					ResourceType: resourceTypeEc2,
					LaunchTime:   *curInstance.LaunchTime,
					Properties:   instanceProperties,
					Tags:         instanceTags}
				virtualMachines = append(virtualMachines, &resource)
			}
		}
	}

	handler(&region, virtualMachines, finalErr)
}

func (a *AWS) constructAWSDescribeInstancesInput(filter *Filter) *ec2.DescribeInstancesInput {
	if len(filter.Tags) == 0 {
		return nil
	}

	var filters []*ec2.Filter
	for curTagKey, curTagValue := range filter.Tags {
		filterName := "tag:" + curTagKey
		filters = append(filters, &ec2.Filter{
			Name:   &filterName,
			Values: []*string{&curTagValue},
		})
	}

	return &ec2.DescribeInstancesInput{
		Filters: filters,
	}
}

func (a *AWS) addStringProperty(propName string, propValue *string, properties *map[string]string) {
	if propValue != nil {
		(*properties)[propName] = *propValue
	}
}

func (a *AWS) addTimeProperty(propName string, propValue *time.Time, properties *map[string]string) {
	if propValue != nil {
		(*properties)[propName] = time.Time.String(*propValue)
	}
}

func (a *AWS) updateResourceTag(region *string, resource *Resource, tagKey *string, tagValue *string) error {
	if resource.Provider != awsProviderName {
		return fmt.Errorf("Resource's provider is %s instead of AWS. Cannot update the tag", resource.Provider)
	}

	switch resource.ResourceType {
	case resourceTypeEc2:
		return a.updateEC2ResourceTag(region, resource, tagKey, tagValue)
	default:
		return fmt.Errorf("Unknown resource type %s. Cannot update the tag", resource.ResourceType)
	}
}

func (a *AWS) updateEC2ResourceTag(region *string, resource *Resource, tagKey *string, tagValue *string) error {
	if len(resource.ID) == 0 {
		return fmt.Errorf("Could not update the EC2 instance tag because the instance's ID is not set")
	}
	if len(*tagKey) == 0 {
		return fmt.Errorf("Could not update the EC2 instance tag because the tag key is not set")
	}

	awsConfig := aws.Config{
		Region: region}
	// Load session from shared config
	session := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2Service := ec2.New(session)

	tag := ec2.Tag{Key: tagKey, Value: tagValue}
	_, creatTagErr := ec2Service.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{&resource.ID},
		Tags:      []*ec2.Tag{&tag},
	})

	return creatTagErr
}

func (a *AWS) updateResourceState(resource *Resource, safe bool, state string) error {
	if resource.Provider != awsProviderName {
		return fmt.Errorf("Resource's provider is %s instead of AWS. Cannot update the tag", resource.Provider)
	}

	switch resource.ResourceType {
	case resourceTypeEc2:
		return a.updateEC2ResourceState(resource, safe, state)
	default:
		return fmt.Errorf("Unknown resource type %s. Cannot update the tag", resource.ResourceType)
	}
}

func (a *AWS) updateEC2ResourceState(resource *Resource, safe bool, state string) error {

	return nil
}

package cloud

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWS struct {
	session *session.Session
}

const awsProviderName string = "AWS"
const resourceTypeEc2 string = "EC2"

func (a *AWS) init() {
	awsConfig := aws.Config{}
	// Load session from shared config
	a.session = session.Must(session.NewSessionWithOptions(session.Options{
		Config:            awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func (aws *AWS) getName() string {
	return awsProviderName
}

func (a *AWS) getAllResources(filter *Filter, quiet bool) ([]*Resource, error) {
	resources := []*Resource{}
	regions, regionErr := a.getRegions()
	if regionErr != nil {
		return nil, regionErr
	}

	for _, curRegion := range regions {
		if considerRegion(curRegion, filter) {
			if considerResourceType(resourceTypeEc2, filter) {
				curVirtMachines, curRegErr := a.getEC2InstancesInRegion(&curRegion, filter, quiet)
				if curRegErr != nil {
					return nil, curRegErr
				}

				resources = append(resources, curVirtMachines...)
			}
		}
	}

	return resources, nil
}

func (a *AWS) getRegions() ([]string, error) {
	regions := []string{}
	ec2Service := ec2.New(a.session)
	awsRegions, regionErr := ec2Service.DescribeRegions(nil)
	for _, curRegion := range awsRegions.Regions {
		regions = append(regions, *curRegion.RegionName)
	}
	if regionErr != nil {
		return nil, regionErr
	}

	return regions, nil
}

func (a *AWS) getEC2InstancesInRegion(region *string, filter *Filter, quiet bool) ([]*Resource, error) {
	if !quiet {
		fmt.Printf("Getting AWS EC2 instances in %s\n", *region)
	}

	ec2Service := ec2.New(a.session)
	ec2Instances, ec2InstancesErr := ec2Service.DescribeInstances(nil)

	if ec2InstancesErr != nil {
		return nil, ec2InstancesErr
	}

	virtualMachines := []*Resource{}

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

			if considerTags(instanceTags, filter) {
				resource := Resource{
					Provider:     awsProviderName,
					ID:           *curInstance.InstanceId,
					Location:     *curInstance.Placement.AvailabilityZone,
					ResourceType: resourceTypeEc2,
					LaunchTime:   *curInstance.LaunchTime,
					Region:       *region,
					Properties:   instanceProperties,
					Tags:         instanceTags}
				virtualMachines = append(virtualMachines, &resource)
			}
		}
	}

	return virtualMachines, nil
}

func (a *AWS) updateResourceTag(resource *Resource, tagKey *string, tagValue *string) error {
	if resource.Provider == awsProviderName {
		switch resource.ResourceType {
		case resourceTypeEc2:
			a.updateEC2ResourceTag(resource, tagKey, tagValue)
		default:
			return errors.New("Unknown resource type " + resource.ResourceType)
		}
	}

	return nil
}

func (a *AWS) updateEC2ResourceTag(resource *Resource, tagKey *string, tagValue *string) error {
	ec2Service := ec2.New(a.session)
	tag := ec2.Tag{Key: tagKey, Value: tagValue}
	input := &ec2.CreateTagsInput{
		Tags: []*ec2.Tag{&tag},
	}
	_, createTagsErr := ec2Service.CreateTags(input)

	return createTagsErr
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

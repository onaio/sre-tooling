package cloud

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWS struct {
}

const awsProviderName string = "AWS"
const resourceTypeEc2 string = "EC2"

func (aws AWS) getName() string {
	return awsProviderName
}

func (a AWS) getAllResources(filter *Filter, quiet bool) ([]*Resource, error) {
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

func (a AWS) getRegions() ([]string, error) {
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

func (a AWS) getEC2InstancesInRegion(region *string, filter *Filter, quiet bool) ([]*Resource, error) {
	if !quiet {
		fmt.Printf("Getting AWS EC2 instances in %s\n", *region)
	}
	awsConfig := aws.Config{
		Region: region}
	// Load session from shared config
	session := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2Service := ec2.New(session)
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
			instanceProperties["availability-zone"] = *curInstance.Placement.AvailabilityZone
			instanceProperties["id"] = *curInstance.InstanceId
			instanceProperties["launch-time"] = time.Time.String(*curInstance.LaunchTime)
			instanceProperties["private-ip"] = *curInstance.PrivateIpAddress
			instanceProperties["private-dns-name"] = *curInstance.PrivateDnsName
			if curInstance.PublicIpAddress != nil {
				instanceProperties["public-ip"] = *curInstance.PublicIpAddress
			}
			instanceProperties["public-dns-name"] = *curInstance.PublicDnsName
			instanceProperties["image-id"] = *curInstance.ImageId
			instanceProperties["vpc-id"] = *curInstance.VpcId
			instanceProperties["instance-type"] = *curInstance.InstanceType
			instanceProperties["key-name"] = *curInstance.KeyName
			instanceProperties["state"] = *curInstance.State.Name
			instanceProperties["architecture"] = *curInstance.Architecture
			if curInstance.Platform != nil {
				instanceProperties["platform"] = *curInstance.Platform
			}

			if considerTags(instanceTags, filter) {
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

	return virtualMachines, nil
}

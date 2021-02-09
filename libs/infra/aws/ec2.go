package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/onaio/sre-tooling/libs/types"
)

type EC2 struct {
	dataMutex sync.Mutex
	session   *session.Session
}
type ec2ResourceHandler func(region *string, resources []*types.InfraResource, err error)

const resourceTypeEc2 string = "EC2"

func (e *EC2) init(session *session.Session) error {
	e.session = session

	return nil
}

func (e *EC2) getResources(filter *types.InfraFilter) ([]*types.InfraResource, error) {
	allResources := []*types.InfraResource{}

	regions, regionErr := e.getRegions()
	if regionErr != nil {
		return allResources, regionErr
	}

	var finalErr error
	dataWG := new(sync.WaitGroup)
	for _, curRegion := range regions {
		if considerRegion(curRegion, filter) {
			handler := func(region *string, resources []*types.InfraResource, err error) {
				e.dataMutex.Lock()
				allResources = append(allResources, resources...)
				if err != nil {
					finalErr = err
				}
				e.dataMutex.Unlock()
			}
			dataWG.Add(1)
			go e.getEC2InstancesInRegion(dataWG, curRegion, filter, handler)
		}
	}

	dataWG.Wait()

	return allResources, finalErr
}

func (e *EC2) getName() string {
	return resourceTypeEc2
}

func (e *EC2) getRegions() ([]string, error) {
	regions := []string{}

	ec2Service := ec2.New(e.session)
	awsRegions, regionErr := ec2Service.DescribeRegions(nil)
	for _, curRegion := range awsRegions.Regions {
		regions = append(regions, *curRegion.RegionName)
	}
	if regionErr != nil {
		return nil, regionErr
	}

	return regions, nil
}

func (e *EC2) getEC2InstancesInRegion(wg *sync.WaitGroup, region string, filter *types.InfraFilter, handler ec2ResourceHandler) {
	defer wg.Done()

	virtualMachines := []*types.InfraResource{}
	var finalErr error

	// Don't use the shared session since you will be updating the region in the session
	awsConfig := aws.Config{
		Region: &region}
	// Load session from shared config
	session := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2Service := ec2.New(session)
	ec2Instances, ec2InstancesErr := ec2Service.DescribeInstances(
		e.constructEC2DescribeInstancesInput(filter),
	)

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
				addStringProperty("availability-zone", curInstance.Placement.AvailabilityZone, &instanceProperties)
				addStringProperty("id", curInstance.InstanceId, &instanceProperties)
				addTimeProperty("launch-time", curInstance.LaunchTime, &instanceProperties)
				addStringProperty("private-ip", curInstance.PrivateIpAddress, &instanceProperties)
				addStringProperty("private-dns-name", curInstance.PrivateDnsName, &instanceProperties)
				addStringProperty("public-ip", curInstance.PublicIpAddress, &instanceProperties)
				addStringProperty("public-dns-name", curInstance.PublicDnsName, &instanceProperties)
				addStringProperty("image-id", curInstance.ImageId, &instanceProperties)
				addStringProperty("vpc-id", curInstance.VpcId, &instanceProperties)
				addStringProperty("instance-type", curInstance.InstanceType, &instanceProperties)
				addStringProperty("key-name", curInstance.KeyName, &instanceProperties)
				addStringProperty("state", curInstance.State.Name, &instanceProperties)
				addStringProperty("architecture", curInstance.Architecture, &instanceProperties)
				addStringProperty("platform", curInstance.Platform, &instanceProperties)

				resource := types.InfraResource{
					Provider:     awsProviderName,
					ID:           *curInstance.InstanceId,
					Location:     region,
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

func (e *EC2) constructEC2DescribeInstancesInput(filter *types.InfraFilter) *ec2.DescribeInstancesInput {
	if len(filter.Tags) == 0 {
		return nil
	}

	var filters []*ec2.Filter
	for curTagKey, curTagValue := range filter.Tags {
		filterName := "tag:" + curTagKey
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(filterName),
			Values: aws.StringSlice([]string{curTagValue}),
		})
	}

	return &ec2.DescribeInstancesInput{
		Filters: filters,
	}
}

func (e *EC2) updateResourceTag(resource *types.InfraResource, tagKey *string, tagValue *string) error {
	if len(resource.ID) == 0 {
		return fmt.Errorf("Could not update the EC2 instance tag because the instance's ID is not set")
	}
	if len(*tagKey) == 0 {
		return fmt.Errorf("Could not update the EC2 instance tag because the tag key is not set")
	}

	awsConfig := aws.Config{
		Region: &resource.Location}
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

func (e *EC2) updateResourceState(resource *types.InfraResource, safe bool, state string) error {
	return fmt.Errorf("Not implemented yet")
}

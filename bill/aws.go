package bill

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getAwsVirtualMachines() ([]*VirtualMachine, error) {
	virtualMachines := []*VirtualMachine{}
	regions, regionErr := getAwsRegions()
	if regionErr != nil {
		return nil, regionErr
	}

	for _, curRegion := range regions {
		curVirtMachines, curRegErr := getAwsVirtualMachinesInRegion(&curRegion)
		if curRegErr != nil {
			return nil, curRegErr
		}

		virtualMachines = append(virtualMachines, curVirtMachines...)
	}

	return virtualMachines, nil
}

func getAwsRegions() ([]string, error) {
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

func getAwsVirtualMachinesInRegion(region *string) ([]*VirtualMachine, error) {
	fmt.Printf("Getting AWS EC2 instances in %s\n", *region)
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

	virtualMachines := []*VirtualMachine{}

	for _, curReservation := range ec2Instances.Reservations {
		for _, curInstance := range curReservation.Instances {
			instanceTags := make(map[string]string)
			for _, curInstanceTag := range curInstance.Tags {
				instanceTags[*curInstanceTag.Key] = *curInstanceTag.Value
			}

			curVirtualMachine := VirtualMachine{
				provider:     "aws",
				id:           *curInstance.InstanceId,
				location:     *curInstance.Placement.AvailabilityZone,
				architecture: *curInstance.Architecture,
				launchTime:   *curInstance.LaunchTime,
				tags:         instanceTags}
			virtualMachines = append(virtualMachines, &curVirtualMachine)
		}
	}

	return virtualMachines, nil
}

package billing

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getAwsVirtualMachines() ([]*VirtualMachine, error) {
	// Load session from shared config
	session := session.Must(session.NewSessionWithOptions(session.Options{
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

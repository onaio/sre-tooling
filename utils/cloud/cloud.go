package cloud

import (
	"reflect"
	"time"
)

type VirtualMachine struct {
	Provider     string
	ID           string
	Location     string
	Architecture string
	LaunchTime   time.Time
	Tags         map[string]string
}

func GetAllVirtualMachines() ([]*VirtualMachine, error) {
	allVirtualMachines, awsErr := getAwsVirtualMachines()
	if awsErr != nil {
		return nil, awsErr
	}

	return allVirtualMachines, nil
}

func GetTagKeys(virtualMachine *VirtualMachine) []string {
	keyObjects := reflect.ValueOf(virtualMachine.Tags).MapKeys()
	keys := make([]string, len(keyObjects))
	for i := 0; i < len(keyObjects); i++ {
		keys[i] = keyObjects[i].String()
	}

	return keys
}

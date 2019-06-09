package bill

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/onaio/sre-tooling/utils/cloud"
)

const validateSubCommand = "validate"

func validateTags() (bool, error) {
	requiredTagsString := os.Getenv(requiredTagsEnvVar)
	if len(requiredTagsString) == 0 {
		return false, errors.New(fmt.Sprintf("%s not set", requiredTagsEnvVar))
	}
	requiredTags := strings.Split(requiredTagsString, ",")

	allVirtualMachines, virtMachinesErr := cloud.GetAllVirtualMachines()
	if virtMachinesErr != nil {
		return false, virtMachinesErr
	}

	fmt.Printf("Checking %d instances\n", len(allVirtualMachines))
	allGood := true
	errMessage := ""
	for _, curVirtualMachine := range allVirtualMachines {
		curTagKeys := cloud.GetTagKeys(curVirtualMachine)
		missingTags := getItemsInANotB(&requiredTags, &curTagKeys)
		if len(missingTags) > 0 {
			allGood = false
			errMessage = errMessage + fmt.Sprintf("%s - %s - %s missing tags %v\n", curVirtualMachine.Provider, curVirtualMachine.Location, curVirtualMachine.ID, missingTags)
		}
	}

	return allGood, errors.New(errMessage)
}

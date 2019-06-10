package bill

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/onaio/sre-tooling/libs/cloud"
)

const validateSubCommand = "validate"

func validateTags() (bool, error) {
	requiredTagsString := os.Getenv(requiredTagsEnvVar)
	if len(requiredTagsString) == 0 {
		return false, errors.New(fmt.Sprintf("%s not set", requiredTagsEnvVar))
	}
	requiredTags := strings.Split(requiredTagsString, ",")

	allResources, resourcesErr := cloud.GetAllCloudResources(nil)
	if resourcesErr != nil {
		return false, resourcesErr
	}

	fmt.Printf("Checking %d instances\n", len(allResources))
	allGood := true
	errMessage := ""
	for _, curResource := range allResources {
		curTagKeys := cloud.GetTagKeys(curResource)
		missingTags := getItemsInANotB(&requiredTags, &curTagKeys)
		if len(missingTags) > 0 {
			allGood = false
			errMessage = errMessage + fmt.Sprintf("%s - %s - %s - %s missing tags %v\n", curResource.Provider, curResource.ResourceType, curResource.Location, curResource.ID, missingTags)
		}
	}

	return allGood, errors.New(errMessage)
}

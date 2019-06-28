package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/onaio/sre-tooling/libs/cloud"
	"github.com/onaio/sre-tooling/libs/notification"
)

const Command string = "validate"
const RequiredTagsEnvVar = "SRE_INFRA_BILL_REQUIRED_TAGS"

func Validate() {
	requiredTagsString := os.Getenv(RequiredTagsEnvVar)
	if len(requiredTagsString) == 0 {
		notification.SendMessage(fmt.Sprintf("%s not set", RequiredTagsEnvVar))
		os.Exit(1)
	}
	requiredTags := strings.Split(requiredTagsString, ",")

	allResources, resourcesErr := cloud.GetAllCloudResources(nil)
	if resourcesErr != nil {
		notification.SendMessage(resourcesErr.Error())
		os.Exit(1)
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

	if !allGood {
		notification.SendMessage(errMessage)
		os.Exit(1)
	}
}

func getItemsInANotB(a *[]string, b *[]string) []string {
	itemsInANotB := []string{}

	for _, curAItem := range *a {
		if !stringInSlice(curAItem, b) {
			itemsInANotB = append(itemsInANotB, curAItem)
		}
	}

	return itemsInANotB
}

func stringInSlice(x string, slice *[]string) bool {
	for _, curItem := range *slice {
		if curItem == x {
			return true
		}
	}

	return false
}

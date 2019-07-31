package validate

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/cloud"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "validate"
const requiredTagsEnvVar = "SRE_INFRA_BILL_REQUIRED_TAGS"

type Validate struct {
	helpFlag     *bool
	flagSet      *flag.FlagSet
	providerFlag *flags.StringArray
	regionFlag   *flags.StringArray
	typeFlag     *flags.StringArray
	tagFlag      *flags.StringArray
}

func (validate *Validate) Init(helpFlagName string, helpFlagDescription string) {
	validate.flagSet = flag.NewFlagSet(validate.GetName(), flag.ExitOnError)
	validate.helpFlag = validate.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	validate.providerFlag, validate.regionFlag, validate.typeFlag, validate.tagFlag = cloud.AddFilterFlags(validate.flagSet)
}

func (validate *Validate) GetName() string {
	return name
}

func (validate *Validate) GetDescription() string {
	return "Validates whether billing tags for resources are okay"
}

func (validate *Validate) ParseArgs(args []string) {
	validate.flagSet.Parse(args)
	if *validate.helpFlag {
		validate.printHelp()
	} else {
		validate.validate()
	}
}

func (validate *Validate) validate() {
	requiredTagsString := os.Getenv(requiredTagsEnvVar)
	if len(requiredTagsString) == 0 {
		notification.SendMessage(fmt.Sprintf("%s not set", requiredTagsEnvVar))
		os.Exit(1)
	}
	requiredTags := strings.Split(requiredTagsString, ",")

	allResources, resourcesErr := cloud.GetAllCloudResources(cloud.GetFiltersFromCommandFlags(validate.providerFlag, validate.regionFlag, validate.typeFlag, validate.tagFlag), false)
	if resourcesErr != nil {
		notification.SendMessage(resourcesErr.Error())
		os.Exit(1)
	}

	fmt.Printf("Checking %d resources\n", len(allResources))
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

func (validate *Validate) printHelp() {
	fmt.Println(validate.GetDescription())
	validate.flagSet.PrintDefaults()
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

package calculate

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/infra"
	"github.com/onaio/sre-tooling/libs/notification"
	"github.com/onaio/sre-tooling/libs/numbers"
)

const name string = "calculate"

// Calculate determines whether a resource needs a new index. Exit codes are:
// 	0 - If the resource needs a new index (proposed index is returned)
//  1 - If a command execution error occurs
//  2 - If a command processing error occurs (e.g if an unavailable flag is provided by the user)
//  3 - If the resource doesn't need a new index (whatever index it has right now is just fine)
type Calculate struct {
	helpFlag        *bool
	flagSet         *flag.FlagSet
	providerFlag    *flags.StringArray
	regionFlag      *flags.StringArray
	typeFlag        *flags.StringArray
	tagFlag         *flags.StringArray
	idFlag          *string
	indexTagFlag    *string
	randomSleepFlag *int
	subCommands     []cli.Command
}

// Init initializes the command object
func (calculate *Calculate) Init(helpFlagName string, helpFlagDescription string) {
	calculate.flagSet = flag.NewFlagSet(calculate.GetName(), flag.ExitOnError)
	calculate.helpFlag = calculate.flagSet.Bool(helpFlagName, false, helpFlagDescription)

	calculate.providerFlag,
		calculate.regionFlag,
		calculate.typeFlag,
		calculate.tagFlag,
		calculate.idFlag,
		calculate.indexTagFlag,
		calculate.randomSleepFlag = AddCalculateFlags(calculate.flagSet)

	calculate.subCommands = []cli.Command{}
}

// AddCalculateFlags returns the flags required to calculate a resource's index in the order:
//    id: The ID of the resource
//    index tag: The index tag to filter the resource group using
//    random sleep: The maximum random number of seconds to sleep before calculating the index
func AddCalculateFlags(flagSet *flag.FlagSet) (*flags.StringArray, *flags.StringArray, *flags.StringArray, *flags.StringArray, *string, *string, *int) {
	providerFlag,
		regionFlag,
		typeFlag,
		tagFlag := infra.AddFilterFlags(flagSet)

	idFlag := flagSet.String("id", "", "The ID of the resource to check the index")
	indexTagFlag := flagSet.String("index-tag", "", "The name of the tag containing the indexes of the resources")
	randomSleepFlag := flagSet.Int("random-sleep", 0, "Sleep for a random number of seconds between 0 and what is defined before trying to calculate")

	return providerFlag,
		regionFlag,
		typeFlag,
		tagFlag,
		idFlag,
		indexTagFlag,
		randomSleepFlag
}

// GetName returns the value of the name constant
func (calculate *Calculate) GetName() string {
	return name
}

// GetDescription returns the description for the calculate command
func (calculate *Calculate) GetDescription() string {
	return "Calculates the index of an infrastructure resource in the group filtered by the provided filter flags"
}

// GetFlagSet returns a pointer to the flag.FlagSet associated to the command
func (calculate *Calculate) GetFlagSet() *flag.FlagSet {
	return calculate.flagSet
}

// GetSubCommands returns a slice of subcommands under the calculate command
// (expect empty slice if none)
func (calculate *Calculate) GetSubCommands() []cli.Command {
	return calculate.subCommands
}

// GetHelpFlag returns a pointer to the initialized help flag for the command
func (calculate *Calculate) GetHelpFlag() *bool {
	return calculate.helpFlag
}

// Process calculates whether a resource needs to be assigned a new index and
// sends the index to configured notification channels or exit with an exit code
// 3 if the resource doesn't need to be assigned a new index
func (calculate *Calculate) Process() {
	newIndex, newIndexErr := FetchAndCalculateResourceIndex(
		calculate.randomSleepFlag,
		calculate.providerFlag,
		calculate.regionFlag,
		calculate.typeFlag,
		calculate.tagFlag,
		calculate.idFlag,
		calculate.indexTagFlag,
	)

	if newIndexErr != nil {
		notification.SendMessage(newIndexErr.Error())
		cli.ExitCommandExecutionError()
	}

	notification.SendMessage(strconv.Itoa(newIndex))
}

func FetchAndCalculateResourceIndex(
	randomSleepFlag *int,
	providerFlag *flags.StringArray,
	regionFlag *flags.StringArray,
	typeFlag *flags.StringArray,
	tagFlag *flags.StringArray,
	idFlag *string,
	indexTagFlag *string) (int, error) {
	if len(*idFlag) == 0 {
		return -1, fmt.Errorf("You need to provide the ID of the resource you want to check its index")
	}
	if len(*indexTagFlag) == 0 {
		return -1, fmt.Errorf("You need to provide the name of the tag containing resource indexes")
	}

	if len(*regionFlag) == 0 &&
		len(*typeFlag) == 0 &&
		len(*tagFlag) == 0 {
		return -1, fmt.Errorf("You need to filter resources using at least one region, type, or tag")
	}

	// Sleep for some random amount of time
	sleepTime := numbers.GetRandomInt(*randomSleepFlag)
	if sleepTime > 0 {
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	allResources, resourcesErr := infra.GetAllCloudResources(
		infra.GetFiltersFromCommandFlags(
			providerFlag,
			regionFlag,
			typeFlag,
			tagFlag),
		true)
	if resourcesErr != nil {
		return -1, resourcesErr
	}

	// Calculate the new index
	newIndex, newIndexErr := GetNewResourceIndex(
		idFlag,
		indexTagFlag,
		allResources)
	if newIndexErr != nil {
		return -1, newIndexErr
	}

	return newIndex, nil
}

// GetNewResourceIndex calculates the new index for the resource with the ID specified in resourceID.
// An error will be returned if:
// 	- No resource with the ID specified in resourceID is found in resourceMap
// 	- The new index is not different from the current index of the resource
func GetNewResourceIndex(
	resourceID *string,
	indexTag *string,
	resources []*infra.Resource) (int, error) {

	// Map with the resource index as the key and the number of resources tagged with the index as the value
	indexMap := make(map[int]int)
	resourceMap := make(map[string]*infra.Resource)
	largestIndex := 0
	for _, curResource := range resources {
		resourceMap[curResource.ID] = curResource

		// Get the resource's index from the resource object
		curResourceIndex, indexErr := getResourceIndex(curResource, indexTag)
		if indexErr != nil {
			return 0, indexErr
		}

		// Add index to the index map
		if indexCount, indexCountSet := indexMap[curResourceIndex]; indexCountSet {
			indexMap[curResourceIndex] = indexCount + 1
		} else {
			indexMap[curResourceIndex] = 1
		}

		// Check if the resource's index is the largest index
		if curResourceIndex > largestIndex {
			largestIndex = curResourceIndex
		}
	}

	resource, resourceExists := resourceMap[*resourceID]
	// Check if we have a resource in the group with the ID stored in calculate.idFlag
	if !resourceExists {
		return 0, fmt.Errorf("Resource with the ID %s was not found in the resource group", *resourceID)
	}

	// Check if there is just one resource with the resource's index
	// Don't check if error emitted since that would have already been caught before
	resourceIndex, _ := getResourceIndex(resource, indexTag)
	if indexMap[resourceIndex] == 1 {
		return resourceIndex, nil
	}

	// Try calculate a new index for the resource
	// Find an index between 0 and largestIndex that isn't set
	for curIndex := 0; curIndex <= largestIndex; curIndex++ {
		if _, indexExists := indexMap[curIndex]; !indexExists {
			return curIndex, nil
		}
	}

	// Suggest index to be largestIndex + 1
	return largestIndex + 1, nil
}

// getResourceIndex returns the tagged value of a resource's index or the default index (0)
// if the resource has not been tagged with an index
func getResourceIndex(resource *infra.Resource, indexTag *string) (int, error) {
	indexString, indexTagSet := resource.Tags[*indexTag]
	resourceIndex := 0
	if indexTagSet && len(indexString) > 0 {
		index, indexErr := strconv.Atoi(indexString)

		if indexErr != nil {
			return resourceIndex, indexErr
		}
		resourceIndex = index
	}

	return resourceIndex, nil
}

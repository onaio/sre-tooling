package bill

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/onaio/sre-tooling/notify"
)

type VirtualMachine struct {
	provider     string
	id           string
	location     string
	architecture string
	launchTime   time.Time
	tags         map[string]string
}

const Command string = "bill"
const validateSubCommand = "validate"
const requiredTagsEnvVar = "SRE_BILLING_REQUIRED_TAGS"

func Cli(commandName string, helpSubCommand string, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case validateSubCommand:
			tagsValid, tagsErr := validateTags()
			if !tagsValid {
				notify.SendMessage(tagsErr.Error())
				os.Exit(1)
			}
		case helpSubCommand:
			fmt.Println(help(commandName, helpSubCommand))
		default:
			fmt.Println(help(commandName, helpSubCommand))
			os.Exit(1)
		}
		os.Exit(0)
	} else {
		fmt.Println(help(commandName, helpSubCommand))
		os.Exit(1)
	}
}

func help(commandName string, helpSubCommand string) string {
	text := `
Usage: %s %s [command]

Common commands:
	%s	Validates whether billing tags for resources are okay
	%s		Prints this help message
`
	return fmt.Sprintf(text, commandName, Command, validateSubCommand, helpSubCommand)
}

func getAllVirtualMachines() ([]*VirtualMachine, error) {
	allVirtualMachines, awsErr := getAwsVirtualMachines()
	if awsErr != nil {
		return nil, awsErr
	}

	return allVirtualMachines, nil
}

func validateTags() (bool, error) {
	requiredTagsString := os.Getenv(requiredTagsEnvVar)
	if len(requiredTagsString) == 0 {
		return false, errors.New(fmt.Sprintf("%s not set", requiredTagsEnvVar))
	}
	requiredTags := strings.Split(requiredTagsString, ",")

	allVirtualMachines, virtMachinesErr := getAllVirtualMachines()
	if virtMachinesErr != nil {
		return false, virtMachinesErr
	}

	fmt.Printf("Checking %d instances\n", len(allVirtualMachines))
	allGood := true
	errMessage := ""
	for _, curVirtualMachine := range allVirtualMachines {
		curTagKeys := getTagKeys(curVirtualMachine)
		missingTags := getItemsInANotB(&requiredTags, &curTagKeys)
		if len(missingTags) > 0 {
			allGood = false
			errMessage = errMessage + fmt.Sprintf("%s - %s - %s missing tags %v\n", curVirtualMachine.provider, curVirtualMachine.location, curVirtualMachine.id, missingTags)
		}
	}

	return allGood, errors.New(errMessage)
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

func getTagKeys(virtualMachine *VirtualMachine) []string {
	keyObjects := reflect.ValueOf(virtualMachine.tags).MapKeys()
	keys := make([]string, len(keyObjects))
	for i := 0; i < len(keyObjects); i++ {
		keys[i] = keyObjects[i].String()
	}

	return keys
}

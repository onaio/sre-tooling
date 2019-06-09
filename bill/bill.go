package bill

import (
	"fmt"
	"os"

	"github.com/onaio/sre-tooling/notify"
)

const Command string = "bill"
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

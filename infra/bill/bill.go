package bill

import (
	"fmt"
	"os"

	"github.com/onaio/sre-tooling/infra/bill/validate"
)

const Command string = "bill"

func Cli(commandName string, infraSubCommand string, helpSubCommand string, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case validate.Command:
			validate.Validate()
		case helpSubCommand:
			fmt.Println(help(commandName, infraSubCommand, helpSubCommand))
		default:
			fmt.Println(help(commandName, infraSubCommand, helpSubCommand))
			os.Exit(1)
		}
		os.Exit(0)
	} else {
		fmt.Println(help(commandName, infraSubCommand, helpSubCommand))
		os.Exit(1)
	}
}

func help(commandName string, infraSubCommand string, helpSubCommand string) string {
	text := `
Usage: %s %s %s [command]

Common commands:
	%s	Validates whether billing tags for resources are okay
	%s		Prints this help message
`
	return fmt.Sprintf(text, commandName, infraSubCommand, Command, validate.Command, helpSubCommand)
}

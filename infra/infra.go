package infra

import (
	"fmt"
	"os"

	"github.com/onaio/sre-tooling/infra/bill"
)

const Command string = "infra"

func Cli(commandName string, helpSubCommand string, args []string) {
	if len(args) > 0 {
		subCommandArgs := []string{}
		if len(args) > 1 {
			subCommandArgs = args[1:]
		}

		switch args[0] {
		case bill.Command:
			bill.Cli(commandName, Command, helpSubCommand, subCommandArgs)
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
	%s		Infrastructure billing specific commands
	%s		Prints this help message
`
	return fmt.Sprintf(text, commandName, Command, bill.Command, helpSubCommand)
}

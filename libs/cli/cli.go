package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Command is the interface to implement if you want to define a (sub)command
type Command interface {
	// Init initializes the command. Place all initialization code here. Function should be
	// called when initializing the command hierarch (if sub-command, within the parent command's
	// Init function).
	Init(helpFlagName string, helpFlagDescription string)
	// GetName returns the name of the command. This is what will be displayed when a user requests to see
	// the command's manual.
	GetName() string
	// GetDescription returns the description of the command. This is what will be displayed when a user requests to
	// see the command's manual.
	GetDescription() string
	// GetFlagSet returns flag.FlagSet holding arguments linked to the command. If command's arguments linked to
	// the default FlagSet, then return nil.
	GetFlagSet() *flag.FlagSet
	// GetSubCommands returns a list sub-commands that are children of this command.
	GetSubCommands() []Command
	// GetHelpFlag returns a pointer to the bool linked to the command's help flag.
	GetHelpFlag() *bool
	// Process executes the logic for the command. If command has sub-commands, this function will never be called.
	Process()
}

// ParseArgs recursively parses arguments through the command hierarchy until it gets
// to the subcommand that needs to be executed
func ParseArgs(command Command, args []string) {
	flagSet := command.GetFlagSet()
	subCommands := command.GetSubCommands()
	if flagSet != nil && !flagSet.Parsed() {
		flagSet.Parse(args)
	}

	if *command.GetHelpFlag() {
		PrintHelp(command, false)
		return
	}

	if len(args) == 0 && len(subCommands) > 0 {
		PrintHelp(command, true)
	} else if len(args) > 0 && len(subCommands) > 0 {
		subCommandFound := false
		for _, curSubCommand := range subCommands {
			if curSubCommand.GetName() == args[0] {
				ParseArgs(curSubCommand, args[1:])
				subCommandFound = true
				break
			}
		}

		if !subCommandFound {
			PrintHelp(command, true)
		}
	} else {
		command.Process()
	}
}

// PrintHelp prints the help message for the provided command. If exitWithError is set
// to true, then the command exits with a command interpretation error status code
func PrintHelp(command Command, exitWithError bool) {
	subCommandSnippet := ""
	if len(command.GetSubCommands()) > 0 {
		subCommandSnippet = " <command>"
	}

	fmt.Printf("%s\n\nUsage:\n\n\t%s%s [arguments]\n\n%sAvailable arguments:\n",
		command.GetDescription(),
		command.GetName(),
		subCommandSnippet,
		getSubCommandsHelpText(command))
	if command.GetFlagSet() != nil {
		command.GetFlagSet().PrintDefaults()
	} else {
		flag.PrintDefaults()
	}
	fmt.Printf("\n")

	if exitWithError {
		ExitCommandInterpretationError()
	}
}

// getSubCommandsHelpText returns a string of subcommand names and descriptions
// to be printed as part of the provided command's help text
func getSubCommandsHelpText(command Command) string {
	text := ""
	longestName := 0

	for _, curSubCommand := range command.GetSubCommands() {
		if len(curSubCommand.GetName()) > longestName {
			longestName = len(curSubCommand.GetName())
		}
	}

	for _, curSubCommand := range command.GetSubCommands() {
		text = text + fmt.Sprintf("\t%s%s\t%s\n",
			curSubCommand.GetName(),
			strings.Repeat(" ", longestName-len(curSubCommand.GetName())),
			curSubCommand.GetDescription())
	}

	if len(text) > 0 {
		text = fmt.Sprintf("Commands are:\n%s\n", text)
	}

	return text
}

// ExitCommandInterpretationError exits with the exit code to return when there was
// an error interpreting the command typed by the user. Should technically be returned
// if the help message will be printed and the user didn't explicitly request for the
// help message to be printed.
func ExitCommandInterpretationError() {
	os.Exit(2)
}

// ExitCommandExecutionError exits with the exit code to return when an error occurred when
// executing a command.
func ExitCommandExecutionError() {
	os.Exit(1)
}

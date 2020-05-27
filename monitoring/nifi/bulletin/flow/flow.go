package flow

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi/bulletin/flow/ingest"
)

const name string = "flow"

// Flow holds data for the NiFi flow sub-command
type Flow struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init initializes the flow struct curresponding to the flow sub-command
func (flow *Flow) Init(helpFlagName string, helpFlagDescription string) {
	flow.flagSet = flag.NewFlagSet(flow.GetName(), flag.ExitOnError)
	flow.helpFlag = flow.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	ingest := new(ingest.Ingest)
	ingest.Init(helpFlagName, helpFlagDescription)
	flow.subCommands = []cli.Command{ingest}
}

// GetName returns the name of the flow sub-command
func (flow *Flow) GetName() string {
	return name
}

// GetDescription returns the description for the flow sub-command
func (flow *Flow) GetDescription() string {
	return "NiFi flow bulletin specific commands"
}

// GetFlagSet returns a flag.FlagSet corresponding to the flow sub-command
func (flow *Flow) GetFlagSet() *flag.FlagSet {
	return flow.flagSet
}

// GetSubCommands returns a list of sub-commands under the flow sub-command
func (flow *Flow) GetSubCommands() []cli.Command {
	return flow.subCommands
}

// GetHelpFlag returns the value for the help flag in the flow sub-command.
// TRUE means the user provided the flag and wants to print the sub-command's
// help message
func (flow *Flow) GetHelpFlag() *bool {
	return flow.helpFlag
}

// Process is expected to be empty since this sub-command doesn't do any processing
func (flow *Flow) Process() {}

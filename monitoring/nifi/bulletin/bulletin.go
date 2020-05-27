package bulletin

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi/bulletin/flow"
)

const name string = "bulletin"

// Bulletin holds data for the NiFi bulletin sub-command
type Bulletin struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init initializes the bulletin struct curresponding to the bulletin sub-command
func (bulletin *Bulletin) Init(helpFlagName string, helpFlagDescription string) {
	bulletin.flagSet = flag.NewFlagSet(bulletin.GetName(), flag.ExitOnError)
	bulletin.helpFlag = bulletin.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	flow := new(flow.Flow)
	flow.Init(helpFlagName, helpFlagDescription)
	bulletin.subCommands = []cli.Command{flow}
}

// GetName returns the name of the bulletin sub-command
func (bulletin *Bulletin) GetName() string {
	return name
}

// GetDescription returns the description for the bulletin sub-command
func (bulletin *Bulletin) GetDescription() string {
	return "NiFi bulletin specific commands"
}

// GetFlagSet returns a flag.FlagSet corresponding to the bulletin sub-command
func (bulletin *Bulletin) GetFlagSet() *flag.FlagSet {
	return bulletin.flagSet
}

// GetSubCommands returns a list of sub-commands under the bulletin sub-command
func (bulletin *Bulletin) GetSubCommands() []cli.Command {
	return bulletin.subCommands
}

// GetHelpFlag returns the value for the help flag in the bulletin sub-command.
// TRUE means the user provided the flag and wants to print the sub-command's
// help message
func (bulletin *Bulletin) GetHelpFlag() *bool {
	return bulletin.helpFlag
}

// Process is expected to be empty since this sub-command doesn't do any processing
func (bulletin *Bulletin) Process() {}

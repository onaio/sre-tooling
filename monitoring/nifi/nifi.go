package nifi

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi/bulletin"
)

const name string = "nifi"

// NiFi holds data for the NiFi monitoring sub-command
type NiFi struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init initializes the nifi struct curresponding to the bulletin sub-command
func (nifi *NiFi) Init(helpFlagName string, helpFlagDescription string) {
	nifi.flagSet = flag.NewFlagSet(nifi.GetName(), flag.ExitOnError)
	nifi.helpFlag = nifi.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	bulletin := new(bulletin.Bulletin)
	bulletin.Init(helpFlagName, helpFlagDescription)
	nifi.subCommands = []cli.Command{bulletin}
}

// GetName returns the name of the nifi sub-command
func (nifi *NiFi) GetName() string {
	return name
}

// GetDescription returns the description for the nifi sub-command
func (nifi *NiFi) GetDescription() string {
	return "NiFi specific commands"
}

// GetFlagSet returns a flag.FlagSet corresponding to the nifi sub-command
func (nifi *NiFi) GetFlagSet() *flag.FlagSet {
	return nifi.flagSet
}

// GetSubCommands returns a list of sub-commands under the nifi sub-command
func (nifi *NiFi) GetSubCommands() []cli.Command {
	return nifi.subCommands
}

// GetHelpFlag returns the value for the help flag in the bulletin sub-command.
// TRUE means the user provided the flag and wants to print the sub-command's
// help message
func (nifi *NiFi) GetHelpFlag() *bool {
	return nifi.helpFlag
}

// Process is expected to be empty since this sub-command doesn't do any processing
func (nifi *NiFi) Process() {}

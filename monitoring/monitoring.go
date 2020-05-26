package monitoring

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi"
)

const name string = "monitoring"

// Monitoring holds data for the NiFi monitoring sub-command
type Monitoring struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init initializes the monitoring struct curresponding to the monitoring sub-command
func (monitoring *Monitoring) Init(helpFlagName string, helpFlagDescription string) {
	monitoring.flagSet = flag.NewFlagSet(monitoring.GetName(), flag.ExitOnError)
	monitoring.helpFlag = monitoring.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	nifi := new(nifi.NiFi)
	nifi.Init(helpFlagName, helpFlagDescription)
	monitoring.subCommands = []cli.Command{nifi}
}

// GetName returns the name of the monitoring sub-command
func (monitoring *Monitoring) GetName() string {
	return name
}

// GetDescription returns the description for the monitoring sub-command
func (monitoring *Monitoring) GetDescription() string {
	return "Monitoring specific commands"
}

// GetFlagSet returns a flag.FlagSet corresponding to the monitoring sub-command
func (monitoring *Monitoring) GetFlagSet() *flag.FlagSet {
	return monitoring.flagSet
}

// GetSubCommands returns a list of sub-commands under the monitoring sub-command
func (monitoring *Monitoring) GetSubCommands() []cli.Command {
	return monitoring.subCommands
}

// GetHelpFlag returns the value for the help flag in the monitoring sub-command.
// TRUE means the user provided the flag and wants to print the sub-command's
// help message
func (monitoring *Monitoring) GetHelpFlag() *bool {
	return monitoring.helpFlag
}

// Process is expected to be empty since this sub-command doesn't do any processing
func (monitoring *Monitoring) Process() {}

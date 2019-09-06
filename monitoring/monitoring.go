package monitoring

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi"
)

const name string = "monitoring"

// Monitoring does...
type Monitoring struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init does...
func (monitoring *Monitoring) Init(helpFlagName string, helpFlagDescription string) {
	monitoring.flagSet = flag.NewFlagSet(monitoring.GetName(), flag.ExitOnError)
	monitoring.helpFlag = monitoring.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	nifi := new(nifi.NiFi)
	nifi.Init(helpFlagName, helpFlagDescription)
	monitoring.subCommands = []cli.Command{nifi}
}

// GetName does...
func (monitoring *Monitoring) GetName() string {
	return name
}

// GetDescription does...
func (monitoring *Monitoring) GetDescription() string {
	return "Monitoring specific commands"
}

// GetFlagSet does..
func (monitoring *Monitoring) GetFlagSet() *flag.FlagSet {
	return monitoring.flagSet
}

// GetSubCommands does..
func (monitoring *Monitoring) GetSubCommands() []cli.Command {
	return monitoring.subCommands
}

// GetHelpFlag does..
func (monitoring *Monitoring) GetHelpFlag() *bool {
	return monitoring.helpFlag
}

// Process does...
func (monitoring *Monitoring) Process() {}

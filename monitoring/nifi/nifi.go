package nifi

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi/bulletin"
)

const name string = "nifi"

// NiFi does...
type NiFi struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init does...
func (nifi *NiFi) Init(helpFlagName string, helpFlagDescription string) {
	nifi.flagSet = flag.NewFlagSet(nifi.GetName(), flag.ExitOnError)
	nifi.helpFlag = nifi.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	bulletin := new(bulletin.Bulletin)
	bulletin.Init(helpFlagName, helpFlagDescription)
	nifi.subCommands = []cli.Command{bulletin}
}

// GetName does...
func (nifi *NiFi) GetName() string {
	return name
}

// GetDescription does...
func (nifi *NiFi) GetDescription() string {
	return "NiFi specific commands"
}

// GetFlagSet does..
func (nifi *NiFi) GetFlagSet() *flag.FlagSet {
	return nifi.flagSet
}

// GetSubCommands does..
func (nifi *NiFi) GetSubCommands() []cli.Command {
	return nifi.subCommands
}

// GetHelpFlag does..
func (nifi *NiFi) GetHelpFlag() *bool {
	return nifi.helpFlag
}

// Process does...
func (nifi *NiFi) Process() {}

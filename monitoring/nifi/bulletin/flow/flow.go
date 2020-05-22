package flow

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi/bulletin/flow/ingest"
)

const name string = "flow"

// Flow does...
type Flow struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init does...
func (flow *Flow) Init(helpFlagName string, helpFlagDescription string) {
	flow.flagSet = flag.NewFlagSet(flow.GetName(), flag.ExitOnError)
	flow.helpFlag = flow.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	ingest := new(ingest.Ingest)
	ingest.Init(helpFlagName, helpFlagDescription)
	flow.subCommands = []cli.Command{ingest}
}

// GetName does...
func (flow *Flow) GetName() string {
	return name
}

// GetDescription does...
func (flow *Flow) GetDescription() string {
	return "NiFi flow bulletin specific commands"
}

// GetFlagSet does..
func (flow *Flow) GetFlagSet() *flag.FlagSet {
	return flow.flagSet
}

// GetSubCommands does..
func (flow *Flow) GetSubCommands() []cli.Command {
	return flow.subCommands
}

// GetHelpFlag does..
func (flow *Flow) GetHelpFlag() *bool {
	return flow.helpFlag
}

// Process does...
func (flow *Flow) Process() {}

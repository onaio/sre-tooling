package index

import (
	"flag"

	"github.com/onaio/sre-tooling/infra/index/calculate"
	"github.com/onaio/sre-tooling/infra/index/update"
	"github.com/onaio/sre-tooling/libs/cli"
)

const name string = "index"

// Index deals with commands related to resource indexes
type Index struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init initializes the command object
func (index *Index) Init(helpFlagName string, helpFlagDescription string) {
	index.flagSet = flag.NewFlagSet(index.GetName(), flag.ExitOnError)
	index.helpFlag = index.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	calc := new(calculate.Calculate)
	calc.Init(helpFlagName, helpFlagDescription)
	update := new(update.Update)
	update.Init(helpFlagName, helpFlagDescription)
	index.subCommands = []cli.Command{calc, update}
}

// GetName returns the value of the name constant
func (index *Index) GetName() string {
	return name
}

// GetDescription returns the description for the index command
func (index *Index) GetDescription() string {
	return "Related to infrastructure indexes within infrastructure groups"
}

// GetFlagSet returns a pointer to the flag.FlagSet associated to the command
func (index *Index) GetFlagSet() *flag.FlagSet {
	return index.flagSet
}

// GetSubCommands returns a slice of subcommands under the index command
// (expect empty slice if none)
func (index *Index) GetSubCommands() []cli.Command {
	return index.subCommands
}

// GetHelpFlag returns a pointer to the initialized help flag for the command
func (index *Index) GetHelpFlag() *bool {
	return index.helpFlag
}

// Process does nothing, since this command has subcommands that actually do the processing
func (index *Index) Process() {}

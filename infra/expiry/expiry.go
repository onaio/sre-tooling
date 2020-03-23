package expiry

import (
	"flag"

	"github.com/onaio/sre-tooling/infra/expiry/query"
	"github.com/onaio/sre-tooling/libs/cli"
)

const name string = "expiry"

// Expiry deals with commands related to infrastructure expiry
type Expiry struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init initializes the command object
func (expiry *Expiry) Init(helpFlagName string, helpFlagDescription string) {
	expiry.flagSet = flag.NewFlagSet(expiry.GetName(), flag.ExitOnError)
	expiry.helpFlag = expiry.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	query := new(query.Query)
	query.Init(helpFlagName, helpFlagDescription)

	expiry.subCommands = []cli.Command{query}
}

// GetName returns the value of the name constant
func (expiry *Expiry) GetName() string {
	return name
}

// GetDescription returns the description for the expiry command
func (expiry *Expiry) GetDescription() string {
	return "Related to infrastructure expiry"
}

// GetFlagSet returns a pointer to the flag.FlagSet associated to the command
func (expiry *Expiry) GetFlagSet() *flag.FlagSet {
	return expiry.flagSet
}

// GetSubCommands returns a slice of subcommands under the expiry command
// (expect empty slice if none)
func (expiry *Expiry) GetSubCommands() []cli.Command {
	return expiry.subCommands
}

// GetHelpFlag returns a pointer to the initialized help flag for the command
func (expiry *Expiry) GetHelpFlag() *bool {
	return expiry.helpFlag
}

// Process does nothing, since this command has subcommands that actually do the processing
func (expiry *Expiry) Process() {}

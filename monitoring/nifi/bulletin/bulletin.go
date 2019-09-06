package bulletin

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/monitoring/nifi/bulletin/ingest"
)

const name string = "bulletin"

// Bulletin does...
type Bulletin struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init does...
func (bulletin *Bulletin) Init(helpFlagName string, helpFlagDescription string) {
	bulletin.flagSet = flag.NewFlagSet(bulletin.GetName(), flag.ExitOnError)
	bulletin.helpFlag = bulletin.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	ingest := new(ingest.Ingest)
	ingest.Init(helpFlagName, helpFlagDescription)
	bulletin.subCommands = []cli.Command{ingest}
}

// GetName does...
func (bulletin *Bulletin) GetName() string {
	return name
}

// GetDescription does...
func (bulletin *Bulletin) GetDescription() string {
	return "NiFi bulletin specific commands"
}

// GetFlagSet does..
func (bulletin *Bulletin) GetFlagSet() *flag.FlagSet {
	return bulletin.flagSet
}

// GetSubCommands does..
func (bulletin *Bulletin) GetSubCommands() []cli.Command {
	return bulletin.subCommands
}

// GetHelpFlag does..
func (bulletin *Bulletin) GetHelpFlag() *bool {
	return bulletin.helpFlag
}

// Process does...
func (bulletin *Bulletin) Process() {}

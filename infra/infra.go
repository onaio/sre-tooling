package infra

import (
	"flag"

	"github.com/onaio/sre-tooling/infra/bill"
	"github.com/onaio/sre-tooling/infra/index"
	"github.com/onaio/sre-tooling/infra/query"
	"github.com/onaio/sre-tooling/libs/cli"
)

const name string = "infra"

type Infra struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

func (infra *Infra) Init(helpFlagName string, helpFlagDescription string) {
	infra.flagSet = flag.NewFlagSet(infra.GetName(), flag.ExitOnError)
	infra.helpFlag = infra.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	bill := new(bill.Bill)
	bill.Init(helpFlagName, helpFlagDescription)
	query := new(query.Query)
	query.Init(helpFlagName, helpFlagDescription)
	index := new(index.Index)
	index.Init(helpFlagName, helpFlagDescription)
	infra.subCommands = []cli.Command{bill, query, index}
}

func (infra *Infra) GetName() string {
	return name
}

func (infra *Infra) GetDescription() string {
	return "Infrastructure specific commands"
}

func (infra *Infra) GetFlagSet() *flag.FlagSet {
	return infra.flagSet
}

func (infra *Infra) GetSubCommands() []cli.Command {
	return infra.subCommands
}

func (infra *Infra) GetHelpFlag() *bool {
	return infra.helpFlag
}

func (infra *Infra) Process() {}

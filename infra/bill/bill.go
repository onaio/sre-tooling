package bill

import (
	"flag"

	"github.com/onaio/sre-tooling/infra/bill/validate"
	"github.com/onaio/sre-tooling/libs/cli"
)

const name string = "bill"

type Bill struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

func (bill *Bill) Init(helpFlagName string, helpFlagDescription string) {
	bill.flagSet = flag.NewFlagSet(bill.GetName(), flag.ExitOnError)
	bill.helpFlag = bill.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	validate := new(validate.Validate)
	validate.Init(helpFlagName, helpFlagDescription)
	bill.subCommands = []cli.Command{validate}
}

func (bill *Bill) GetName() string {
	return name
}

func (bill *Bill) GetDescription() string {
	return "Infrastructure billing commands"
}

func (bill *Bill) GetFlagSet() *flag.FlagSet {
	return bill.flagSet
}

func (bill *Bill) GetSubCommands() []cli.Command {
	return bill.subCommands
}

func (bill *Bill) GetHelpFlag() *bool {
	return bill.helpFlag
}

func (bill *Bill) Process() {}

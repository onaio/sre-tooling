package infra

import (
	"flag"
	"fmt"
	"os"

	"github.com/onaio/sre-tooling/infra/bill"
	"github.com/onaio/sre-tooling/infra/query"
)

const name string = "infra"

type Infra struct {
	helpFlag *bool
	flagSet  *flag.FlagSet
	bill     *bill.Bill
	query    *query.Query
}

func (infra *Infra) Init(helpFlagName string, helpFlagDescription string) {
	infra.flagSet = flag.NewFlagSet(infra.GetName(), flag.ExitOnError)
	infra.helpFlag = infra.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	infra.bill = new(bill.Bill)
	infra.bill.Init(helpFlagName, helpFlagDescription)
	infra.query = new(query.Query)
	infra.query.Init(helpFlagName, helpFlagDescription)
}

func (infra *Infra) GetName() string {
	return name
}

func (infra *Infra) GetDescription() string {
	return "Infrastructure specific commands"
}

func (infra *Infra) ParseArgs(args []string) {
	infra.flagSet.Parse(args)
	if *infra.helpFlag {
		infra.printHelp()
	} else if len(args) > 0 {
		subCommand := args[0]

		switch subCommand {
		case infra.bill.GetName():
			infra.bill.ParseArgs(args[1:])
		case infra.query.GetName():
			infra.query.ParseArgs(args[1:])
		}
	} else {
		infra.printHelp()
		os.Exit(2)
	}
}

func (infra *Infra) printHelp() {
	fmt.Println(infra.GetDescription())
	infra.flagSet.PrintDefaults()
	text := `
Common commands:
	%s		%s
	%s		%s
`
	fmt.Printf(text,
		infra.bill.GetName(), infra.bill.GetDescription(),
		infra.query.GetName(), infra.query.GetDescription())
}

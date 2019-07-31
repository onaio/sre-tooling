package bill

import (
	"flag"
	"fmt"
	"os"

	"github.com/onaio/sre-tooling/infra/bill/validate"
)

const name string = "bill"

type Bill struct {
	helpFlag *bool
	flagSet  *flag.FlagSet
	validate *validate.Validate
}

func (bill *Bill) Init(helpFlagName string, helpFlagDescription string) {
	bill.flagSet = flag.NewFlagSet(bill.GetName(), flag.ExitOnError)
	bill.helpFlag = bill.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	bill.validate = new(validate.Validate)
	bill.validate.Init(helpFlagName, helpFlagDescription)
}

func (bill *Bill) GetName() string {
	return name
}

func (bill *Bill) GetDescription() string {
	return "Infrastructure billing commands"
}

func (bill *Bill) ParseArgs(args []string) {
	bill.flagSet.Parse(args)
	if *bill.helpFlag {
		bill.printHelp()
	} else if len(args) > 0 {
		subCommand := args[0]

		switch subCommand {
		case bill.validate.GetName():
			bill.validate.ParseArgs(args[1:])
		}
	} else {
		bill.printHelp()
		os.Exit(2)
	}
}

func (bill *Bill) printHelp() {
	fmt.Println(bill.GetDescription())
	bill.flagSet.PrintDefaults()
	text := `
Common commands:
	%s		%s
`
	fmt.Printf(text, bill.validate.GetName(), bill.validate.GetDescription())
}

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/onaio/sre-tooling/infra"
)

type SRETooling struct {
	helpFlag *bool
	infra    *infra.Infra
}

func (sreTooling *SRETooling) Init(helpFlagName string, helpFlagDescription string) {
	sreTooling.helpFlag = flag.Bool(helpFlagName, false, helpFlagDescription)

	sreTooling.infra = new(infra.Infra)
	sreTooling.infra.Init(helpFlagName, helpFlagDescription)

	flag.Parse()
}

func (sreTooling *SRETooling) GetName() string {
	return os.Args[0]
}

func (sreTooling *SRETooling) GetDescription() string {
	return "SRE swiss army knife"
}

func (sreTooling *SRETooling) ParseArgs(args []string) {
	if *sreTooling.helpFlag {
		sreTooling.printHelp()
	} else if len(args) > 0 {
		subCommand := args[0]

		switch subCommand {
		case sreTooling.infra.GetName():
			sreTooling.infra.ParseArgs(args[1:])
		}
	} else {
		sreTooling.printHelp()
		os.Exit(2)
	}
}

func (sreTooling *SRETooling) printHelp() {
	fmt.Println(sreTooling.GetDescription())
	flag.PrintDefaults()
	text := `
Common commands:
	%s		%s
`
	fmt.Printf(text, sreTooling.infra.GetName(), sreTooling.infra.GetDescription())

}

package main

import (
	"flag"
	"os"

	"github.com/onaio/sre-tooling/infra"
	"github.com/onaio/sre-tooling/libs/cli"
)

type SRETooling struct {
	helpFlag    *bool
	subCommands []cli.Command
}

func (sreTooling *SRETooling) Init(helpFlagName string, helpFlagDescription string) {
	sreTooling.helpFlag = flag.Bool(helpFlagName, false, helpFlagDescription)

	infra := new(infra.Infra)
	infra.Init(helpFlagName, helpFlagDescription)
	sreTooling.subCommands = []cli.Command{infra}
	flag.Parse()
}

func (sreTooling *SRETooling) GetName() string {
	return os.Args[0]
}

func (sreTooling *SRETooling) GetDescription() string {
	return "SRE swiss army knife"
}

func (sreTooling *SRETooling) GetFlagSet() *flag.FlagSet {
	return nil
}

func (sreTooling *SRETooling) GetSubCommands() []cli.Command {
	return sreTooling.subCommands
}

func (sreTooling *SRETooling) GetHelpFlag() *bool {
	return sreTooling.helpFlag
}

func (sreTooling *SRETooling) Process() {}

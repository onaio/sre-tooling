package audit

import (
	"flag"

	"github.com/onaio/sre-tooling/audit/ssl"
	"github.com/onaio/sre-tooling/libs/cli"
)

const name string = "audit"

type Audit struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

func (audit *Audit) Init(helpFlagName string, helpFlagDescription string) {
	audit.flagSet = flag.NewFlagSet(audit.GetName(), flag.ExitOnError)
	audit.helpFlag = audit.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	ssl := new(ssl.SSL)
	ssl.Init(helpFlagName, helpFlagDescription)
	audit.subCommands = []cli.Command{ssl}
}

func (audit *Audit) GetName() string {
	return name
}

func (audit *Audit) GetDescription() string {
	return "Audit specific commands"
}

func (audit *Audit) GetFlagSet() *flag.FlagSet {
	return audit.flagSet
}

func (audit *Audit) GetSubCommands() []cli.Command {
	return audit.subCommands
}

func (audit *Audit) GetHelpFlag() *bool {
	return audit.helpFlag
}

func (audit *Audit) Process() {}

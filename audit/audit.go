package audit

import (
	"flag"
	"fmt"

	"github.com/onaio/sre-tooling/libs/notification"

	auditlib "github.com/onaio/sre-tooling/libs/audit"
	"github.com/onaio/sre-tooling/libs/cli"
)

const name string = "audit"

type Audit struct {
	helpFlag      *bool
	flagSet       *flag.FlagSet
	auditFileFlag *string
	subCommands   []cli.Command
}

func (audit *Audit) Init(helpFlagName string, helpFlagDescription string) {
	audit.flagSet = flag.NewFlagSet(audit.GetName(), flag.ExitOnError)
	audit.helpFlag = audit.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	audit.auditFileFlag = audit.flagSet.String(
		"audit-file",
		"",
		"Absolute path to yaml file containing audit tests to run",
	)
	audit.subCommands = []cli.Command{}
}

func (audit *Audit) GetName() string {
	return name
}

func (audit *Audit) GetDescription() string {
	return "Run audit tests"
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

func (audit *Audit) Process() {
	// validate auditFile is provided
	if len(*audit.auditFileFlag) == 0 {
		notification.SendMessage("Audit file path is required")
		cli.ExitCommandInterpretationError()
	}

	auditResults, err := auditlib.Run(*audit.auditFileFlag)
	if err != nil {
		notification.SendMessage(err.Error())
		cli.ExitCommandExecutionError()
	}

	for _, res := range auditResults {
		fmt.Printf("[%s] [%s] %s\n", res.Type, res.Status, res.StatusMessage)
	}
}

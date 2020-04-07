package version

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/notification"
)

const name string = "version"

type Version struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

func (version *Version) Init(helpFlagName string, helpFlagDescription string) {
	version.flagSet = flag.NewFlagSet(version.GetName(), flag.ExitOnError)
	version.helpFlag = version.flagSet.Bool(helpFlagName, false, helpFlagDescription)
}

func (version *Version) GetName() string {
	return name
}

func (version *Version) GetDescription() string {
	return "Returns the SRE Tooling version"
}

func (version *Version) GetFlagSet() *flag.FlagSet {
	return version.flagSet
}

func (version *Version) GetSubCommands() []cli.Command {
	return version.subCommands
}

func (version *Version) GetHelpFlag() *bool {
	return version.helpFlag
}

func (version *Version) Process() {
	notification.SendMessage(versionString)
}

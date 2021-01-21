package ssl

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
)

const name string = "ssl"

type SSL struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	hostsFile   *string
	subCommands []cli.Command
}

func (ssl *SSL) Init(helpFlagName string, helpFlagDescription string) {
	ssl.flagSet = flag.NewFlagSet(ssl.GetName(), flag.ExitOnError)
	ssl.helpFlag = ssl.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	// TODO: Add hostsFile flag
	ssl.subCommands = []cli.Command{}
}

func (ssl *SSL) GetName() string {
	return name
}

func (ssl *SSL) GetDescription() string {
	return "Run SSL scan on provided hosts"
}

func (ssl *SSL) GetFlagSet() *flag.FlagSet {
	return ssl.flagSet
}

func (ssl *SSL) GetSubCommands() []cli.Command {
	return ssl.subCommands
}

func (ssl *SSL) GetHelpFlag() *bool {
	return ssl.helpFlag
}

func (ssl *SSL) Process() {

}

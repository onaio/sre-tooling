package version

import (
	"flag"
	"fmt"

	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/version"
)

const name string = "version"

// Version holds data for the version sub-command
type Version struct {
	helpFlag    *bool
	flagSet     *flag.FlagSet
	subCommands []cli.Command
}

// Init initializes Version struct
func (v *Version) Init(helpFlagName string, helpFlagDescription string) {
	v.flagSet = flag.NewFlagSet(v.GetName(), flag.ExitOnError)
	v.helpFlag = v.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	v.subCommands = []cli.Command{}
}

// GetName returns the name of the version sub-command
func (v *Version) GetName() string {
	return name
}

// GetDescription returns the description for the version sub-command
func (v *Version) GetDescription() string {
	return "Print sre-tooling version"
}

// GetFlagSet returns a flag.FlagSet corresponding to the version sub-command
func (v *Version) GetFlagSet() *flag.FlagSet {
	return v.flagSet
}

// GetSubCommands returns a list of sub-commands under the version sub-command
func (v *Version) GetSubCommands() []cli.Command {
	return v.subCommands
}

// GetHelpFlag returns the value for the help flag in the version sub-command.
func (v *Version) GetHelpFlag() *bool {
	return v.helpFlag
}

// Process executes the logic involved in printing sre-tooling version
func (v *Version) Process() {
	fmt.Println(version.Current)
}

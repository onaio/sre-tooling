package main

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
)

const helpFlagName string = "help"
const helpFlagDescription string = "Prints the full help message"

func main() {
	sreTooling := new(SRETooling)
	sreTooling.Init(helpFlagName, helpFlagDescription)
	cli.ParseArgs(sreTooling, flag.Args())
}

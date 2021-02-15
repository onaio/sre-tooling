package main

import (
	"flag"

	"github.com/onaio/sre-tooling/libs/cli"
	SREToolingVersion "github.com/onaio/sre-tooling/libs/version"
)

const helpFlagName string = "help"
const helpFlagDescription string = "Prints the full help message"

/*
version will be set by GoReleaser.
It will be the current Git tag (with v prefix stripped) or
the name of the snapshot if you're using the --snapshot flag.
*/
var version = "master"

func main() {
	SREToolingVersion.Current = version

	sreTooling := new(SRETooling)
	sreTooling.Init(helpFlagName, helpFlagDescription)
	cli.ParseArgs(sreTooling, flag.Args())
}

package main

import "flag"

const helpFlagName string = "help"
const helpFlagDescription string = "Prints the full help message"

func main() {
	sreTooling := SRETooling{}
	sreTooling.Init(helpFlagName, helpFlagDescription)
	sreTooling.ParseArgs(flag.Args())
}

package ssl

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/onaio/sre-tooling/libs/notification"
	"gopkg.in/yaml.v2"

	"github.com/onaio/sre-tooling/libs/cli"
	ssllib "github.com/onaio/sre-tooling/libs/ssl"
)

const name string = "ssl"

type SSL struct {
	helpFlag      *bool
	flagSet       *flag.FlagSet
	hostsFileFlag *string
	subCommands   []cli.Command
}

func (ssl *SSL) Init(helpFlagName string, helpFlagDescription string) {
	ssl.flagSet = flag.NewFlagSet(ssl.GetName(), flag.ExitOnError)
	ssl.helpFlag = ssl.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	ssl.hostsFileFlag = ssl.flagSet.String("hosts-file", "", "Absolute path to yaml file containing hosts to scan.")
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
	// validate hostsFile is provided
	if len(*ssl.hostsFileFlag) == 0 {
		notification.SendMessage("Hosts file path is required")
		cli.ExitCommandInterpretationError()
	}

	hostsFile, err := ioutil.ReadFile(*ssl.hostsFileFlag)
	if err != nil {
		notification.SendMessage(err.Error())
		cli.ExitCommandExecutionError()
	}

	hosts := ssllib.SSLHosts{}
	err = yaml.Unmarshal(hostsFile, &hosts)
	if err != nil {
		notification.SendMessage(err.Error())
		cli.ExitCommandExecutionError()
	}

	err = hosts.Scan()
	if err != nil {
		notification.SendMessage(err.Error())
		cli.ExitCommandExecutionError()
	}

	for _, host := range hosts.Hosts {
		fmt.Printf("Scan Info (%s) - %s\n", host.Host, host.ScanInfo.Status)

		if host.ScanInfo.Status == "ERROR" {
			fmt.Printf("\tError: %s\n", host.ScanInfoError)
		}

		for _, endpoint := range host.ScanInfo.Endpoints {
			fmt.Printf("\tIP: %s, Server name: %s, Grade: %s\n", endpoint.IPAdress, endpoint.ServerName, endpoint.Grade)
		}

		fmt.Printf("\n")
	}
}

package prune

import (
	"flag"
	"fmt"
	"time"

	"github.com/onaio/sre-tooling/infra/expiry/query"
	"github.com/onaio/sre-tooling/libs/cli"
	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/infra"
	"github.com/onaio/sre-tooling/libs/notification"
)

// Make sure there's a -safe flag (that's defaulted to true) where, e.g you are not allowed to terminate VMs that aren't already shut down
// Make sure there's an -action flag (with possible values being 'terminate' & 'shutdown' and default value is 'shutdown') for controlling what action
//  should be taken on expired resources
// Make sure at least one filter flag (region, resource, tag) is provided to avoid catastrophic situations where all resources in an entire
// cloud account are shut-down
// Make sure there's a -yes flag (defaulted to false) that if not set to true will require a confirmation before resources are pruned

const name string = "prune"
const actionStop string = "stop"
const actionTerminate string = "terminate"

// Prune queries then notifies (using configured notification channels) infrastructure that has expired
type Prune struct {
	helpFlag             *bool
	flagSet              *flag.FlagSet
	providerFlag         *flags.StringArray
	regionFlag           *flags.StringArray
	typeFlag             *flags.StringArray
	tagFlag              *flags.StringArray
	maxAgeFlag           *string
	expiryTagFlag        *string
	expiryTagNAValueFlag *string
	expiryTagFormatFlag  *string
	safeFlag             *bool
	actionFlag           *string
	yesFlag              *bool
	subCommands          []cli.Command
}

// Init initializes the command object
func (prune *Prune) Init(helpFlagName string, helpFlagDescription string) {
	prune.flagSet = flag.NewFlagSet(prune.GetName(), flag.ExitOnError)
	prune.helpFlag = prune.flagSet.Bool(helpFlagName, false, helpFlagDescription)
	prune.safeFlag = prune.flagSet.Bool("unsafe", false, "If set to true, command will not error if you try to terminate a resource that is stoppable but not stopped")
	prune.actionFlag = prune.flagSet.String("action", actionStop, fmt.Sprintf("What action to take on a resource that has expired. Possible values are '%s' and '%s'", actionStop, actionTerminate))
	prune.yesFlag = prune.flagSet.Bool("yes", false, "Whether to skip requiring a confirmation before pruning a resource")

	prune.providerFlag,
		prune.regionFlag,
		prune.typeFlag,
		prune.tagFlag,
		prune.maxAgeFlag,
		prune.expiryTagFlag,
		prune.expiryTagNAValueFlag,
		prune.expiryTagFormatFlag = query.AddQueryFlags(prune.flagSet)
}

// GetName returns the value of the name constant
func (prune *Prune) GetName() string {
	return name
}

// GetDescription returns the description for the prune command
func (prune *Prune) GetDescription() string {
	return "Prunes (stops or destroys) infrastructure that has expired"
}

// GetFlagSet returns a pointer to the flag.FlagSet associated to the command
func (prune *Prune) GetFlagSet() *flag.FlagSet {
	return prune.flagSet
}

// GetSubCommands returns a slice of subcommands under the prune command
// (expect empty slice if none)
func (prune *Prune) GetSubCommands() []cli.Command {
	return prune.subCommands
}

// GetHelpFlag returns a pointer to the initialized help flag for the command
func (prune *Prune) GetHelpFlag() *bool {
	return prune.helpFlag
}

// Process fetches the list of infrastructure that matches the criteria provided by the user
// and that has expired and sends notifications to the configured notification channels
func (prune *Prune) Process() {
	hasResourceErr := false
	resourceErr := query.GetExpiredResources(
		prune.providerFlag,
		prune.regionFlag,
		prune.typeFlag,
		prune.tagFlag,
		prune.maxAgeFlag,
		prune.expiryTagFlag,
		prune.expiryTagNAValueFlag,
		prune.expiryTagFormatFlag,
		func(resource *infra.Resource, hasExpired bool, expiryTime time.Time, err error) {
			if err != nil {
				notification.SendMessage(fmt.Errorf("Could not figure out which resources have expired: %w", err).Error())
				hasResourceErr = true
				return
			}

			if !hasExpired {
				return
			}

			// prune the resource
		},
	)

	if resourceErr != nil {
		notification.SendMessage(resourceErr.Error())
		hasResourceErr = true
	}

	if hasResourceErr {
		cli.ExitCommandExecutionError()
	}
}

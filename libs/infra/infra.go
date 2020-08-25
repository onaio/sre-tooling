package infra

import (
	"flag"
	"fmt"
	"reflect"
	"strings"

	"github.com/onaio/sre-tooling/libs/cli/flags"
	"github.com/onaio/sre-tooling/libs/infra/aws"
	"github.com/onaio/sre-tooling/libs/types"
)

type Provider interface {
	Init() error
	GetName() string
	GetResources(filter *types.InfraFilter) ([]*types.InfraResource, error)
	UpdateResourceTag(resource *types.InfraResource, tagKey *string, tagValue *string) error
	UpdateResourceState(resource *types.InfraResource, safe bool, state string) error
}

const tagFlagSeparator = ":"

func GetResources(filter *types.InfraFilter) ([]*types.InfraResource, error) {
	allResources := []*types.InfraResource{}

	providers, providerErr := getProviders()
	if providerErr != nil {
		return nil, providerErr
	}

	for _, curProvider := range providers {
		if considerProvider(curProvider, filter) {
			pResources, curErr := curProvider.GetResources(filter)
			if curErr != nil {
				return nil, curErr
			}
			allResources = append(allResources, pResources...)
		}
	}

	return allResources, nil
}

func getProviders() ([]Provider, error) {
	providers := []Provider{}

	aws := new(aws.AWS)
	awsErr := aws.Init()
	if awsErr != nil {
		return nil, awsErr
	}
	providers = append(providers, aws)

	return providers, nil
}

func GetTagKeys(resource *types.InfraResource) []string {
	keyObjects := reflect.ValueOf(resource.Tags).MapKeys()
	keys := make([]string, len(keyObjects))
	for i := 0; i < len(keyObjects); i++ {
		keys[i] = keyObjects[i].String()
	}

	return keys
}

func AddFilterFlags(flagSet *flag.FlagSet) (*flags.StringArray, *flags.StringArray, *flags.StringArray, *flags.StringArray) {
	providerFlag := new(flags.StringArray)
	flagSet.Var(providerFlag, "filter-provider", "Name of provider to filter using. Multiple values can be provided by specifying multiple -filter-provider")
	regionFlag := new(flags.StringArray)
	flagSet.Var(regionFlag, "filter-region", "Name of a provider region to filter using. Multiple values can be provided by specifying multiple -filter-region")
	typeFlag := new(flags.StringArray)
	flagSet.Var(typeFlag, "filter-type", "Resource type to filter using e.g. \"EC2\". Multiple values can be provided by specifying multiple -filter-type")
	tagFlag := new(flags.StringArray)
	flagSet.Var(tagFlag, "filter-tag", "Resource tag to filter using. Use the format \"tagKey"+tagFlagSeparator+"tagValue\". Multiple values can be provided by specifying multiple -filter-tag")

	return providerFlag, regionFlag, typeFlag, tagFlag
}

func GetFiltersFromCommandFlags(providerFlag *flags.StringArray, regionFlag *flags.StringArray, typeFlag *flags.StringArray, tagFlag *flags.StringArray) *types.InfraFilter {
	filter := types.InfraFilter{}
	if len(*providerFlag) > 0 {
		filter.Providers = *providerFlag
	}
	if len(*typeFlag) > 0 {
		filter.ResourceTypes = *typeFlag
	}
	if len(*regionFlag) > 0 {
		filter.Regions = *regionFlag
	}
	if len(*tagFlag) > 0 {
		for _, curTagPair := range *tagFlag {
			curKeyValue := strings.Split(curTagPair, tagFlagSeparator)
			if len(curKeyValue) == 2 {
				if filter.Tags == nil {
					filter.Tags = make(map[string]string)
				}
				filter.Tags[curKeyValue[0]] = curKeyValue[1]
			}
		}
	}

	return &filter
}

func UpdateResourceTag(resource *types.InfraResource, tagKey *string, tagValue *string) error {
	providers, providerErr := getProviders()
	if providerErr != nil {
		return providerErr
	}

	for _, curProvider := range providers {
		if curProvider.GetName() == resource.Provider {
			return curProvider.UpdateResourceTag(resource, tagKey, tagValue)
		}
	}

	return fmt.Errorf("Provider '%s' isn't implemented yet", resource.Provider)
}

func considerProvider(providerIface interface{}, filter *types.InfraFilter) bool {
	provider := providerIface.(Provider)
	if len(filter.Providers) == 0 {
		return true
	}

	for _, curProviderName := range filter.Providers {
		if strings.ToLower(curProviderName) == strings.ToLower(provider.GetName()) {
			return true
		}
	}

	return false
}

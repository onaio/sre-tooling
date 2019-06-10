package cloud

import (
	"reflect"
	"time"
)

type Provider interface {
	getAllResources(filter *Filter) ([]*Resource, error)
}

type Resource struct {
	Provider     string
	ID           string
	Location     string
	ResourceType string
	LaunchTime   time.Time
	Tags         map[string]string
}

type Filter struct {
	Provider     string
	ResourceType string
	Tags         map[string]string
}

func GetAllCloudResources(filter *Filter) ([]*Resource, error) {
	aws := AWS{}
	allResources, awsErr := aws.getAllResources(filter)
	if awsErr != nil {
		return nil, awsErr
	}

	return allResources, nil
}

func GetTagKeys(resource *Resource) []string {
	keyObjects := reflect.ValueOf(resource.Tags).MapKeys()
	keys := make([]string, len(keyObjects))
	for i := 0; i < len(keyObjects); i++ {
		keys[i] = keyObjects[i].String()
	}

	return keys
}

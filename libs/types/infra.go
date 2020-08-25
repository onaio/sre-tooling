package types

import (
	"time"
)

type InfraResource struct {
	Provider     string
	ID           string
	Location     string
	ResourceType string
	LaunchTime   time.Time
	Tags         map[string]string
	Properties   map[string]string
	Data         map[string]string
}

type InfraFilter struct {
	Providers     []string
	ResourceTypes []string
	Regions       []string
	Tags          map[string]string
}

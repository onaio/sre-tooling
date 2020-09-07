package types

// CostAndUsageOutput defines output to be returned by `GetCostAndUsage`
type CostAndUsageOutput struct {
	Provider  string
	Groups    map[string]float64
	StartDate string
	EndDate   string
}

// CostAndUsageFilter defines parameters used to filter costs
type CostAndUsageFilter struct {
	Providers     []string
	ResourceTypes []string
	Regions       []string
	Tags          map[string]string
	Granularity   string
	StartDate     string
	EndDate       string
	GroupBy       map[string]string
}

type CostSpikeOutput struct {
	Provider         string
	GroupKey         string
	StartDate        string
	EndDate          string
	CurPeriodAmount  float64
	PrevPeriodAmount float64
	IncreaseRate     float64
}

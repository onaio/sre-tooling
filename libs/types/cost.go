package types

// CostAndUsagePeriod defines period used to calculate costs and usage
type CostAndUsagePeriod struct {
	StartDate string
	EndDate   string
}

// CostAndUsageOutput defines output to be returned by `GetCostAndUsage`
type CostAndUsageOutput struct {
	Provider string
	Groups   map[string]float64
	Period   *CostAndUsagePeriod
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
	SortRate      string
	SortCurAmount string
}

type CostSpikeOutput struct {
	Provider         string
	GroupKey         string
	CurPeriod        *CostAndUsagePeriod
	PrevPeriod       *CostAndUsagePeriod
	CurPeriodAmount  float64
	PrevPeriodAmount float64
	IncreaseRate     float64
}

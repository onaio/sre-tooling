package types

type CostAndUsageOutput struct {
	Date   string
	Key    string
	Amount string
}

type CostAndUsageFilter struct {
	Providers  []string
	Ganularity string
	StartDate  string
	EndDate    string
}

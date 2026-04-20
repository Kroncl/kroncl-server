package dm

import "time"

// ---------
// ANALYSIS TYPES
// ---------

type GroupBy string

const (
	GroupByType     GroupBy = "type"
	GroupByStatus   GroupBy = "status"
	GroupByEmployee GroupBy = "employee"
	GroupByClient   GroupBy = "client"
	GroupByDay      GroupBy = "day"
	GroupByMonth    GroupBy = "month"
	GroupByYear     GroupBy = "year"
)

type GroupedStats struct {
	GroupKey  string `json:"group_key"`
	GroupName string `json:"group_name"`
	Count     int64  `json:"count"`
}

type DealAnalysisSummary struct {
	TotalDeals        int64   `json:"total_deals"`
	DefaultStatusID   *string `json:"default_status_id,omitempty"`
	DefaultStatusName *string `json:"default_status_name,omitempty"`
	DealsInDefault    int64   `json:"deals_in_default"`
	AvgDealAmount     float64 `json:"avg_deal_amount,omitempty"`
}

type GetAnalysisParams struct {
	StartDate  *time.Time
	EndDate    *time.Time
	TypeID     *string
	StatusID   *string
	ClientID   *string
	EmployeeID *string
}

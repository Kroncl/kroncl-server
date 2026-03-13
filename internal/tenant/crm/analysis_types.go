package crm

type GroupBy string

const (
	GroupBySource GroupBy = "source"
	GroupByDay    GroupBy = "day"
	GroupByMonth  GroupBy = "month"
)

type GroupedStats struct {
	GroupKey     string `json:"group_key"`     // source_id, date
	GroupName    string `json:"group_name"`    // source_name, date
	ClientsCount int64  `json:"clients_count"` // количество клиентов
}

// ClientsSummary represents summary statistics for a period
type ClientsSummary struct {
	TotalClients      int64 `json:"total_clients"`
	ActiveClients     int64 `json:"active_clients"`
	InactiveClients   int64 `json:"inactive_clients"`
	IndividualClients int64 `json:"individual_clients"`
	LegalClients      int64 `json:"legal_clients"`
	NewClients        int64 `json:"new_clients"` // клиенты за период
}

package fm

type GroupBy string

const (
	GroupByCategory GroupBy = "category"
	GroupByEmployee GroupBy = "employee"
	GroupByDay      GroupBy = "day"
	GroupByMonth    GroupBy = "month"
)

type GroupedStats struct {
	GroupKey  string `json:"group_key"`  // category_id, employee_id, date
	GroupName string `json:"group_name"` // category_name, employee_name, date
	Income    int64  `json:"income"`
	Expense   int64  `json:"expense"`
	Net       int64  `json:"net"`
	Count     int64  `json:"count"`
}

// AnalysisSummary represents financial summary for a period
type AnalysisSummary struct {
	TotalIncome      int64   `json:"total_income"`
	TotalExpense     int64   `json:"total_expense"`
	NetBalance       int64   `json:"net_balance"`
	TransactionCount int64   `json:"transaction_count"`
	AvgTransaction   float64 `json:"avg_transaction"`
}

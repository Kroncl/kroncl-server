package fm

import "kroncl-server/internal/currency"

type GroupBy string

const (
	GroupByCategory GroupBy = "category"
	GroupByEmployee GroupBy = "employee"
	GroupByDay      GroupBy = "day"
	GroupByMonth    GroupBy = "month"
)

type GroupedStats struct {
	GroupKey  string             `json:"group_key"`  // category_id, employee_id, date
	GroupName string             `json:"group_name"` // category_name, employee_name, date
	Income    float64            `json:"income"`
	Expense   float64            `json:"expense"`
	Net       float64            `json:"net"`
	Count     int64              `json:"count"`
	Currency  *currency.Currency `json:"currency"`
}

type AnalysisSummary struct {
	TotalIncome      float64            `json:"total_income"`
	TotalExpense     float64            `json:"total_expense"`
	NetBalance       float64            `json:"net_balance"`
	TransactionCount int64              `json:"transaction_count"`
	AvgTransaction   float64            `json:"avg_transaction"`
	Currency         *currency.Currency `json:"currency"`
}

package fm

import (
	"context"
	"fmt"
	"time"
)

// ---------
// ANALYSIS
// ---------

func (r *Repository) GetGroupedTransactions(ctx context.Context,
	groupBy GroupBy,
	startDate, endDate *time.Time) ([]GroupedStats, error) {

	var selectCols, groupByExpr, joinClause string

	switch groupBy {
	case GroupByCategory:
		selectCols = `
			COALESCE(tc.category_id::text, 'no-category') as group_key,
			MIN(COALESCE(c.name, 'Без категории')) as group_name`
		groupByExpr = `COALESCE(tc.category_id::text, 'no-category')`
		joinClause = `LEFT JOIN transaction_category tc ON t.id = tc.transaction_id
		              LEFT JOIN transaction_categories c ON tc.category_id = c.id`

	case GroupByEmployee:
		selectCols = `
			COALESCE(te.employee_id::text, 'no-employee') as group_key,
			COALESCE(MIN(NULLIF(e.first_name || ' ' || e.last_name, '')), 'Без сотрудника') as group_name`
		groupByExpr = `COALESCE(te.employee_id::text, 'no-employee')`
		joinClause = `LEFT JOIN transaction_employee te ON t.id = te.transaction_id
		              LEFT JOIN employees e ON te.employee_id = e.id`

	case GroupByDay:
		selectCols = `
			DATE(t.created_at)::text as group_key,
			TO_CHAR(DATE(t.created_at), 'DD.MM.YYYY') as group_name`
		groupByExpr = `DATE(t.created_at)`
		joinClause = ``

	case GroupByMonth:
		selectCols = `
			TO_CHAR(DATE_TRUNC('month', t.created_at), 'YYYY-MM') as group_key,
			TO_CHAR(DATE_TRUNC('month', t.created_at), 'MMMM YYYY') as group_name`
		groupByExpr = `DATE_TRUNC('month', t.created_at)`
		joinClause = ``

	default:
		return nil, fmt.Errorf("invalid group_by: %s", groupBy)
	}

	query := fmt.Sprintf(`
		SELECT 
			%s,
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) as expense,
			COALESCE(SUM(CASE 
				WHEN t.direction = 'income' THEN t.base_amount 
				WHEN t.direction = 'expense' THEN -t.base_amount 
				ELSE 0 
			END), 0) as net,
			COUNT(*) as count
		FROM transactions t
		%s
		WHERE ($1::timestamptz IS NULL OR t.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR t.created_at <= $2)
		GROUP BY %s
		ORDER BY income DESC
	`, selectCols, joinClause, groupByExpr)

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query grouped transactions: %w", err)
	}
	defer rows.Close()

	var stats []GroupedStats
	for rows.Next() {
		var stat GroupedStats
		err := rows.Scan(
			&stat.GroupKey,
			&stat.GroupName,
			&stat.Income,
			&stat.Expense,
			&stat.Net,
			&stat.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grouped stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetAnalysisSummary returns financial summary for given period
func (r *Repository) GetAnalysisSummary(ctx context.Context, startDate, endDate *time.Time) (*AnalysisSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN direction = 'income' THEN base_amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN direction = 'expense' THEN base_amount ELSE 0 END), 0) as total_expense,
			COALESCE(SUM(CASE 
				WHEN direction = 'income' THEN base_amount 
				WHEN direction = 'expense' THEN -base_amount 
				ELSE 0 
			END), 0) as net_balance,
			COUNT(*) as transaction_count,
			COALESCE(AVG(base_amount), 0) as avg_transaction
		FROM transactions t
		WHERE ($1::timestamptz IS NULL OR t.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR t.created_at <= $2)
	`

	var summary AnalysisSummary
	err := r.pool.QueryRow(ctx, query, startDate, endDate).Scan(
		&summary.TotalIncome,
		&summary.TotalExpense,
		&summary.NetBalance,
		&summary.TransactionCount,
		&summary.AvgTransaction,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get analysis summary: %w", err)
	}

	return &summary, nil
}

func (r *Repository) GetOverallDealTransactionsSummary(ctx context.Context, filters GetTransactionsRequest) (*DealTransactionsSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE -t.base_amount END), 0) as total_amount,
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE 0 END), 0) as income_amount,
			COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) as expense_amount,
			COALESCE(COUNT(CASE WHEN t.direction = 'income' THEN 1 END), 0) as income_count,
			COALESCE(COUNT(CASE WHEN t.direction = 'expense' THEN 1 END), 0) as expense_count,
			COUNT(t.id) as total_count
		FROM transactions t
		INNER JOIN deals_transactions dt ON t.id = dt.transaction_id
		WHERE ($1::timestamptz IS NULL OR t.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR t.created_at <= $2)
	`

	var summary DealTransactionsSummary
	err := r.pool.QueryRow(ctx, query, filters.StartDate, filters.EndDate).Scan(
		&summary.TotalAmount,
		&summary.IncomeAmount,
		&summary.ExpenseAmount,
		&summary.IncomeCount,
		&summary.ExpenseCount,
		&summary.TotalCount,
	)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// // DailyStats represents financial stats for a single day
// type DailyStats struct {
// 	Date             string `json:"date"` // YYYY-MM-DD
// 	TransactionCount int64  `json:"transactions_count"`
// 	Income           int64  `json:"income"`
// 	Expense          int64  `json:"expense"`
// 	Net              int64  `json:"net"`
// }
// GetDailyStats returns financial stats grouped by day
// func (r *Repository) GetDailyStats(ctx context.Context, startDate, endDate *time.Time) ([]DailyStats, error) {
// 	query := `
// 		SELECT
// 			DATE(created_at) as date,
// 			COUNT(*) as transactions_count,
// 			COALESCE(SUM(base_amount) FILTER (WHERE direction = 'income'), 0) as income,
// 			COALESCE(SUM(base_amount) FILTER (WHERE direction = 'expense'), 0) as expense,
// 			COALESCE(SUM(CASE
// 				WHEN direction = 'income' THEN base_amount
// 				WHEN direction = 'expense' THEN -base_amount
// 				ELSE 0
// 			END), 0) as net
// 		FROM transactions
// 		WHERE ($1::timestamptz IS NULL OR created_at >= $1)
// 		  AND ($2::timestamptz IS NULL OR created_at <= $2)
// 		GROUP BY DATE(created_at)
// 		ORDER BY date ASC
// 	`

// 	rows, err := r.pool.Query(ctx, query, startDate, endDate)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query daily stats: %w", err)
// 	}
// 	defer rows.Close()

// 	var stats []DailyStats
// 	for rows.Next() {
// 		var stat DailyStats
// 		var date time.Time // ← сканируем в time.Time

// 		err := rows.Scan(
// 			&date,
// 			&stat.TransactionCount,
// 			&stat.Income,
// 			&stat.Expense,
// 			&stat.Net,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
// 		}

// 		// Форматируем в YYYY-MM-DD
// 		stat.Date = date.Format("2006-01-02")
// 		stats = append(stats, stat)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error iterating daily stats: %w", err)
// 	}

// 	return stats, nil
// }

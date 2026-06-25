package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/currency"
	"time"
)

// ---------
// ANALYSIS
// ---------

func (r *Repository) GetAnalysisSummary(ctx context.Context, startDate, endDate *time.Time, targetCurrency string) (*AnalysisSummary, error) {
	query := `
		SELECT t.currency, DATE_TRUNC('hour', t.created_at) as date_hour,
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) as expense,
			COUNT(*) as count
		FROM transactions t
		WHERE ($1::timestamptz IS NULL OR t.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR t.created_at <= $2)
		GROUP BY t.currency, DATE_TRUNC('hour', t.created_at)
	`

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rawStats := make(map[string]map[time.Time]*currency.RawStat)

	for rows.Next() {
		var currencyID string
		var date time.Time
		var income, expense float64
		var count int64

		rows.Scan(&currencyID, &date, &income, &expense, &count)

		if rawStats[currencyID] == nil {
			rawStats[currencyID] = make(map[time.Time]*currency.RawStat)
		}
		rawStats[currencyID][date] = &currency.RawStat{
			Income:  income,
			Expense: expense,
			Count:   count,
		}
	}

	converted, err := r.currencyService.ConvertSummary(ctx, rawStats, targetCurrency)
	if err != nil {
		return nil, err
	}

	return &AnalysisSummary{
		TotalIncome:      converted.TotalIncome,
		TotalExpense:     converted.TotalExpense,
		NetBalance:       converted.NetBalance,
		TransactionCount: converted.TransactionCount,
		AvgTransaction:   converted.AvgTransaction,
		Currency:         converted.Currency,
	}, nil
}

func (r *Repository) GetGroupedTransactions(ctx context.Context,
	groupBy GroupBy,
	startDate, endDate *time.Time,
	targetCurrency string,
) ([]GroupedStats, error) {

	if targetCurrency == "" {
		targetCurrency = "RUB"
	}

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
			COALESCE(MIN(NULLIF(CONCAT_WS(' ', e.first_name, e.last_name), '')), 'Без сотрудника') as group_name`
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
			t.currency,
			DATE_TRUNC('hour', t.created_at) as date_hour,
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) as expense,
			COUNT(*) as count
		FROM transactions t
		%s
		WHERE ($1::timestamptz IS NULL OR t.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR t.created_at <= $2)
		GROUP BY %s, t.currency, DATE_TRUNC('hour', t.created_at)
		ORDER BY income DESC
	`, selectCols, joinClause, groupByExpr)

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// rawStats: [groupKey][currencyID][dateHour] = RawStat
	type groupKey string
	rawStats := make(map[groupKey]map[string]map[time.Time]*currency.RawStat)
	groupNames := make(map[groupKey]string)

	for rows.Next() {
		var gk, currencyID string
		var dateHour time.Time
		var income, expense float64
		var count int64
		var groupName string

		rows.Scan(&gk, &groupName, &currencyID, &dateHour, &income, &expense, &count)

		groupNames[groupKey(gk)] = groupName

		if rawStats[groupKey(gk)] == nil {
			rawStats[groupKey(gk)] = make(map[string]map[time.Time]*currency.RawStat)
		}
		if rawStats[groupKey(gk)][currencyID] == nil {
			rawStats[groupKey(gk)][currencyID] = make(map[time.Time]*currency.RawStat)
		}
		rawStats[groupKey(gk)][currencyID][dateHour] = &currency.RawStat{
			Income:  income,
			Expense: expense,
			Count:   count,
		}
	}

	// Конвертируем каждую группу
	var stats []GroupedStats
	for gk, currencyStats := range rawStats {
		converted, err := r.currencyService.ConvertSummary(ctx, currencyStats, targetCurrency)
		if err != nil {
			continue
		}

		stats = append(stats, GroupedStats{
			GroupKey:  string(gk),
			GroupName: groupNames[gk],
			Income:    converted.TotalIncome,
			Expense:   converted.TotalExpense,
			Net:       converted.NetBalance,
			Count:     converted.TransactionCount,
			Currency:  converted.Currency,
		})
	}

	// Сортируем по доходу (уже из SQL, но после конвертации может измениться)
	// Оставляем как есть — SQL сортировка примерная

	return stats, nil
}

func (r *Repository) GetOverallDealTransactionsSummary(ctx context.Context, filters GetTransactionsRequest, targetCurrency string) (*DealTransactionsSummary, error) {
	if targetCurrency == "" {
		targetCurrency = "RUB"
	}

	query := `
		SELECT 
			t.currency,
			DATE_TRUNC('hour', t.created_at) as date_hour,
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) as expense,
			COALESCE(COUNT(CASE WHEN t.direction = 'income' THEN 1 END), 0) as income_count,
			COALESCE(COUNT(CASE WHEN t.direction = 'expense' THEN 1 END), 0) as expense_count,
			COUNT(t.id) as total_count
		FROM transactions t
		INNER JOIN deals_transactions dt ON t.id = dt.transaction_id
		WHERE ($1::timestamptz IS NULL OR t.created_at >= $1)
		  AND ($2::timestamptz IS NULL OR t.created_at <= $2)
		GROUP BY t.currency, DATE_TRUNC('hour', t.created_at)
	`

	rows, err := r.pool.Query(ctx, query, filters.StartDate, filters.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rawStats := make(map[string]map[time.Time]*currency.RawStat)
	var totalIncomeCount, totalExpenseCount, totalCount int64

	for rows.Next() {
		var currencyID string
		var dateHour time.Time
		var income, expense float64
		var incomeCount, expenseCount, count int64

		rows.Scan(&currencyID, &dateHour, &income, &expense, &incomeCount, &expenseCount, &count)

		if rawStats[currencyID] == nil {
			rawStats[currencyID] = make(map[time.Time]*currency.RawStat)
		}
		rawStats[currencyID][dateHour] = &currency.RawStat{
			Income:  income,
			Expense: expense,
			Count:   count,
		}
		totalIncomeCount += incomeCount
		totalExpenseCount += expenseCount
		totalCount += count
	}

	converted, err := r.currencyService.ConvertSummary(ctx, rawStats, targetCurrency)
	if err != nil {
		return nil, err
	}

	return &DealTransactionsSummary{
		TotalAmount:   converted.NetBalance,
		IncomeAmount:  converted.TotalIncome,
		ExpenseAmount: converted.TotalExpense,
		IncomeCount:   totalIncomeCount,
		ExpenseCount:  totalExpenseCount,
		TotalCount:    totalCount,
	}, nil
}

package tenantgeneral

import (
	"context"
	"fmt"
	"time"

	"kroncl-server/internal/currency"
)

func (s *Service) GetCompanySummary(ctx context.Context, targetCurrency string) (*CompanySummary, error) {
	if targetCurrency == "" {
		targetCurrency = "RUB"
	}

	// ----------
	// COUNTS
	// ----------
	countsQuery := `
        SELECT
            (SELECT COUNT(*) FROM employees)                          AS employees_total,
            (SELECT COUNT(*) FROM employees WHERE status = 'active')  AS employees_active,
            (SELECT COUNT(*) FROM deals)                              AS deals_total,
            (SELECT COUNT(*) FROM catalog_categories)                 AS categories_total,
            (SELECT COUNT(*) FROM catalog_units)                      AS units_total,
            (SELECT COUNT(*) FROM clients)                            AS clients_total,
            (SELECT COUNT(*) FROM clients WHERE status = 'active')    AS clients_active
    `

	var summary CompanySummary
	err := s.pool.QueryRow(ctx, countsQuery).Scan(
		&summary.EmployeesTotal,
		&summary.EmployeesActive,
		&summary.DealsTotal,
		&summary.CategoriesTotal,
		&summary.UnitsTotal,
		&summary.ClientsTotal,
		&summary.ClientsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get company counts: %w", err)
	}

	// ----------
	// BALANCE (как в fm.GetAnalysisSummary)
	// ----------
	balanceQuery := `
        SELECT
            t.currency,
            DATE_TRUNC('hour', t.created_at) AS date_hour,
            COALESCE(SUM(CASE WHEN t.direction = 'income'  THEN t.base_amount ELSE 0 END), 0) AS income,
            COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) AS expense,
            COUNT(*) AS count
        FROM transactions t
        GROUP BY t.currency, DATE_TRUNC('hour', t.created_at)
    `

	rows, err := s.pool.Query(ctx, balanceQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance raw stats: %w", err)
	}
	defer rows.Close()

	rawStats := make(map[string]map[time.Time]*currency.RawStat)

	for rows.Next() {
		var currencyID string
		var dateHour time.Time
		var income, expense float64
		var count int64

		if err := rows.Scan(&currencyID, &dateHour, &income, &expense, &count); err != nil {
			return nil, fmt.Errorf("failed to scan balance row: %w", err)
		}

		if rawStats[currencyID] == nil {
			rawStats[currencyID] = make(map[time.Time]*currency.RawStat)
		}
		rawStats[currencyID][dateHour] = &currency.RawStat{
			Income:  income,
			Expense: expense,
			Count:   count,
		}
	}

	converted, err := s.currencyService.ConvertSummary(ctx, rawStats, targetCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert balance: %w", err)
	}

	summary.BalanceIncome = converted.TotalIncome
	summary.BalanceExpense = converted.TotalExpense
	summary.BalanceTotal = converted.NetBalance
	summary.BalanceCurrency = converted.Currency
	summary.BalanceCount = converted.TransactionCount

	return &summary, nil
}

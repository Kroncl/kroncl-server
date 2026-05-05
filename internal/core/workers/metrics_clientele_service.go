package coreworkers

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) CollectClienteleMetrics(ctx context.Context) (*MetricsClienteleSnapshot, error) {
	stats := &MetricsClienteleSnapshot{
		RecordedAt: time.Now(),
	}

	// Получаем метрики из сервиса аккаунтов
	accountsMetrics, err := s.accountsService.GetAccountsMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts metrics: %w", err)
	}

	stats.TotalAccounts = accountsMetrics.TotalAccounts
	stats.ConfirmedAccounts = accountsMetrics.ConfirmedAccounts
	stats.WaitingAccounts = accountsMetrics.WaitingAccounts
	stats.AdminAccounts = accountsMetrics.AdminAccounts

	stats.AccountTypeOwner = accountsMetrics.AccountsByType["owner"]
	stats.AccountTypeEmployee = accountsMetrics.AccountsByType["employee"]
	stats.AccountTypeAdmin = accountsMetrics.AccountsByType["admin"]
	stats.AccountTypeOutsourcing = accountsMetrics.AccountsByType["outsourcing"]
	stats.AccountTypeTech = accountsMetrics.AccountsByType["tech"]

	// Получаем метрики из сервиса компаний
	companiesMetrics, err := s.companiesService.GetCompaniesMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get companies metrics: %w", err)
	}

	stats.TotalCompanies = companiesMetrics.TotalCompanies
	stats.PublicCompanies = companiesMetrics.PublicCompanies
	stats.PrivateCompanies = companiesMetrics.PrivateCompanies
	stats.TotalCompanyAccounts = companiesMetrics.TotalCompanyAccounts
	stats.AvgAccountsPerCompany = companiesMetrics.AvgAccountsPerCompany
	stats.MaxAccountsInCompany = companiesMetrics.MaxAccountsInCompany
	stats.ActiveCompanies7d = companiesMetrics.ActiveCompanies7d
	stats.ActiveCompanies30d = companiesMetrics.ActiveCompanies30d

	// Получаем метрики из сервиса платежей
	pricingMetrics, err := s.pricingService.GetPricingMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing metrics: %w", err)
	}

	stats.TotalTransactions = pricingMetrics.TotalTransactions
	stats.SuccessTransactions = pricingMetrics.SuccessTransactions
	stats.PendingTransactions = pricingMetrics.PendingTransactions
	stats.TrialTransactions = pricingMetrics.TrialTransactions

	// Схемы без данных (созданы но пустые)
	query := `
		SELECT COUNT(DISTINCT nspname)
		FROM pg_namespace
		WHERE nspname LIKE 'company_%'
		AND NOT EXISTS (
			SELECT 1 FROM pg_tables WHERE schemaname = nspname LIMIT 1
		)
	`
	err = s.pool.QueryRow(ctx, query).Scan(&stats.CompanySchemasWithoutData)
	if err != nil {
		stats.CompanySchemasWithoutData = 0
	}

	return stats, nil
}

func (s *Service) SaveClienteleMetricsSnapshot(ctx context.Context, stats *MetricsClienteleSnapshot) error {
	query := `
		INSERT INTO metrics_clientele_history (
			recorded_at,
			total_accounts, confirmed_accounts, waiting_accounts, admin_accounts,
			account_type_owner, account_type_employee, account_type_admin,
			account_type_outsourcing, account_type_tech,
			total_companies, public_companies, private_companies,
			total_company_accounts, avg_accounts_per_company, max_accounts_in_company,
			total_transactions, success_transactions, pending_transactions, trial_transactions,
			active_companies_7d, active_companies_30d,
			company_schemas_without_data
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20,
			$21, $22, $23
		)
	`

	_, err := s.pool.Exec(ctx, query,
		stats.RecordedAt,
		stats.TotalAccounts, stats.ConfirmedAccounts, stats.WaitingAccounts, stats.AdminAccounts,
		stats.AccountTypeOwner, stats.AccountTypeEmployee, stats.AccountTypeAdmin,
		stats.AccountTypeOutsourcing, stats.AccountTypeTech,
		stats.TotalCompanies, stats.PublicCompanies, stats.PrivateCompanies,
		stats.TotalCompanyAccounts, stats.AvgAccountsPerCompany, stats.MaxAccountsInCompany,
		stats.TotalTransactions, stats.SuccessTransactions, stats.PendingTransactions, stats.TrialTransactions,
		stats.ActiveCompanies7d, stats.ActiveCompanies30d,
		stats.CompanySchemasWithoutData,
	)

	if err != nil {
		return fmt.Errorf("failed to save clientele metrics snapshot: %w", err)
	}

	return nil
}

func (s *Service) GetClienteleMetricsHistory(ctx context.Context, startDate, endDate *time.Time, limit int) ([]MetricsClienteleSnapshot, error) {
	query := `
		SELECT 
			recorded_at,
			total_accounts, confirmed_accounts, waiting_accounts, admin_accounts,
			account_type_owner, account_type_employee, account_type_admin,
			account_type_outsourcing, account_type_tech,
			total_companies, public_companies, private_companies,
			total_company_accounts, avg_accounts_per_company, max_accounts_in_company,
			total_transactions, success_transactions, pending_transactions, trial_transactions,
			active_companies_7d, active_companies_30d,
			company_schemas_without_data
		FROM metrics_clientele_history
		WHERE 1=1
	`

	args := []interface{}{}
	argCounter := 1

	if startDate != nil {
		query += fmt.Sprintf(" AND recorded_at >= $%d", argCounter)
		args = append(args, *startDate)
		argCounter++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argCounter)
		args = append(args, *endDate)
		argCounter++
	}

	query += " ORDER BY recorded_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCounter)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get clientele metrics history: %w", err)
	}
	defer rows.Close()

	var metrics []MetricsClienteleSnapshot
	for rows.Next() {
		var m MetricsClienteleSnapshot
		err := rows.Scan(
			&m.RecordedAt,
			&m.TotalAccounts, &m.ConfirmedAccounts, &m.WaitingAccounts, &m.AdminAccounts,
			&m.AccountTypeOwner, &m.AccountTypeEmployee, &m.AccountTypeAdmin,
			&m.AccountTypeOutsourcing, &m.AccountTypeTech,
			&m.TotalCompanies, &m.PublicCompanies, &m.PrivateCompanies,
			&m.TotalCompanyAccounts, &m.AvgAccountsPerCompany, &m.MaxAccountsInCompany,
			&m.TotalTransactions, &m.SuccessTransactions, &m.PendingTransactions, &m.TrialTransactions,
			&m.ActiveCompanies7d, &m.ActiveCompanies30d,
			&m.CompanySchemasWithoutData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clientele metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

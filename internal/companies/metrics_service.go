package companies

import (
	"context"
	"fmt"
)

type CompaniesMetrics struct {
	TotalCompanies        int     `json:"total_companies"`
	PublicCompanies       int     `json:"public_companies"`
	PrivateCompanies      int     `json:"private_companies"`
	TotalCompanyAccounts  int     `json:"total_company_accounts"`
	AvgAccountsPerCompany float64 `json:"avg_accounts_per_company"`
	MaxAccountsInCompany  int     `json:"max_accounts_in_company"`
	ActiveCompanies7d     int     `json:"active_companies_7d"`
	ActiveCompanies30d    int     `json:"active_companies_30d"`
}

func (s *Service) GetCompaniesMetrics(ctx context.Context) (*CompaniesMetrics, error) {
	var metrics CompaniesMetrics

	// Основная статистика по компаниям и связям
	query := `
        SELECT 
            COUNT(DISTINCT c.id) as total_companies,
            COUNT(CASE WHEN c.is_public = true THEN 1 END) as public_companies,
            COUNT(CASE WHEN c.is_public = false THEN 1 END) as private_companies,
            COUNT(ca.account_id) as total_company_accounts,
            COALESCE(ROUND(AVG(acc_count)::numeric, 2), 0) as avg_accounts_per_company,
            COALESCE(MAX(acc_count), 0) as max_accounts_in_company
        FROM companies c
        LEFT JOIN company_accounts ca ON c.id = ca.company_id
        LEFT JOIN (
            SELECT company_id, COUNT(*) as acc_count
            FROM company_accounts
            GROUP BY company_id
        ) acc_stats ON c.id = acc_stats.company_id
    `

	err := s.pool.QueryRow(ctx, query).Scan(
		&metrics.TotalCompanies,
		&metrics.PublicCompanies,
		&metrics.PrivateCompanies,
		&metrics.TotalCompanyAccounts,
		&metrics.AvgAccountsPerCompany,
		&metrics.MaxAccountsInCompany,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get companies metrics: %w", err)
	}

	// Активные компании за 7 дней
	query7d := `
        SELECT COUNT(DISTINCT ca.company_id)
        FROM company_accounts ca
        JOIN accounts a ON ca.account_id = a.id
        WHERE a.updated_at >= NOW() - INTERVAL '7 days'
    `
	err = s.pool.QueryRow(ctx, query7d).Scan(&metrics.ActiveCompanies7d)
	if err != nil {
		metrics.ActiveCompanies7d = 0
	}

	// Активные компании за 30 дней
	query30d := `
        SELECT COUNT(DISTINCT ca.company_id)
        FROM company_accounts ca
        JOIN accounts a ON ca.account_id = a.id
        WHERE a.updated_at >= NOW() - INTERVAL '30 days'
    `
	err = s.pool.QueryRow(ctx, query30d).Scan(&metrics.ActiveCompanies30d)
	if err != nil {
		metrics.ActiveCompanies30d = 0
	}

	return &metrics, nil
}

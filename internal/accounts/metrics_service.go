package accounts

import (
	"context"
	"fmt"
)

type AccountsMetrics struct {
	TotalAccounts     int            `json:"total_accounts"`
	ConfirmedAccounts int            `json:"confirmed_accounts"`
	WaitingAccounts   int            `json:"waiting_accounts"`
	AdminAccounts     int            `json:"admin_accounts"`
	AccountsByType    map[string]int `json:"accounts_by_type"`
}

func (s *Service) GetAccountsMetrics(ctx context.Context) (*AccountsMetrics, error) {
	metrics := &AccountsMetrics{
		AccountsByType: make(map[string]int),
	}

	query := `
        SELECT 
            COUNT(*) as total_accounts,
            COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as confirmed_accounts,
            COUNT(CASE WHEN status = 'waiting' THEN 1 END) as waiting_accounts,
            COUNT(CASE WHEN adm.account_id IS NOT NULL THEN 1 END) as admin_accounts
        FROM accounts a
        LEFT JOIN admins adm ON a.id = adm.account_id
    `

	err := s.pool.QueryRow(ctx, query).Scan(
		&metrics.TotalAccounts,
		&metrics.ConfirmedAccounts,
		&metrics.WaitingAccounts,
		&metrics.AdminAccounts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts metrics: %w", err)
	}

	// Статистика по типам аккаунтов
	typeQuery := `
        SELECT 
            COALESCE(type, 'unknown') as account_type,
            COUNT(*) as count
        FROM accounts
        GROUP BY account_type
    `

	rows, err := s.pool.Query(ctx, typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var accType string
		var count int
		if err := rows.Scan(&accType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan account type: %w", err)
		}
		metrics.AccountsByType[accType] = count
	}

	return metrics, nil
}

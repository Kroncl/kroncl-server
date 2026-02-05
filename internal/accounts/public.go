package accounts

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func scanAccountPublic(row pgx.Row) (*AccountPublic, error) {
	var account AccountPublic
	err := row.Scan(
		&account.ID,
		&account.Name,
		&account.Email,
		&account.Status,
		&account.AvatarURL,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetPublicByID возвращает один AccountPublic по ID
func (s *Service) GetPublicByID(ctx context.Context, accountID string) (*AccountPublic, error) {
	query := `
        SELECT 
            id, name, email, status,
            COALESCE(avatar_url, '') as avatar_url,
            created_at
        FROM accounts 
        WHERE id = $1
    `

	row := s.pool.QueryRow(ctx, query, accountID)
	return scanAccountPublic(row)
}

// GetPublicAccountsByIDs возвращает коллекцию AccountPublic по массиву ID
func (s *Service) GetPublicAccountsByIDs(ctx context.Context, accountIDs []string) (map[string]AccountPublic, error) {
	if len(accountIDs) == 0 {
		return make(map[string]AccountPublic), nil
	}

	// Создаем placeholders для IN запроса
	placeholders := make([]string, len(accountIDs))
	args := make([]interface{}, len(accountIDs))
	for i, id := range accountIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
        SELECT 
            id, name, email, status,
            COALESCE(avatar_url, '') as avatar_url,
            created_at
        FROM accounts 
        WHERE id IN (%s)
    `, strings.Join(placeholders, ", "))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	accounts := make(map[string]AccountPublic)
	for rows.Next() {
		account, err := scanAccountPublic(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts[account.ID] = *account
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return accounts, nil
}

// GetPublicBatch возвращает слайс AccountPublic (удобно для JSON ответов)
func (s *Service) GetPublicBatch(ctx context.Context, accountIDs []string) ([]AccountPublic, error) {
	accountsMap, err := s.GetPublicAccountsByIDs(ctx, accountIDs)
	if err != nil {
		return nil, err
	}

	// Сохраняем порядок из запроса
	result := make([]AccountPublic, 0, len(accountsMap))
	for _, id := range accountIDs {
		if account, ok := accountsMap[id]; ok {
			result = append(result, account)
		}
	}

	return result, nil
}

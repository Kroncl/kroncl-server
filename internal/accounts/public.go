package accounts

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
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

// GetPublicAccounts возвращает список AccountPublic с пагинацией и поиском
// Показывает только аккаунты со статусом 'confirmed'
func (s *Service) GetPublicAccounts(
	ctx context.Context,
	search string,
	params core.PaginationParams,
) ([]AccountPublic, core.Pagination, error) {
	// Базовый запрос - только подтвержденные аккаунты
	baseQuery := `
        SELECT 
            id, name, email, status,
            COALESCE(avatar_url, '') as avatar_url,
            created_at
        FROM accounts
        WHERE status = 'confirmed'
    `

	// Запрос для подсчета общего количества
	countQuery := `
        SELECT COUNT(*) 
        FROM accounts
        WHERE status = 'confirmed'
    `

	// Подготавливаем аргументы
	var args []interface{}
	var argCounter = 1

	// Добавляем условия поиска если есть
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"

		whereCondition := `
            AND (
                LOWER(email) LIKE $` + strconv.Itoa(argCounter) + ` 
                OR LOWER(name) LIKE $` + strconv.Itoa(argCounter) + `
            )
        `

		baseQuery += whereCondition
		countQuery += whereCondition
		args = append(args, searchPattern)
		argCounter++
	}

	// Получаем общее количество для пагинации
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to count accounts: %w", err)
	}

	// Добавляем сортировку и лимиты для основного запроса
	baseQuery += " ORDER BY created_at DESC"
	baseQuery += " LIMIT $" + strconv.Itoa(argCounter) + " OFFSET $" + strconv.Itoa(argCounter+1)
	args = append(args, params.Limit, params.Offset)

	// Выполняем основной запрос
	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	// Собираем результаты
	var accounts []AccountPublic
	for rows.Next() {
		account, err := scanAccountPublic(rows)
		if err != nil {
			return nil, core.Pagination{}, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, *account)
	}

	if err := rows.Err(); err != nil {
		return nil, core.Pagination{}, fmt.Errorf("rows iteration error: %w", err)
	}

	// Создаем пагинацию
	pagination := core.NewPagination(total, params.Page, params.Limit)

	return accounts, pagination, nil
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

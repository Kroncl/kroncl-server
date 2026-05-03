package adminaccounts

import (
	"context"
	"fmt"
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"strconv"
	"strings"
)

func (s *Service) GetAllAccounts(ctx context.Context, search string, params core.PaginationParams) ([]accounts.Account, core.Pagination, error) {
	baseQuery := `
		SELECT 
			a.id, a.email, a.name, a.auth_type, a.status, 
			a.created_at, a.updated_at, 
			COALESCE(a.avatar_url, '') as avatar_url,
			COALESCE(a.description, '') as description,
			COALESCE(a.type, '') as type,
			COALESCE(adm.level, 0) as admin_level,
			adm.level IS NOT NULL as is_admin
		FROM accounts a
		LEFT JOIN admins adm ON a.id = adm.account_id
	`

	countQuery := `
		SELECT COUNT(*)
		FROM accounts a
	`

	var args []interface{}
	var argCounter = 1

	// WHERE conditions
	var whereClauses []string

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(a.email) LIKE $%d OR LOWER(a.name) LIKE $%d)", argCounter, argCounter))
		args = append(args, searchPattern)
		argCounter++
	}

	if len(whereClauses) > 0 {
		where := " WHERE " + strings.Join(whereClauses, " AND ")
		baseQuery += where
		countQuery += where
	}

	// Count total
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to count accounts: %w", err)
	}

	// Main query with pagination
	baseQuery += " ORDER BY a.created_at DESC"
	baseQuery += " LIMIT $" + strconv.Itoa(argCounter) + " OFFSET $" + strconv.Itoa(argCounter+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accountsList []accounts.Account
	for rows.Next() {
		var acc accounts.Account
		var adminLevel int
		var isAdmin bool

		err := rows.Scan(
			&acc.ID,
			&acc.Email,
			&acc.Name,
			&acc.AuthType,
			&acc.Status,
			&acc.CreatedAt,
			&acc.UpdatedAt,
			&acc.AvatarURL,
			&acc.Description,
			&acc.Type,
			&adminLevel,
			&isAdmin,
		)
		if err != nil {
			return nil, core.Pagination{}, fmt.Errorf("failed to scan account: %w", err)
		}

		acc.IsAdmin = isAdmin
		acc.AdminLevel = adminLevel

		accountsList = append(accountsList, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, core.Pagination{}, fmt.Errorf("rows iteration error: %w", err)
	}

	pagination := core.NewPagination(total, params.Page, params.Limit)

	return accountsList, pagination, nil
}

func (s *Service) GetUserStats(ctx context.Context) (*UserStats, error) {
	stats := &UserStats{
		AccountsWithType: make(map[string]int),
	}

	// Общая статистика по аккаунтам
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as confirmed,
			COUNT(CASE WHEN status = 'waiting' THEN 1 END) as waiting,
			COUNT(CASE WHEN adm.account_id IS NOT NULL THEN 1 END) as admin_accounts
		FROM accounts a
		LEFT JOIN admins adm ON a.id = adm.account_id
	`

	err := s.pool.QueryRow(ctx, query).Scan(
		&stats.TotalAccounts,
		&stats.ConfirmedAccounts,
		&stats.WaitingAccounts,
		&stats.AdminAccounts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
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
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var accType string
		var count int
		if err := rows.Scan(&accType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type stats: %w", err)
		}
		stats.AccountsWithType[accType] = count
	}

	return stats, nil
}

// adminaccounts/service.go

func (s *Service) PromoteToAdmin(ctx context.Context, callerAccountID, targetAccountID string, level int) error {
	// Проверяем, что не пытаемся назначить себя
	if callerAccountID == targetAccountID {
		return fmt.Errorf("cannot promote yourself")
	}

	// Получаем уровень вызвавшего админа
	callerLevel, err := s.adminAuthService.GetAdminLevel(ctx, callerAccountID)
	if err != nil {
		return fmt.Errorf("failed to get caller admin level: %w", err)
	}

	if callerLevel == 0 {
		return fmt.Errorf("caller is not an admin")
	}

	// Получаем текущий уровень целевого аккаунта (если есть)
	targetLevel, err := s.adminAuthService.GetAdminLevel(ctx, targetAccountID)
	if err != nil && err.Error() != "account not found" {
		return fmt.Errorf("failed to get target admin level: %w", err)
	}

	// Уровень вызвавшего должен быть строго больше уровня целевого
	if targetLevel > 0 && callerLevel <= targetLevel {
		return fmt.Errorf("cannot promote account with equal or higher admin level")
	}

	// Если уровень целевого выше или равен - запрещаем
	if callerLevel <= level {
		return fmt.Errorf("cannot set admin level %d, your level is %d", level, callerLevel)
	}

	// Проверяем уровень в допустимых пределах
	if level < config.ADMIN_LEVEL_MIN || level > config.ADMIN_LEVEL_MAX {
		return fmt.Errorf("level must be between %d and %d", config.ADMIN_LEVEL_MIN, config.ADMIN_LEVEL_MAX)
	}

	return s.adminAuthService.PromoteToAdmin(ctx, targetAccountID, level)
}

func (s *Service) DemoteFromAdmin(ctx context.Context, callerAccountID, targetAccountID string) error {
	// Проверяем, что не пытаемся разжаловать себя
	if callerAccountID == targetAccountID {
		return fmt.Errorf("cannot demote yourself")
	}

	// Получаем уровень вызвавшего админа
	callerLevel, err := s.adminAuthService.GetAdminLevel(ctx, callerAccountID)
	if err != nil {
		return fmt.Errorf("failed to get caller admin level: %w", err)
	}

	if callerLevel == 0 {
		return fmt.Errorf("caller is not an admin")
	}

	// Получаем уровень целевого аккаунта
	targetLevel, err := s.adminAuthService.GetAdminLevel(ctx, targetAccountID)
	if err != nil {
		return fmt.Errorf("failed to get target admin level: %w", err)
	}

	if targetLevel == 0 {
		return fmt.Errorf("target account is not an admin")
	}

	// Уровень вызвавшего должен быть строго больше уровня целевого
	if callerLevel <= targetLevel {
		return fmt.Errorf("cannot demote account with equal or higher admin level")
	}

	return s.adminAuthService.DemoteFromAdmin(ctx, targetAccountID)
}

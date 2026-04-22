package accounts

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *Service) GetByEmail(ctx context.Context, email string) (*Account, error) {
	query := `
		SELECT 
			id, email, name, auth_type, status, 
			created_at, updated_at, 
			COALESCE(avatar_url, '') as avatar_url,
			COALESCE(description, '') as description,
			COALESCE(type, '') as type
		FROM accounts 
		WHERE email = $1
	`

	var account Account
	err := s.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&account.ID,
		&account.Email,
		&account.Name,
		&account.AuthType,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.AvatarURL,
		&account.Description,
		&account.Type,
	)

	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	return &account, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Account, error) {
	query := `
		SELECT 
			id, email, name, auth_type, status, 
			created_at, updated_at, 
			COALESCE(avatar_url, '') as avatar_url,
			COALESCE(description, '') as description,
			COALESCE(type, '') as type
		FROM accounts 
		WHERE id = $1
	`

	var account Account
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.Email,
		&account.Name,
		&account.AuthType,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.AvatarURL,
		&account.Description,
		&account.Type,
	)

	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	return &account, nil
}

func (s *Service) UpdateById(ctx context.Context, accountID string, req *UpdateRequest) (*Account, error) {
	updater := core.NewUpdater("accounts")

	if req.Name != nil && *req.Name != "" {
		if err := s.validateName(*req.Name); err != nil {
			return nil, err
		}
		updater.SetString("name", *req.Name)
	}
	if req.AvatarUrl != nil {
		updater.SetString("avatar_url", *req.AvatarUrl)
	}
	if req.Description != nil {
		if *req.Description == "" {
			updater.SetNull("description")
		} else {
			updater.SetString("description", *req.Description)
		}
	}
	if req.Type != nil {
		if !validAccountTypes[*req.Type] {
			return nil, fmt.Errorf("invalid account type: %s, valid types: owner, employee, admin, outsourcing, tech", *req.Type)
		}
		updater.SetString("type", *req.Type)
	}

	updater.Where("id = $1", accountID)

	query, args := updater.Build()
	if query == "" {
		return s.GetByID(ctx, accountID)
	}

	_, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return s.GetByID(ctx, accountID)
}

func (s *Service) GetPublicAccounts(
	ctx context.Context,
	search string,
	params core.PaginationParams,
) ([]AccountPublic, core.Pagination, error) {
	baseQuery := `
        SELECT 
            id, name, email, status,
            COALESCE(avatar_url, '') as avatar_url,
            COALESCE(description, '') as description,
            COALESCE(type, '') as type,
            created_at
        FROM accounts
        WHERE status = 'confirmed'
    `

	countQuery := `
        SELECT COUNT(*) 
        FROM accounts
        WHERE status = 'confirmed'
    `

	var args []interface{}
	var argCounter = 1

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

	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to count accounts: %w", err)
	}

	baseQuery += " ORDER BY created_at DESC"
	baseQuery += " LIMIT $" + strconv.Itoa(argCounter) + " OFFSET $" + strconv.Itoa(argCounter+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

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

	pagination := core.NewPagination(total, params.Page, params.Limit)

	return accounts, pagination, nil
}

func (s *Service) GetPublicByID(ctx context.Context, accountID string) (*AccountPublic, error) {
	query := `
        SELECT 
            id, name, email, status,
            COALESCE(avatar_url, '') as avatar_url,
            COALESCE(description, '') as description,
            COALESCE(type, '') as type,
            created_at
        FROM accounts 
        WHERE id = $1
    `

	row := s.pool.QueryRow(ctx, query, accountID)
	return scanAccountPublic(row)
}

func (s *Service) GetPublicAccountsByIDs(ctx context.Context, accountIDs []string) (map[string]AccountPublic, error) {
	if len(accountIDs) == 0 {
		return make(map[string]AccountPublic), nil
	}

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
            COALESCE(description, '') as description,
            COALESCE(type, '') as type,
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

func (s *Service) GetPublicBatch(ctx context.Context, accountIDs []string) ([]AccountPublic, error) {
	accountsMap, err := s.GetPublicAccountsByIDs(ctx, accountIDs)
	if err != nil {
		return nil, err
	}

	result := make([]AccountPublic, 0, len(accountsMap))
	for _, id := range accountIDs {
		if account, ok := accountsMap[id]; ok {
			result = append(result, account)
		}
	}

	return result, nil
}

func scanAccountPublic(row pgx.Row) (*AccountPublic, error) {
	var account AccountPublic
	err := row.Scan(
		&account.ID,
		&account.Name,
		&account.Email,
		&account.Status,
		&account.AvatarURL,
		&account.Description,
		&account.Type,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strings"
)

// RemoveCompanyMember удаляет участника из компании (если он не владелец)
func (s *Service) RemoveCompanyMember(ctx context.Context, companyID, memberID string) error {
	var isOwner bool
	checkOwnerQuery := `
		SELECT EXISTS(
			SELECT 1 
			FROM company_accounts ca
			INNER JOIN roles r ON ca.role_id = r.id
			WHERE ca.company_id = $1 
			AND ca.account_id = $2 
			AND r.code = 'owner'
		)
	`

	err := s.pool.QueryRow(ctx, checkOwnerQuery, companyID, memberID).Scan(&isOwner)
	if err != nil {
		return fmt.Errorf("failed to check if member is owner: %w", err)
	}

	if isOwner {
		return fmt.Errorf("cannot remove owner from company")
	}

	// Удаляем участника из компании
	deleteQuery := `
		DELETE FROM company_accounts 
		WHERE company_id = $1 AND account_id = $2
	`

	result, err := s.pool.Exec(ctx, deleteQuery, companyID, memberID)
	if err != nil {
		return fmt.Errorf("failed to remove member from company: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		// Если удалили 0 строк, возможно участника нет в компании
		// Проверяем существует ли участник в компании
		var exists bool
		checkExistsQuery := `
			SELECT EXISTS(
				SELECT 1 
				FROM company_accounts 
				WHERE company_id = $1 AND account_id = $2
			)
		`

		err = s.pool.QueryRow(ctx, checkExistsQuery, companyID, memberID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check if member exists: %w", err)
		}

		if !exists {
			return fmt.Errorf("member not found in company")
		}

		// Если участник существует, но почему-то не удалился
		return fmt.Errorf("failed to remove member")
	}

	return nil
}

// RemoveCompanyMemberSimple еще более простая версия - только проверка и удаление
func (s *Service) RemoveCompanyMemberSimple(ctx context.Context, companyID, memberID string) error {
	// Одним запросом проверяем и удаляем если не владелец
	query := `
		WITH delete_check AS (
			SELECT ca.company_id, ca.account_id, r.code as role_code
			FROM company_accounts ca
			INNER JOIN roles r ON ca.role_id = r.id
			WHERE ca.company_id = $1 AND ca.account_id = $2
		)
		DELETE FROM company_accounts ca
		USING delete_check dc
		WHERE ca.company_id = dc.company_id 
		AND ca.account_id = dc.account_id
		AND dc.role_code != 'owner'
		RETURNING ca.account_id
	`

	var deletedID string
	err := s.pool.QueryRow(ctx, query, companyID, memberID).Scan(&deletedID)
	if err != nil {
		// Если ничего не вернулось - либо участника нет, либо он владелец

		// Проверяем существует ли участник
		var exists bool
		checkExistsQuery := `
			SELECT EXISTS(
				SELECT 1 
				FROM company_accounts ca
				INNER JOIN roles r ON ca.role_id = r.id
				WHERE ca.company_id = $1 AND ca.account_id = $2
			)
		`

		err = s.pool.QueryRow(ctx, checkExistsQuery, companyID, memberID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check member: %w", err)
		}

		if !exists {
			return fmt.Errorf("member not found in company")
		}

		// Если участник существует, значит он владелец
		return fmt.Errorf("cannot remove owner from company")
	}

	return nil
}

// GetCompanyMember возвращает информацию об одном участнике компании
func (s *Service) GetCompanyMember(ctx context.Context, companyID, memberID string) (*CompanyPublicMember, error) {
	query := `
		SELECT 
			a.id,
			a.name,
			a.email,
			a.status,
			a.avatar_url,
			a.created_at,
			ca.role_id,
			r.code as role_code,
			r.name as role_name,
			r.description as role_description,
			ca.created_at as joined_at
		FROM company_accounts ca
		INNER JOIN accounts a ON ca.account_id = a.id
		INNER JOIN roles r ON ca.role_id = r.id
		WHERE ca.company_id = $1 AND ca.account_id = $2
	`

	var member CompanyPublicMember
	var roleDescription *string

	err := s.pool.QueryRow(ctx, query, companyID, memberID).Scan(
		&member.ID,
		&member.Name,
		&member.Email,
		&member.Status,
		&member.AvatarURL,
		&member.CreatedAt,
		&member.RoleID,
		&member.RoleCode,
		&member.RoleName,
		&roleDescription,
		&member.JoinedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company member: %w", err)
	}

	return &member, nil
}

func (s *Service) GetCompanyMembers(ctx context.Context, companyID string, req *GetCompanyMembersRequest) (*GetCompanyMembersResponse, error) {
	// Валидация параметров
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// Строим SQL запрос
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{companyID}
	argIndex := 2

	// Базовый запрос для счетчика
	countBuilder.WriteString(`
		SELECT COUNT(*) 
		FROM company_accounts ca
		INNER JOIN accounts a ON ca.account_id = a.id
		INNER JOIN roles r ON ca.role_id = r.id
		WHERE ca.company_id = $1
	`)

	// Базовый запрос для данных
	queryBuilder.WriteString(`
		SELECT 
			a.id,
			a.name,
			a.email,
			a.status,
			a.avatar_url,
			a.created_at,
			ca.role_id,
			r.code as role_code,
			r.name as role_name,
			r.description as role_description,
			ca.created_at as joined_at
		FROM company_accounts ca
		INNER JOIN accounts a ON ca.account_id = a.id
		INNER JOIN roles r ON ca.role_id = r.id
		WHERE ca.company_id = $1
	`)

	// Добавляем поиск если есть
	if req.Search != "" {
		searchClause := fmt.Sprintf(" AND (a.name ILIKE $%d OR a.email ILIKE $%d)", argIndex, argIndex+1)
		queryBuilder.WriteString(searchClause)
		countBuilder.WriteString(searchClause)

		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// Добавляем фильтр по роли если есть
	if req.Role != "" && req.Role != "all" {
		roleClause := fmt.Sprintf(" AND r.code = $%d", argIndex)
		queryBuilder.WriteString(roleClause)
		countBuilder.WriteString(roleClause)

		args = append(args, req.Role)
		argIndex++
	}

	// Добавляем сортировку
	sortField := "a.name"
	if req.SortBy == "joined_at" {
		sortField = "ca.created_at"
	} else if req.SortBy == "role" {
		sortField = `
			CASE r.code 
				WHEN 'owner' THEN 1
				WHEN 'admin' THEN 2
				WHEN 'member' THEN 3
				WHEN 'guest' THEN 4
				ELSE 5
			END
		`
	}

	sortOrder := "ASC"
	if req.SortOrder == "desc" {
		sortOrder = "DESC"
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder))

	// Выполняем запрос на получение количества
	var total int
	err := s.pool.QueryRow(ctx, countBuilder.String(), args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count company members: %w", err)
	}

	// Добавляем пагинацию
	offset := (req.Page - 1) * req.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, req.Limit, offset)

	// Выполняем запрос на получение данных
	rows, err := s.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get company members: %w", err)
	}
	defer rows.Close()

	var members []CompanyPublicMember
	for rows.Next() {
		var member CompanyPublicMember
		var roleDescription *string

		err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.Email,
			&member.Status,
			&member.AvatarURL,
			&member.CreatedAt,
			&member.RoleID,
			&member.RoleCode,
			&member.RoleName,
			&roleDescription,
			&member.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company member: %w", err)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Рассчитываем количество страниц
	pages := total / req.Limit
	if total%req.Limit > 0 {
		pages++
	}

	pagination := &core.Pagination{
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
		Pages: pages,
	}

	return &GetCompanyMembersResponse{
		Members:    members,
		Pagination: *pagination,
	}, nil
}

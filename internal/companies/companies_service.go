package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strings"
	"time"

	"github.com/google/uuid"
)

// обновление компании
func (s *Service) UpdateById(ctx context.Context, companyID string, req *UpdateRequest) (*Company, error) {
	updater := core.NewUpdater("companies")

	if req.Name != nil && *req.Name != "" {
		if err := s.ValidateCompanyName(*req.Name); err != nil {
			return nil, err
		}
		updater.SetString("name", *req.Name)
	}
	if req.Description != nil {
		updater.Set("description", *req.Description)
	}
	if req.AvatarUrl != nil {
		updater.SetString("avatar_url", *req.AvatarUrl)
	}

	updater.Where("id = $1", companyID)

	query, args := updater.Build()
	if query == "" {
		return s.GetCompanyByID(ctx, companyID)
	}

	// Выполняем запрос
	_, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	// Возвращаем обновленную компанию
	return s.GetCompanyByID(ctx, companyID)
}

// получение организации
// без принадлежности к пользователю
func (s *Service) GetCompanyByID(ctx context.Context, companyID string) (*Company, error) {
	query := `
		SELECT id, slug, name, description, avatar_url, is_public,
		       created_at, updated_at
		FROM companies 
		WHERE id = $1
	`

	var company Company
	err := s.pool.QueryRow(ctx, query, companyID).Scan(
		&company.ID,
		&company.Slug,
		&company.Name,
		&company.Description,
		&company.AvatarUrl,
		&company.IsPublic,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}

// получение организации с метками
// принадлежности пользователя
func (s *Service) GetUserCompanyById(ctx context.Context, userID string, companyID string) (*UserCompany, error) {
	query := `
		SELECT 
			c.id, c.slug, c.name, c.description, c.avatar_url, c.is_public,
			c.created_at, c.updated_at,
			ca.role_id, r.code as role_code, r.name as role_name,
			ca.created_at as joined_at
		FROM companies c
		INNER JOIN company_accounts ca ON c.id = ca.company_id
		INNER JOIN roles r ON ca.role_id = r.id
		WHERE c.id = $1 AND ca.account_id = $2
	`

	var company UserCompany
	err := s.pool.QueryRow(ctx, query, companyID, userID).Scan(
		&company.ID,
		&company.Slug,
		&company.Name,
		&company.Description,
		&company.AvatarUrl,
		&company.IsPublic,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.RoleID,
		&company.RoleCode,
		&company.RoleName,
		&company.JoinedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("company not found for user: %w", err)
	}

	return &company, nil
}

// возвращает список организаций пользователя с ролью и пагинацией
func (s *Service) GetUserCompanies(ctx context.Context, userID string, req *GetUserCompaniesRequest) (*GetUserCompaniesResponse, error) {
	// Валидация параметров
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Role == "" {
		req.Role = "all"
	}

	// Проверяем валидность роли
	validRoles := map[string]bool{
		"all":    true,
		"owner":  true,
		"admin":  true,
		"member": true,
		"guest":  true,
	}
	if !validRoles[req.Role] && req.Role != "all" {
		return nil, fmt.Errorf("invalid role filter. Allowed values: all, owner, admin, member, guest")
	}

	// Строим SQL запрос
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{userID}
	argIndex := 2 // начинаем с $2, т.к. $1 = userID

	// Базовый запрос для компаний
	baseQuery := `
		SELECT 
			c.id, c.slug, c.name, c.description, c.avatar_url, c.is_public,
			c.created_at, c.updated_at,
			ca.role_id, r.code as role_code, r.name as role_name,
			ca.created_at as joined_at
		FROM companies c
		INNER JOIN company_accounts ca ON c.id = ca.company_id
		INNER JOIN roles r ON ca.role_id = r.id
		WHERE ca.account_id = $1
	`

	queryBuilder.WriteString(baseQuery)
	countBuilder.WriteString("SELECT COUNT(*) FROM companies c INNER JOIN company_accounts ca ON c.id = ca.company_id INNER JOIN roles r ON ca.role_id = r.id WHERE ca.account_id = $1")

	// Добавляем фильтр по роли
	if req.Role != "all" {
		queryBuilder.WriteString(fmt.Sprintf(" AND r.code = $%d", argIndex))
		countBuilder.WriteString(fmt.Sprintf(" AND r.code = $%d", argIndex))
		args = append(args, req.Role)
		argIndex++
	}

	// Добавляем поиск по названию или slug
	if req.Search != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND (c.name ILIKE $%d OR c.slug ILIKE $%d)", argIndex, argIndex+1))
		countBuilder.WriteString(fmt.Sprintf(" AND (c.name ILIKE $%d OR c.slug ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// Добавляем сортировку
	queryBuilder.WriteString(" ORDER BY c.created_at DESC")

	// Добавляем пагинацию
	offset := (req.Page - 1) * req.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, req.Limit, offset)

	// Выполняем запрос на получение количества
	var total int
	countArgs := args[:len(args)-2] // убираем LIMIT и OFFSET для count запроса
	err := s.pool.QueryRow(ctx, countBuilder.String(), countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count user companies: %w", err)
	}

	// Выполняем запрос на получение данных
	rows, err := s.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user companies: %w", err)
	}
	defer rows.Close()

	var companies []UserCompany
	for rows.Next() {
		var uc UserCompany
		err := rows.Scan(
			&uc.ID,
			&uc.Slug,
			&uc.Name,
			&uc.Description,
			&uc.AvatarUrl,
			&uc.IsPublic,
			&uc.CreatedAt,
			&uc.UpdatedAt,
			&uc.RoleID,
			&uc.RoleCode,
			&uc.RoleName,
			&uc.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, uc)
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
	return &GetUserCompaniesResponse{
		Companies:  companies,
		Pagination: *pagination,
	}, nil
}

func (s *Service) Create(ctx context.Context, ownerId string, slug string, name string, description string, avatarURL string, isPublic bool) (*CreateCompanyResponse, error) {
	// 1. Валидация
	if err := s.ValidateCompanyName(name); err != nil {
		return nil, err
	}

	// 2. Проверка slug (можно в транзакции, но проверяем до нее для раннего фейла)
	isUnique, err := s.checkSlugUnique(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("slug uniqueness check failed: %w", err)
	}
	if !isUnique {
		return nil, fmt.Errorf("company slug isn't unique")
	}

	// 3. Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	currentTime := time.Now()

	// 4. Генерируем UUID для компании
	companyID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate company UUID: %w", err)
	}

	// 5. Создаем компанию
	companyQuery := `
		INSERT INTO companies (
			id, slug, name, description, avatar_url, 
			is_public, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, slug, name, description, avatar_url, 
		          is_public, created_at, updated_at
	`

	var company Company
	err = tx.QueryRow(
		ctx, companyQuery,
		companyID,
		slug,
		name,
		description,
		avatarURL,
		isPublic,
		currentTime,
		currentTime,
	).Scan(
		&company.ID,
		&company.Slug,
		&company.Name,
		&company.Description,
		&company.AvatarUrl,
		&company.IsPublic,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// 6. Получаем ID роли
	var ownerRoleID int
	err = tx.QueryRow(
		ctx,
		`SELECT id FROM roles WHERE code = $1`,
		RoleOwner,
	).Scan(&ownerRoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find owner role: %w", err)
	}

	// 7. Добавляем создателя как владельца в company_accounts
	memberQuery := `
		INSERT INTO company_accounts (
			company_id, account_id, role_id, permissions,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (company_id, account_id) DO NOTHING
	`

	_, err = tx.Exec(
		ctx, memberQuery,
		companyID,
		ownerId,
		ownerRoleID,
		`{}`,
		currentTime,
		currentTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add owner to company: %w", err)
	}

	// 8. Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Обнуляем tx, чтобы defer не откатил
	tx = nil

	// Запускаем процесс создания хранилища
	storage, err := s.storage.InitStorage(ctx, company.ID)
	if err != nil || storage == nil {
		return nil, fmt.Errorf("error init company storage: %w", err)
	}
	companyWithStorage := CreateCompanyResponse{
		Company: company,
		Storage: storage,
	}

	return &companyWithStorage, nil
}

func (s *Service) checkSlugUnique(ctx context.Context, slug string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM companies WHERE slug = $1`

	err := s.pool.QueryRow(ctx, query, strings.ToLower(slug)).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (s *Service) CheckCompanyMembership(ctx context.Context, companyID, userID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM companies c
			JOIN company_accounts ca ON c.id = ca.company_id
			WHERE c.id = $1 AND ca.account_id = $2
		)
	`

	err := s.pool.QueryRow(ctx, query, companyID, userID).Scan(&exists)
	return exists, err
}

// -----------
// MEMBERSHIP
// -----------

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

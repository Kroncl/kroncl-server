package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strings"
)

func (s *Service) GetCompanyVisitCard(ctx context.Context, slug string) (*Company, error) {
	query := `
		SELECT id, slug, name, description, avatar_url, is_public,
		       email, region, site, metadata, created_at, updated_at
		FROM companies 
		WHERE slug = $1 AND is_public = true
	`

	var company Company
	err := s.pool.QueryRow(ctx, query, slug).Scan(
		&company.ID,
		&company.Slug,
		&company.Name,
		&company.Description,
		&company.AvatarUrl,
		&company.IsPublic,
		&company.Email,
		&company.Region,
		&company.Site,
		&company.Metadata,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("company not found or not public: %w", err)
	}

	return &company, nil
}

func (s *Service) UpdateById(ctx context.Context, userID string, companyID string, req *UpdateRequest) (*UserCompany, error) {
	updater := core.NewUpdater("companies")

	if req.Name != nil && *req.Name != "" {
		if err := s.ValidateCompanyName(*req.Name); err != nil {
			return nil, err
		}
		updater.SetString("name", *req.Name)
	}
	if req.Description != nil {
		updater.SetString("description", *req.Description)
	}
	if req.AvatarUrl != nil {
		updater.SetString("avatar_url", *req.AvatarUrl)
	}
	if req.IsPublic != nil {
		updater.SetBool("is_public", *req.IsPublic)
	}
	if req.Region != nil && *req.Region != "" {
		if !IsValidRegion(*req.Region) {
			return nil, fmt.Errorf("invalid region: %s", *req.Region)
		}
		updater.SetString("region", *req.Region)
	}
	if req.Site != nil {
		if *req.Site == "" {
			updater.SetNull("site")
		} else {
			updater.SetString("site", *req.Site)
		}
	}
	if req.Email != nil {
		if *req.Email == "" {
			updater.SetNull("email")
		} else {
			updater.SetString("email", *req.Email)
		}
	}

	updater.Where("id = $1", companyID)

	query, args := updater.Build()
	if query != "" {
		_, err := s.pool.Exec(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update company: %w", err)
		}
	}

	return s.GetUserCompanyById(ctx, userID, companyID)
}

func (s *Service) GetCompanyByID(ctx context.Context, companyID string) (*Company, error) {
	query := `
		SELECT id, slug, name, description, avatar_url, is_public,
		       email, region, site, metadata, created_at, updated_at
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
		&company.Email,
		&company.Region,
		&company.Site,
		&company.Metadata,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}

func (s *Service) GetUserCompanyById(ctx context.Context, userID string, companyID string) (*UserCompany, error) {
	query := `
		SELECT 
			c.id, c.slug, c.name, c.description, c.avatar_url, c.is_public,
			c.email, c.region, c.site, c.metadata, c.created_at, c.updated_at,
			ca.role_code,
			ca.created_at as joined_at
		FROM companies c
		INNER JOIN company_accounts ca ON c.id = ca.company_id
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
		&company.Email,
		&company.Region,
		&company.Site,
		&company.Metadata,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.RoleCode,
		&company.JoinedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("company not found for user: %w", err)
	}

	return &company, nil
}

func (s *Service) GetUserCompanies(ctx context.Context, userID string, req *GetUserCompaniesRequest) (*GetUserCompaniesResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Role == "" {
		req.Role = "all"
	}

	validRoles := map[string]bool{
		"all":   true,
		"owner": true,
		"guest": true,
	}
	if !validRoles[req.Role] && req.Role != "all" {
		return nil, fmt.Errorf("invalid role filter. Allowed values: all, owner, guest")
	}

	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{userID}
	argIndex := 2

	baseQuery := `
		SELECT 
			c.id, c.slug, c.name, c.description, c.avatar_url, c.is_public,
			c.email, c.region, c.site, c.metadata, c.created_at, c.updated_at,
			ca.role_code,
			ca.created_at as joined_at
		FROM companies c
		INNER JOIN company_accounts ca ON c.id = ca.company_id
		WHERE ca.account_id = $1
	`

	queryBuilder.WriteString(baseQuery)
	countBuilder.WriteString("SELECT COUNT(*) FROM companies c INNER JOIN company_accounts ca ON c.id = ca.company_id WHERE ca.account_id = $1")

	if req.Role != "all" {
		queryBuilder.WriteString(fmt.Sprintf(" AND ca.role_code = $%d", argIndex))
		countBuilder.WriteString(fmt.Sprintf(" AND ca.role_code = $%d", argIndex))
		args = append(args, req.Role)
		argIndex++
	}

	if req.Search != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND (c.name ILIKE $%d OR c.slug ILIKE $%d)", argIndex, argIndex+1))
		countBuilder.WriteString(fmt.Sprintf(" AND (c.name ILIKE $%d OR c.slug ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	queryBuilder.WriteString(" ORDER BY c.created_at DESC")

	offset := (req.Page - 1) * req.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, req.Limit, offset)

	var total int
	countArgs := args[:len(args)-2]
	err := s.pool.QueryRow(ctx, countBuilder.String(), countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count user companies: %w", err)
	}

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
			&uc.Email,
			&uc.Region,
			&uc.Site,
			&uc.Metadata,
			&uc.CreatedAt,
			&uc.UpdatedAt,
			&uc.RoleCode,
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

func (s *Service) RemoveCompanyMember(ctx context.Context, companyID, memberID string) error {
	var isOwner bool
	checkOwnerQuery := `
		SELECT EXISTS(
			SELECT 1 
			FROM company_accounts
			WHERE company_id = $1 
			AND account_id = $2 
			AND role_code = 'owner'
		)
	`
	err := s.pool.QueryRow(ctx, checkOwnerQuery, companyID, memberID).Scan(&isOwner)
	if err != nil {
		return fmt.Errorf("failed to check if member is owner: %w", err)
	}

	if isOwner {
		return fmt.Errorf("cannot remove owner from company")
	}

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
		return fmt.Errorf("failed to remove member")
	}

	return nil
}

func (s *Service) GetCompanyMember(ctx context.Context, companyID, memberID string) (*CompanyPublicMember, error) {
	query := `
		SELECT 
			a.id,
			a.name,
			a.email,
			a.status,
			a.avatar_url,
			a.created_at,
			ca.role_code,
			ca.created_at as joined_at
		FROM company_accounts ca
		INNER JOIN accounts a ON ca.account_id = a.id
		WHERE ca.company_id = $1 AND ca.account_id = $2
	`

	var member CompanyPublicMember
	err := s.pool.QueryRow(ctx, query, companyID, memberID).Scan(
		&member.ID,
		&member.Name,
		&member.Email,
		&member.Status,
		&member.AvatarURL,
		&member.CreatedAt,
		&member.RoleCode,
		&member.JoinedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company member: %w", err)
	}

	return &member, nil
}

func (s *Service) GetCompanyMembers(ctx context.Context, companyID string, req *GetCompanyMembersRequest) (*GetCompanyMembersResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{companyID}
	argIndex := 2

	countBuilder.WriteString(`
		SELECT COUNT(*) 
		FROM company_accounts ca
		INNER JOIN accounts a ON ca.account_id = a.id
		WHERE ca.company_id = $1
	`)

	queryBuilder.WriteString(`
		SELECT 
			a.id,
			a.name,
			a.email,
			a.status,
			a.avatar_url,
			a.created_at,
			ca.role_code,
			ca.created_at as joined_at
		FROM company_accounts ca
		INNER JOIN accounts a ON ca.account_id = a.id
		WHERE ca.company_id = $1
	`)

	if req.Search != "" {
		searchClause := fmt.Sprintf(" AND (a.name ILIKE $%d OR a.email ILIKE $%d)", argIndex, argIndex+1)
		queryBuilder.WriteString(searchClause)
		countBuilder.WriteString(searchClause)

		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if req.Role != "" && req.Role != "all" {
		roleClause := fmt.Sprintf(" AND ca.role_code = $%d", argIndex)
		queryBuilder.WriteString(roleClause)
		countBuilder.WriteString(roleClause)
		args = append(args, req.Role)
		argIndex++
	}

	sortField := "a.name"
	if req.SortBy == "joined_at" {
		sortField = "ca.created_at"
	} else if req.SortBy == "role" {
		sortField = `
			CASE ca.role_code 
				WHEN 'owner' THEN 1
				WHEN 'guest' THEN 2
				ELSE 3
			END
		`
	}

	sortOrder := "ASC"
	if req.SortOrder == "desc" {
		sortOrder = "DESC"
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder))

	var total int
	err := s.pool.QueryRow(ctx, countBuilder.String(), args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count company members: %w", err)
	}

	offset := (req.Page - 1) * req.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, req.Limit, offset)

	rows, err := s.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get company members: %w", err)
	}
	defer rows.Close()

	var members []CompanyPublicMember
	for rows.Next() {
		var member CompanyPublicMember
		err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.Email,
			&member.Status,
			&member.AvatarURL,
			&member.CreatedAt,
			&member.RoleCode,
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

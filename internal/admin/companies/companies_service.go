package admincompanies

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strings"
)

func (s *Service) GetAllCompanies(ctx context.Context, search string, params core.PaginationParams) ([]AdminCompany, core.Pagination, error) {
	baseQuery := `
		SELECT 
			c.id, c.slug, c.name, c.description, c.avatar_url, c.is_public,
			c.email, c.region, c.site, c.metadata, c.created_at, c.updated_at,
			COALESCE(cs.status, 'none') as storage_status,
			cs.schema_name
		FROM companies c
		LEFT JOIN company_storage cs ON c.id = cs.company_id
	`

	countQuery := `
		SELECT COUNT(*)
		FROM companies c
	`

	var args []interface{}
	var whereClauses []string
	argCounter := 1

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(c.name) LIKE $%d OR LOWER(c.slug) LIKE $%d)", argCounter, argCounter+1))
		args = append(args, searchPattern, searchPattern)
		argCounter += 2
	}

	if len(whereClauses) > 0 {
		where := " WHERE " + strings.Join(whereClauses, " AND ")
		baseQuery += where
		countQuery += where
	}

	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to count companies: %w", err)
	}

	baseQuery += " ORDER BY c.created_at DESC"
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to query companies: %w", err)
	}
	defer rows.Close()

	var companiesList []AdminCompany
	for rows.Next() {
		var ac AdminCompany
		var metadata []byte
		var email, site *string

		err := rows.Scan(
			&ac.ID,
			&ac.Slug,
			&ac.Name,
			&ac.Description,
			&ac.AvatarUrl,
			&ac.IsPublic,
			&email,
			&ac.Region,
			&site,
			&metadata,
			&ac.CreatedAt,
			&ac.UpdatedAt,
			&ac.StorageStatus,
			&ac.SchemaName,
		)
		if err != nil {
			return nil, core.Pagination{}, fmt.Errorf("failed to scan company: %w", err)
		}

		ac.Email = email
		ac.Site = site
		ac.Metadata = make(map[string]interface{})
		ac.StorageReady = ac.StorageStatus == "active"

		companiesList = append(companiesList, ac)
	}

	pagination := core.NewPagination(total, params.Page, params.Limit)

	return companiesList, pagination, nil
}

func (s *Service) GetCompanyByID(ctx context.Context, companyID string) (*AdminCompany, error) {
	query := `
		SELECT 
			c.id, c.slug, c.name, c.description, c.avatar_url, c.is_public,
			c.email, c.region, c.site, c.metadata, c.created_at, c.updated_at,
			COALESCE(cs.status, 'none') as storage_status,
			cs.schema_name
		FROM companies c
		LEFT JOIN company_storage cs ON c.id = cs.company_id
		WHERE c.id = $1
	`

	var ac AdminCompany
	var metadata []byte
	var email, site *string

	err := s.pool.QueryRow(ctx, query, companyID).Scan(
		&ac.ID,
		&ac.Slug,
		&ac.Name,
		&ac.Description,
		&ac.AvatarUrl,
		&ac.IsPublic,
		&email,
		&ac.Region,
		&site,
		&metadata,
		&ac.CreatedAt,
		&ac.UpdatedAt,
		&ac.StorageStatus,
		&ac.SchemaName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	ac.Email = email
	ac.Site = site
	ac.Metadata = make(map[string]interface{})
	ac.StorageReady = ac.StorageStatus == "active"

	return &ac, nil
}

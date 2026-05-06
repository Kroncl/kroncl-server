package adminpartners

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"kroncl-server/internal/public"
	"strings"
)

func (s *Service) GetAllPartners(ctx context.Context, status *string, search string, params core.PaginationParams) ([]public.IncomingPartner, core.Pagination, error) {
	baseQuery := `
		SELECT id, name, type, text, email, status, created_at, updated_at
		FROM incoming_partners
	`

	countQuery := `SELECT COUNT(*) FROM incoming_partners`

	var args []interface{}
	var whereClauses []string
	argCounter := 1

	if status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, *status)
		argCounter++
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(email) LIKE $%d)", argCounter, argCounter+1))
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
		return nil, core.Pagination{}, fmt.Errorf("failed to count partners: %w", err)
	}

	baseQuery += " ORDER BY created_at DESC"
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := s.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to query partners: %w", err)
	}
	defer rows.Close()

	var partners []public.IncomingPartner
	for rows.Next() {
		var p public.IncomingPartner
		err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.Text, &p.Email, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, core.Pagination{}, fmt.Errorf("failed to scan partner: %w", err)
		}
		partners = append(partners, p)
	}

	pagination := core.NewPagination(total, params.Page, params.Limit)

	return partners, pagination, nil
}

func (s *Service) GetPartnerByID(ctx context.Context, partnerID string) (*public.IncomingPartner, error) {
	query := `
		SELECT id, name, type, text, email, status, created_at, updated_at
		FROM incoming_partners
		WHERE id = $1
	`

	var p public.IncomingPartner
	err := s.pool.QueryRow(ctx, query, partnerID).Scan(
		&p.ID, &p.Name, &p.Type, &p.Text, &p.Email, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	return &p, nil
}

func (s *Service) UpdatePartnerStatus(ctx context.Context, partnerID string, status string) error {
	query := `
		UPDATE incoming_partners
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := s.pool.Exec(ctx, query, status, partnerID)
	if err != nil {
		return fmt.Errorf("failed to update partner status: %w", err)
	}
	return nil
}

func (s *Service) UpdatePartner(ctx context.Context, partnerID string, req *public.UpdateIncomingPartnerRequest) (*public.IncomingPartner, error) {
	updater := core.NewUpdater("incoming_partners")

	if req.Name != nil {
		updater.SetString("name", *req.Name)
	}
	if req.Type != nil {
		updater.SetString("type", *req.Type)
	}
	if req.Text != nil {
		if *req.Text == "" {
			updater.SetNull("text")
		} else {
			updater.SetString("text", *req.Text)
		}
	}
	if req.Email != nil {
		updater.SetString("email", *req.Email)
	}
	if req.Status != nil {
		updater.SetString("status", *req.Status)
	}

	updater.Where("id = $1", partnerID)

	query, args := updater.Build()
	if query == "" {
		return s.GetPartnerByID(ctx, partnerID)
	}

	_, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update partner: %w", err)
	}

	return s.GetPartnerByID(ctx, partnerID)
}

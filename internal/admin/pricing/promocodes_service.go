package adminpricing

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
)

func (s *Service) GetPromocodes(ctx context.Context, params core.PaginationParams) ([]Promocode, core.Pagination, error) {
	query := `
		SELECT 
			p.id, p.code, p.plan_id, pl.name as plan_name, p.trial_period_days, 
			p.created_at, p.updated_at
		FROM pricing_promocodes p
		LEFT JOIN pricing_plans pl ON p.plan_id = pl.code
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	countQuery := `SELECT COUNT(*) FROM pricing_promocodes`

	var total int
	err := s.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to count promocodes: %w", err)
	}

	rows, err := s.pool.Query(ctx, query, params.Limit, params.Offset)
	if err != nil {
		return nil, core.Pagination{}, fmt.Errorf("failed to get promocodes: %w", err)
	}
	defer rows.Close()

	var promocodes []Promocode
	for rows.Next() {
		var p Promocode
		err := rows.Scan(
			&p.ID,
			&p.Code,
			&p.PlanID,
			&p.PlanName,
			&p.TrialPeriodDays,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, core.Pagination{}, fmt.Errorf("failed to scan promocode: %w", err)
		}
		promocodes = append(promocodes, p)
	}

	pagination := core.NewPagination(total, params.Page, params.Limit)

	return promocodes, pagination, nil
}

func (s *Service) GetPromocodeByID(ctx context.Context, id string) (*Promocode, error) {
	query := `
		SELECT 
			p.id, p.code, p.plan_id, pl.name as plan_name, p.trial_period_days,
			p.created_at, p.updated_at
		FROM pricing_promocodes p
		LEFT JOIN pricing_plans pl ON p.plan_id = pl.code
		WHERE p.id = $1
	`

	var p Promocode
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Code,
		&p.PlanID,
		&p.PlanName,
		&p.TrialPeriodDays,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get promocode: %w", err)
	}

	return &p, nil
}

// adminpricing/promocodes_service.go — исправленный CreatePromocode

func (s *Service) CreatePromocode(ctx context.Context, req *CreatePromocodeRequest) (*Promocode, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.PlanID == "" {
		return nil, fmt.Errorf("plan_id is required")
	}
	if req.TrialPeriodDays < 0 {
		return nil, fmt.Errorf("trial_period_days must be >= 0")
	}

	// проверяем существование плана
	plan, err := s.pricingService.GetPlanByCode(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan with code %s does not exist", req.PlanID)
	}
	_ = plan

	// создаём промокод напрямую через БД
	query := `
		INSERT INTO pricing_promocodes (code, plan_id, trial_period_days)
		VALUES ($1, $2, $3)
		RETURNING id, code, plan_id, trial_period_days, created_at, updated_at
	`

	var p Promocode
	err = s.pool.QueryRow(ctx, query, req.Code, req.PlanID, req.TrialPeriodDays).Scan(
		&p.ID,
		&p.Code,
		&p.PlanID,
		&p.TrialPeriodDays,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create promocode: %w", err)
	}

	// подтягиваем имя плана
	planName := ""
	if plan != nil {
		planName = plan.Name
	}
	p.PlanName = planName

	return &p, nil
}

func (s *Service) GetPromocodeByCode(ctx context.Context, code string) (*Promocode, error) {
	query := `
		SELECT 
			p.id, p.code, p.plan_id, pl.name as plan_name, p.trial_period_days,
			p.created_at, p.updated_at
		FROM pricing_promocodes p
		LEFT JOIN pricing_plans pl ON p.plan_id = pl.code
		WHERE p.code = $1
	`

	var p Promocode
	err := s.pool.QueryRow(ctx, query, code).Scan(
		&p.ID,
		&p.Code,
		&p.PlanID,
		&p.PlanName,
		&p.TrialPeriodDays,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get promocode: %w", err)
	}

	return &p, nil
}

func (s *Service) UpdatePromocode(ctx context.Context, id string, req *UpdatePromocodeRequest) (*Promocode, error) {
	// получаем текущий промокод
	current, err := s.GetPromocodeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("promocode not found: %w", err)
	}

	code := current.Code
	if req.Code != nil {
		code = *req.Code
	}

	planID := current.PlanID
	if req.PlanID != nil {
		// проверяем существование нового плана
		plan, err := s.pricingService.GetPlanByCode(ctx, *req.PlanID)
		if err != nil {
			return nil, fmt.Errorf("plan with code %s does not exist", *req.PlanID)
		}
		_ = plan
		planID = *req.PlanID
	}

	trialPeriodDays := current.TrialPeriodDays
	if req.TrialPeriodDays != nil {
		if *req.TrialPeriodDays < 0 {
			return nil, fmt.Errorf("trial_period_days must be >= 0")
		}
		trialPeriodDays = *req.TrialPeriodDays
	}

	updateQuery := `
		UPDATE pricing_promocodes
		SET code = $1, plan_id = $2, trial_period_days = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err = s.pool.Exec(ctx, updateQuery, code, planID, trialPeriodDays, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update promocode: %w", err)
	}

	return s.GetPromocodeByID(ctx, id)
}

func (s *Service) DeletePromocode(ctx context.Context, id string) error {
	query := `DELETE FROM pricing_promocodes WHERE id = $1`
	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete promocode: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("promocode not found")
	}

	return nil
}

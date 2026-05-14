package pricing

import (
	"context"
	"fmt"
)

// GetPromocodeByCode возвращает промокод по коду
func (s *Service) GetPromocodeByCode(ctx context.Context, code string) (*Promocode, error) {
	query := `
        SELECT id, code, plan_id, trial_period_days, created_at, updated_at
        FROM pricing_promocodes
        WHERE code = $1
    `

	var p Promocode
	err := s.pool.QueryRow(ctx, query, code).Scan(
		&p.ID,
		&p.Code,
		&p.PlanID,
		&p.TrialPeriodDays,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get promocode: %w", err)
	}

	return &p, nil
}

// GetPromocodeByID возвращает промокод по ID
func (s *Service) GetPromocodeByID(ctx context.Context, id string) (*Promocode, error) {
	query := `
        SELECT id, code, plan_id, trial_period_days, created_at, updated_at
        FROM pricing_promocodes
        WHERE id = $1
    `

	var p Promocode
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Code,
		&p.PlanID,
		&p.TrialPeriodDays,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get promocode: %w", err)
	}

	return &p, nil
}

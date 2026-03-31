package pricing

import (
	"context"
	"fmt"
	"strings"
)

// GetPlans возвращает список тарифных планов с пагинацией
func (s *Service) GetPlans(ctx context.Context, page, limit int, search string) ([]PricingPlan, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (page - 1) * limit

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := "SELECT COUNT(*) FROM pricing_plans " + whereClause
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count pricing plans: %w", err)
	}

	// Data
	query := `
		SELECT code, lvl, price_per_month, price_per_year, price_currency,
		       name, description, limit_db_mb, limit_objects_mb, limit_objects_count,
		       created_at, updated_at
		FROM pricing_plans
	` + whereClause + `
		ORDER BY lvl ASC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	allArgs := append(args, limit, offset)
	rows, err := s.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query pricing plans: %w", err)
	}
	defer rows.Close()

	var plans []PricingPlan
	for rows.Next() {
		var p PricingPlan
		err := rows.Scan(
			&p.Code, &p.Lvl, &p.PricePerMonth, &p.PricePerYear, &p.PriceCurrency,
			&p.Name, &p.Description, &p.LimitDbMB, &p.LimitObjectsMB, &p.LimitObjectsCount,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan pricing plan: %w", err)
		}
		plans = append(plans, p)
	}

	return plans, total, nil
}

// GetPlanByCode возвращает тарифный план по коду
func (s *Service) GetPlanByCode(ctx context.Context, code string) (*PricingPlan, error) {
	query := `
		SELECT code, lvl, price_per_month, price_per_year, price_currency,
		       name, description, limit_db_mb, limit_objects_mb, limit_objects_count,
		       created_at, updated_at
		FROM pricing_plans
		WHERE code = $1
	`

	var plan PricingPlan
	err := s.pool.QueryRow(ctx, query, code).Scan(
		&plan.Code, &plan.Lvl, &plan.PricePerMonth, &plan.PricePerYear, &plan.PriceCurrency,
		&plan.Name, &plan.Description, &plan.LimitDbMB, &plan.LimitObjectsMB, &plan.LimitObjectsCount,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get pricing plan by code: %w", err)
	}

	return &plan, nil
}

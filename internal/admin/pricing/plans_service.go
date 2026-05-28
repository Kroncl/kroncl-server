package adminpricing

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"kroncl-server/internal/pricing"
)

func (s *Service) GetPlans(ctx context.Context, page, limit int, search string) ([]pricing.PricingPlan, int, error) {
	return s.pricingService.GetPlans(ctx, page, limit, search)
}

func (s *Service) GetPlanByCode(ctx context.Context, code string) (*pricing.PricingPlan, error) {
	return s.pricingService.GetPlanByCode(ctx, code)
}

func (s *Service) UpdatePlan(ctx context.Context, code string, req UpdatePlanRequest) (*pricing.PricingPlan, error) {
	// Проверяем существование плана
	existing, err := s.pricingService.GetPlanByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	updater := core.NewUpdater("pricing_plans")

	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		updater.SetString("name", *req.Name)
	}

	if req.Description != nil {
		if *req.Description == "" {
			updater.SetNull("description")
		} else {
			updater.SetString("description", *req.Description)
		}
	}

	if req.PricePerMonth != nil {
		if *req.PricePerMonth < 0 {
			return nil, fmt.Errorf("price_per_month cannot be negative")
		}
		updater.SetInt("price_per_month", *req.PricePerMonth)
	}

	if req.PricePerYear != nil {
		if *req.PricePerYear < 0 {
			return nil, fmt.Errorf("price_per_year cannot be negative")
		}
		updater.SetInt("price_per_year", *req.PricePerYear)
	}

	if req.LimitDbMB != nil {
		if *req.LimitDbMB < 0 {
			return nil, fmt.Errorf("limit_db_mb cannot be negative")
		}
		updater.SetInt("limit_db_mb", *req.LimitDbMB)
	}

	if req.LimitObjectsMB != nil {
		if *req.LimitObjectsMB < 0 {
			return nil, fmt.Errorf("limit_objects_mb cannot be negative")
		}
		updater.SetInt("limit_objects_mb", *req.LimitObjectsMB)
	}

	if req.LimitObjectsCount != nil {
		if *req.LimitObjectsCount < 0 {
			return nil, fmt.Errorf("limit_objects_count cannot be negative")
		}
		updater.SetInt("limit_objects_count", *req.LimitObjectsCount)
	}

	query, args := updater.Where("code = $1", code).Build()
	if query == "" {
		return existing, nil
	}

	_, err = s.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return s.pricingService.GetPlanByCode(ctx, code)
}

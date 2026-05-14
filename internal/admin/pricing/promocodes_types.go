package adminpricing

import (
	"kroncl-server/internal/pricing"
)

type Promocode struct {
	pricing.Promocode
	PlanName string `json:"plan_name"`
}

type CreatePromocodeRequest struct {
	Code            string `json:"code"`
	PlanID          string `json:"plan_id"`
	TrialPeriodDays int    `json:"trial_period_days"`
}

type UpdatePromocodeRequest struct {
	Code            *string `json:"code,omitempty"`
	PlanID          *string `json:"plan_id,omitempty"`
	TrialPeriodDays *int    `json:"trial_period_days,omitempty"`
}

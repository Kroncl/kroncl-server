package companies

import (
	"kroncl-server/internal/pricing"
	"time"
)

// CompanyPlanResponse структура для ответа о текущем плане компании
type CompanyPlanResponse struct {
	IsTrial     bool                 `json:"is_trial"`
	ExpiresAt   time.Time            `json:"expires_at"`
	DaysLeft    int                  `json:"days_left"`
	CurrentPlan pricing.PricingPlan  `json:"current_plan"`
	NextPlan    *pricing.PricingPlan `json:"next_plan,omitempty"`
}

// MigratePlanRequest запрос на смену тарифа
type MigratePlanRequest struct {
	PlanCode string `json:"plan_code"`
	Period   string `json:"period"` // "month" или "year"
}

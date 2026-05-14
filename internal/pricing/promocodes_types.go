package pricing

import "time"

type Promocode struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	PlanID          string    `json:"plan_id"`
	TrialPeriodDays int       `json:"trial_period_days"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

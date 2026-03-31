package pricing

import (
	"time"
)

type TransactionStatus string

const (
	TransactionStatusSuccess   TransactionStatus = "success"
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusUnsuccess TransactionStatus = "unsuccess"
)

type PricingTransaction struct {
	ID           string            `json:"id"`
	CompanyID    string            `json:"company_id"`
	AccountID    string            `json:"account_id"`
	Amount       *int              `json:"amount"`
	Currency     Currency          `json:"currency"`
	Status       TransactionStatus `json:"status"`
	PlanCode     *string           `json:"plan_code"`
	IsTrial      bool              `json:"is_trial"`
	NextPlanCode *string           `json:"next_plan_code"`
	ExpiresAt    time.Time         `json:"expires_at"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

package accounts

import "time"

type ConfirmationCode struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Code      string    `json:"code"`
	Type      string    `json:"type"` // email_confirmation, password_reset, etc.
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

package accounts

import "time"

type ConfirmationCode struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	CodeHash  string    `json:"-"`
	Type      string    `json:"type"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

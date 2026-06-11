package accounts

import (
	"kroncl-server/internal/core"
	"time"
)

type ApiKey struct {
	ID            string     `json:"id"`
	AccountID     string     `json:"account_id"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	DailyRequests int        `json:"daily_requests"`
	LastUsedAt    *time.Time `json:"last_used_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	RevokedAt     *time.Time `json:"revoked_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ApiKeyWithRaw struct {
	ApiKey
	RawKey string `json:"raw_key"`
}

type CreateApiKeyRequest struct {
	Name          string `json:"name"`
	ExpiresIn     string `json:"expires_in,omitempty"` // "24h", "30d", "never"
	DailyRequests *int   `json:"daily_requests,omitempty"`
}

type ApiKeyListItem struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	DailyRequests int        `json:"daily_requests"`
	LastUsedAt    *time.Time `json:"last_used_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	RevokedAt     *time.Time `json:"revoked_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ApiKeyListRequest struct {
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	Search *string `json:"search,omitempty"`
	Status *string `json:"status,omitempty"` // "active", "revoked"
}

type ApiKeysResponse struct {
	ApiKeys    []ApiKeyListItem `json:"api_keys"`
	Pagination core.Pagination  `json:"pagination"`
}

package accounts

import (
	"kroncl-server/internal/core"
	"time"
)

// ----------
// FINGERPRINTS
// ----------

type Fingerprint struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"` // active, inactive
	ExpiredAt  *time.Time `json:"expired_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// FingerprintListRequest запрос на получение списка фингерпринтов
type FingerprintListRequest struct {
	Page   int     `json:"page" validate:"omitempty,min=1"`
	Limit  int     `json:"limit" validate:"omitempty,min=1,max=100"`
	Status *string `json:"status,omitempty"` // active, inactive
	Search *string `json:"search,omitempty"` // поиск по id или маске
}

// FingerprintListItem фингерпринт для списка
type FingerprintListItem struct {
	Fingerprint
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	MaskedKey  string     `json:"masked_key"` // fp_abc...xyz
}

// FingerprintsResponse пагинированный ответ
type FingerprintsResponse struct {
	Fingerprints []FingerprintListItem `json:"fingerprints"`
	Pagination   core.Pagination       `json:"pagination"`
}

type FingerprintWithKey struct {
	Fingerprint
	Key string `json:"key"` // Только при создании!
}

// FingerprintCreateRequest запрос на создание фингерпринта
type FingerprintCreateRequest struct {
	ExpiresIn *string `json:"expires_in,omitempty"` // "30d", "24h", "never" или null
}

// FingerprintLoginRequest вход по фингерпринту
type FingerprintLoginRequest struct {
	Key string `json:"key"`
}

// FingerprintLoginResponse ответ при входе по фингерпринту
type FingerprintLoginResponse struct {
	AccessToken string   `json:"access_token"`
	User        *Account `json:"user"`
}

// FingerprintRevokeRequest отзыв фингерпринта
type FingerprintRevokeRequest struct {
	ID string `json:"id"`
}

package accounts

import (
	"kroncl-server/internal/core"
	"time"
)

// Account модель аккаунта
type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AuthType  string    `json:"auth_type"`
	Status    string    `json:"status"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountPublic struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// RegisterResponse ответ на регистрацию
type RegisterResponse struct {
	Message      string `json:"message"`
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	EmailSent    bool   `json:"email_sent"`
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse ответ на вход
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         *Account `json:"user"`
}

// ConfirmRequest запрос на подтверждение
type ConfirmRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// запрос на обновление
type UpdateRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
}

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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// FingerprintRevokeRequest отзыв фингерпринта
type FingerprintRevokeRequest struct {
	ID string `json:"id"`
}

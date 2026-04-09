package hrm

import (
	"time"
)

// AccountSettings настройки аккаунта в рамках компании
type AccountSettings struct {
	AccountID           string    `json:"account_id"`
	IncreasePermissions []string  `json:"increase_permissions"` // дополнительные разрешения
	ReducePermissions   []string  `json:"reduce_permissions"`   // исключенные разрешения
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// UpdateAccountSettingsRequest запрос на обновление настроек аккаунта
type UpdateAccountSettingsRequest struct {
	IncreasePermissions []string `json:"increase_permissions,omitempty"`
	ReducePermissions   []string `json:"reduce_permissions,omitempty"`
}

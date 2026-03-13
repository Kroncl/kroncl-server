package accounts

import "time"

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

// запрос на обновление
type UpdateRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
}

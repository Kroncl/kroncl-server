package accounts

import "time"

// Account модель аккаунта
type Account struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AuthType  string    `json:"auth_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// RegisterResponse ответ на регистрацию
type RegisterResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

package accounts

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

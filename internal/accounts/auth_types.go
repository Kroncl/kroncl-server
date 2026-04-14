package accounts

const (
	ACCOUNT_STATUS_WAITING   = "waiting"
	ACCOUNT_STATUS_CONFIRMED = "confirmed"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Message     string `json:"message"`
	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token,omitempty"`
	EmailSent   bool   `json:"email_sent"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	User        *Account `json:"user"`
}

type ConfirmRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

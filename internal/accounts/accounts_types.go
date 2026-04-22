package accounts

import "time"

const (
	AccountTypeOwner       = "owner"
	AccountTypeEmployee    = "employee"
	AccountTypeAdmin       = "admin"
	AccountTypeOutsourcing = "outsourcing"
	AccountTypeTech        = "tech"
)

var validAccountTypes = map[string]bool{
	AccountTypeOwner:       true,
	AccountTypeEmployee:    true,
	AccountTypeAdmin:       true,
	AccountTypeOutsourcing: true,
	AccountTypeTech:        true,
}

type Account struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	AuthType    string    `json:"auth_type"`
	Status      string    `json:"status"`
	AvatarURL   string    `json:"avatar_url"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AccountPublic struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Status      string    `json:"status"`
	AvatarURL   string    `json:"avatar_url"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
}

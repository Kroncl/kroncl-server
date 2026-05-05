package admincompanies

import (
	"kroncl-server/internal/companies"
	"time"
)

type AdminCompany struct {
	companies.Company
	StorageStatus string `json:"storage_status"`
	StorageReady  bool   `json:"storage_ready"`
	SchemaName    string `json:"schema_name"`
}

type CompanyAccount struct {
	AccountID  string    `json:"account_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Status     string    `json:"status"`
	AvatarURL  *string   `json:"avatar_url"`
	RoleCode   string    `json:"role_code"`
	JoinedAt   time.Time `json:"joined_at"`
	IsAdmin    bool      `json:"is_admin"`
	AdminLevel int       `json:"admin_level"`
}

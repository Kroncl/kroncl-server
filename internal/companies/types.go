package companies

import (
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/storage"
	"time"
)

// Company модель компании
type Company struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AvatarUrl   string    `json:"avatar_url"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarUrl   string `json:"avatar_url"`
	IsPublic    bool   `json:"is_public"`
}

// UserCompany модель для связи пользователя с компанией и ролью
type UserCompany struct {
	Company
	RoleID   string    `json:"role_id"`
	RoleCode string    `json:"role_code"`
	RoleName string    `json:"role_name"`
	JoinedAt time.Time `json:"joined_at"`
}

// GetUserCompaniesRequest запрос для получения компаний пользователя
type GetUserCompaniesRequest struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Role   string `json:"role"` // "owner", "guest", "all"
	Search string `json:"search"`
}

// GetUserCompaniesResponse ответ с пагинацией
type GetUserCompaniesResponse struct {
	Companies  []UserCompany   `json:"companies"`
	Pagination core.Pagination `json:"pagination"`
}

// Role модель роли
type Role struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// запрос на обновление
type UpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
}

type CreateCompanyResponse struct {
	Company
	Storage *storage.Storage `json:"storage"`
}

// CompanyPublicMember публичная информация об участнике компании
type CompanyPublicMember struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"` // дата создания аккаунта
	RoleID    string    `json:"role_id"`
	RoleCode  string    `json:"role_code"`
	RoleName  string    `json:"role_name"`
	JoinedAt  time.Time `json:"joined_at"` // дата присоединения к компании
}

// RoleInfo информация о роли
type RoleInfo struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// GetCompanyMembersResponse ответ с участниками компании
type GetCompanyMembersResponse struct {
	Members    []CompanyPublicMember `json:"accounts"`
	Pagination core.Pagination       `json:"pagination"`
}

type GetCompanyMembersRequest struct {
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Search    string `json:"search,omitempty"`
	Role      string `json:"role,omitempty"`       // "all", "owner", "admin", "member", "guest"
	SortBy    string `json:"sort_by,omitempty"`    // "name", "joined_at", "role"
	SortOrder string `json:"sort_order,omitempty"` // "asc", "desc"
}

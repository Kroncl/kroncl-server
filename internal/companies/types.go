package companies

import "time"

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
	Companies []UserCompany `json:"companies"`
	Total     int           `json:"total"`
	Page      int           `json:"page"`
	Limit     int           `json:"limit"`
	Pages     int           `json:"pages"`
}

// Role модель роли
type Role struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

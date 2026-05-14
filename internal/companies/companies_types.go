package companies

import (
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/storage"
	"time"
)

const (
	RegionRu = "ru-RU"
	RegionKz = "kz-KZ"
)

var ValidRegions = map[string]bool{
	RegionRu: true,
	RegionKz: true,
}

func IsValidRegion(region string) bool {
	return ValidRegions[region]
}

type Company struct {
	ID          string                 `json:"id"`
	Slug        string                 `json:"slug"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	AvatarUrl   string                 `json:"avatar_url"`
	IsPublic    bool                   `json:"is_public"`
	Email       *string                `json:"email,omitempty"`
	Region      string                 `json:"region"`
	Site        *string                `json:"site,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type CreateRequest struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	AvatarUrl   string  `json:"avatar_url"`
	IsPublic    bool    `json:"is_public"`
	PlanCode    string  `json:"plan_code"`
	Region      string  `json:"region"`
	Promocode   *string `json:"promocode,omitempty"`
}

type UserCompany struct {
	Company
	RoleCode string    `json:"role_code"`
	JoinedAt time.Time `json:"joined_at"`
}

type GetUserCompaniesRequest struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Role   string `json:"role"`
	Search string `json:"search"`
}

type GetUserCompaniesResponse struct {
	Companies  []UserCompany   `json:"companies"`
	Pagination core.Pagination `json:"pagination"`
}

type UpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
	Email       *string `json:"email,omitempty"`
	Region      *string `json:"region,omitempty"`
	Site        *string `json:"site,omitempty"`
}

type CreateCompanyResponse struct {
	Company
	Storage *storage.Storage `json:"storage"`
}

type CompanyPublicMember struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	AvatarURL *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	RoleCode  string    `json:"role_code"`
	JoinedAt  time.Time `json:"joined_at"`
}

type GetCompanyMembersResponse struct {
	Members    []CompanyPublicMember `json:"accounts"`
	Pagination core.Pagination       `json:"pagination"`
}

type GetCompanyMembersRequest struct {
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Search    string `json:"search,omitempty"`
	Role      string `json:"role,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

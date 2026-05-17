package companies

import (
	"kroncl-server/internal/core"
	"time"
)

// ----------
// INVITATIONS
// ----------

const (
	RoleOwner = "owner" // все возможности
	RoleGuest = "guest" // read only
)

const (
	InvitationStatusWaiting  = "waiting"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRejected = "rejected"
)

type CompanyInvitation struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CompanyID string    `json:"company_id"`
	Status    string    `json:"status"`
	RoleCode  *string   `json:"role_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInvitationRequest struct {
	Email    string `json:"email" validate:"required,email"`
	RoleCode string `json:"role_code,omitempty"`
}

type InvitationResponse struct {
	Invitation CompanyInvitation `json:"invitation"`
	Message    string            `json:"message,omitempty"`
}

type GetInvitationsRequest struct {
	Search string `json:"search,omitempty"`
	Status string `json:"status,omitempty"`
	core.PaginationParams
}

type GetInvitationsResponse struct {
	Invitations []CompanyInvitation `json:"invitations"`
	Pagination  core.Pagination     `json:"pagination"`
}

type InvitationWithCompany struct {
	CompanyInvitation
	CompanyName      string `json:"company_name"`
	CompanyAvatarURL string `json:"company_avatar_url,omitempty"`
}

type GetInvitationsByEmailRequest struct {
	Status string `json:"status,omitempty"`
	core.PaginationParams
}

type GetInvitationsByEmailResponse struct {
	Invitations []InvitationWithCompany `json:"invitations"`
	Pagination  core.Pagination         `json:"pagination"`
}

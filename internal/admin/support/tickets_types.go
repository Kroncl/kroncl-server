package adminsupport

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/tenant/support"
	"time"
)

type AdminTicket struct {
	ID              string                 `json:"id"`
	Company         companies.Company      `json:"company"`
	Initiator       accounts.AccountPublic `json:"initiator"`
	Theme           string                 `json:"theme"`
	Status          support.TicketStatus   `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastMessage     *support.Message       `json:"last_message,omitempty"`
	AssignedAdminID *string                `json:"assigned_admin_id,omitempty"`
}

type GetTicketsRequest struct {
	Status *support.TicketStatus `json:"status"`
	Page   int                   `json:"page"`
	Limit  int                   `json:"limit"`
}

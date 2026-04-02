package support

import (
	"kroncl-server/internal/accounts"
	"time"
)

type TicketStatus string

const (
	TicketStatusPending TicketStatus = "pending"
	TicketStatusClosed  TicketStatus = "closed"
	TicketStatusRevoked TicketStatus = "revoked"
)

type TicketTheme string

const (
	ThemeTechnicalIssue TicketTheme = "technical_issue"
	ThemeBillingPayment TicketTheme = "billing_payment"
	ThemeAccessRights   TicketTheme = "access_rights"
	ThemeFeatureRequest TicketTheme = "feature_request"
	ThemeConsultation   TicketTheme = "consultation"
)

type Ticket struct {
	ID          string                 `json:"id"`
	CompanyID   string                 `json:"company_id"`
	InitiatorID string                 `json:"initiator_id"`
	Theme       string                 `json:"theme"`
	Status      TicketStatus           `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Initiator   accounts.AccountPublic `json:"initiator"`
	LastMessage *Message               `json:"last_message,omitempty"`
}

type Message struct {
	ID        string                 `json:"id"`
	AccountID string                 `json:"account_id"`
	TicketID  string                 `json:"ticket_id"`
	Text      string                 `json:"text"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Account   accounts.AccountPublic `json:"account"`
	Links     []Link                 `json:"links,omitempty"`
}

type Link struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	Link      string    `json:"link"`
	Capture   string    `json:"capture"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTicketRequest struct {
	Theme string `json:"theme"`
	Text  string `json:"text"`
}

type CreateMessageRequest struct {
	Text string `json:"text"`
}

type UpdateTicketRequest struct {
	Status TicketStatus `json:"status"`
}

func AllThemes() []TicketTheme {
	return []TicketTheme{
		ThemeTechnicalIssue,
		ThemeBillingPayment,
		ThemeAccessRights,
		ThemeFeatureRequest,
		ThemeConsultation,
	}
}

func IsValidTheme(theme string) bool {
	for _, t := range AllThemes() {
		if string(t) == theme {
			return true
		}
	}
	return false
}

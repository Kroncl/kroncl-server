package public

import "time"

const (
	PARTNER_TYPE_PUBLIC  = "public"
	PARTNER_TYPE_PRIVATE = "private"

	PARTNER_STATUS_SUCCESS = "success"
	PARTNER_STATUS_WAITING = "waiting"
	PARTNER_STATUS_BANNED  = "banned"
)

type PartnerStatus string
type PartnerType string

type IncomingPartner struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Text      *string   `json:"text,omitempty"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateIncomingPartnerRequest struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Text  *string `json:"text,omitempty"`
	Email string  `json:"email"`
}

type UpdateIncomingPartnerRequest struct {
	Name   *string `json:"name,omitempty"`
	Type   *string `json:"type,omitempty"`
	Text   *string `json:"text,omitempty"`
	Email  *string `json:"email,omitempty"`
	Status *string `json:"status,omitempty"`
}

package crm

import "time"

// ClientType represents the type of client
type ClientType string

const (
	ClientTypeIndividual ClientType = "individual"
	ClientTypeLegal      ClientType = "legal"
)

// ClientStatus represents the status of a client
type ClientStatus string

const (
	ClientStatusActive   ClientStatus = "active"
	ClientStatusInactive ClientStatus = "inactive"
)

// Client represents a client (individual or legal entity)
type Client struct {
	ID         string                 `json:"id"`
	FirstName  string                 `json:"first_name"`
	LastName   *string                `json:"last_name"`
	Patronymic *string                `json:"patronymic"`
	Phone      *string                `json:"phone"`
	Email      *string                `json:"email"`
	Comment    *string                `json:"comment"`
	Type       ClientType             `json:"type"`
	Status     ClientStatus           `json:"status"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CreateClientRequest represents request to create a client
type CreateClientRequest struct {
	FirstName  string                 `json:"first_name" validate:"required,min=1,max=100"`
	LastName   *string                `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Patronymic *string                `json:"patronymic,omitempty" validate:"omitempty,min=1,max=100"`
	Phone      *string                `json:"phone,omitempty"`
	Email      *string                `json:"email,omitempty"`
	Comment    *string                `json:"comment,omitempty"`
	Type       ClientType             `json:"type" validate:"required,oneof=individual legal"`
	Status     ClientStatus           `json:"status,omitempty"` // defaults to active
	SourceID   string                 `json:"source_id" validate:"required"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateClientRequest represents request to update a client
type UpdateClientRequest struct {
	FirstName  *string                 `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName   *string                 `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Patronymic *string                 `json:"patronymic,omitempty" validate:"omitempty,min=1,max=100"`
	Phone      *string                 `json:"phone,omitempty"`
	Email      *string                 `json:"email,omitempty"`
	Comment    *string                 `json:"comment,omitempty"`
	SourceID   *string                 `json:"source_id,omitempty"`
	Type       *ClientType             `json:"type,omitempty" validate:"omitempty,oneof=individual legal"`
	Status     *ClientStatus           `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Metadata   *map[string]interface{} `json:"metadata,omitempty"`
}

// GetClientsRequest represents request params for listing clients
type GetClientsRequest struct {
	Page     int           `json:"page" validate:"omitempty,min=1"`
	Limit    int           `json:"limit" validate:"omitempty,min=1,max=100"`
	Type     *ClientType   `json:"type,omitempty"`
	Status   *ClientStatus `json:"status,omitempty"`
	Search   *string       `json:"search,omitempty"`
	SourceID *string       `json:"source_id,omitempty"`
}

// ClientsResponse represents paginated response
type ClientsResponse struct {
	Clients []ClientDetail `json:"clients"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	Pages   int            `json:"pages"`
}

// ClientDetail represents detailed client view with sources
type ClientDetail struct {
	Client
	Source ClientSource `json:"source"`
}

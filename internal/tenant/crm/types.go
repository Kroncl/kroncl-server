package crm

import (
	"time"
)

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
	Type       *ClientType             `json:"type,omitempty" validate:"omitempty,oneof=individual legal"`
	Status     *ClientStatus           `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Metadata   *map[string]interface{} `json:"metadata,omitempty"`
}

// GetClientsRequest represents request params for listing clients
type GetClientsRequest struct {
	Page   int           `json:"page" validate:"omitempty,min=1"`
	Limit  int           `json:"limit" validate:"omitempty,min=1,max=100"`
	Type   *ClientType   `json:"type,omitempty"`
	Status *ClientStatus `json:"status,omitempty"`
	Search *string       `json:"search,omitempty"`
}

// ClientsResponse represents paginated response
type ClientsResponse struct {
	Clients []ClientDetail `json:"clients"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	Pages   int            `json:"pages"`
}

// ---------
// SOURCES
// ---------

// SourceType represents the type of traffic source
type SourceType string

const (
	SourceTypeOrganic  SourceType = "organic"
	SourceTypeSocial   SourceType = "social"
	SourceTypeReferral SourceType = "referral"
	SourceTypePaid     SourceType = "paid"
	SourceTypeEmail    SourceType = "email"
	SourceTypeOther    SourceType = "other"
)

// SourceStatus represents the status of a source
type SourceStatus string

const (
	SourceStatusActive   SourceStatus = "active"
	SourceStatusInactive SourceStatus = "inactive"
)

// ClientSource represents a traffic source
type ClientSource struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	URL       *string                `json:"url"`
	Type      SourceType             `json:"type"`
	Comment   *string                `json:"comment"`
	System    bool                   `json:"system"`
	Status    SourceStatus           `json:"status"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// CreateSourceRequest represents request to create a source
type CreateSourceRequest struct {
	Name     string                 `json:"name" validate:"required,min=1,max=255"`
	URL      *string                `json:"url,omitempty"`
	Type     SourceType             `json:"type" validate:"required,oneof=organic social referral paid email other"`
	Comment  *string                `json:"comment,omitempty"`
	System   bool                   `json:"system"`
	Status   SourceStatus           `json:"status,omitempty"` // defaults to active
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateSourceRequest represents request to update a source
type UpdateSourceRequest struct {
	Name     *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	URL      *string                 `json:"url,omitempty"`
	Type     *SourceType             `json:"type,omitempty" validate:"omitempty,oneof=organic social referral paid email other"`
	Comment  *string                 `json:"comment,omitempty"`
	Status   *SourceStatus           `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
}

// GetSourcesRequest represents request params for listing sources
type GetSourcesRequest struct {
	Page   int           `json:"page" validate:"omitempty,min=1"`
	Limit  int           `json:"limit" validate:"omitempty,min=1,max=100"`
	Type   *SourceType   `json:"type,omitempty"`
	Status *SourceStatus `json:"status,omitempty"`
	System *bool         `json:"system,omitempty"`
	Search *string       `json:"search,omitempty"`
}

// SourcesResponse represents paginated response
type SourcesResponse struct {
	Sources []ClientSource `json:"sources"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	Pages   int            `json:"pages"`
}

// ---------
// CLIENT-SOURCE LINKS
// ---------

// ClientSourceLink represents a link between a client and a source
type ClientSourceLink struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"client_id"`
	SourceID  string    `json:"source_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ClientDetail represents detailed client view with sources
type ClientDetail struct {
	Client
	Sources []ClientSource `json:"sources"`
}

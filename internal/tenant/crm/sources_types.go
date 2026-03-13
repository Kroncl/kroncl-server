package crm

import "time"

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

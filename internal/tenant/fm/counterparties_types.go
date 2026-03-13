package fm

import "time"

// -------
// COUNTERPARTIES
// -------

// CounterpartyType represents the type of counterparty
type CounterpartyType string

const (
	CounterpartyTypeBank         CounterpartyType = "bank"
	CounterpartyTypeOrganization CounterpartyType = "organization"
	CounterpartyTypePerson       CounterpartyType = "person"
)

// CounterpartyStatus represents the status of a counterparty
type CounterpartyStatus string

const (
	CounterpartyStatusActive   CounterpartyStatus = "active"
	CounterpartyStatusInactive CounterpartyStatus = "inactive"
)

// Counterparty represents a counterparty (creditor/debtor)
type Counterparty struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Comment   *string                `json:"comment"`
	Type      CounterpartyType       `json:"type"`
	Status    CounterpartyStatus     `json:"status"` // new field
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// CreateCounterpartyRequest represents request to create counterparty
type CreateCounterpartyRequest struct {
	Name     string                 `json:"name" validate:"required,min=1,max=255"`
	Comment  string                 `json:"comment,omitempty" validate:"omitempty,max=1000"`
	Type     CounterpartyType       `json:"type" validate:"required,oneof=bank organization person"`
	Status   CounterpartyStatus     `json:"status" validate:"omitempty,oneof=active inactive"` // optional, defaults to active
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCounterpartyRequest represents request to update counterparty
type UpdateCounterpartyRequest struct {
	Name     *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment  *string                 `json:"comment,omitempty" validate:"omitempty,max=1000"`
	Type     *CounterpartyType       `json:"type,omitempty" validate:"omitempty,oneof=bank organization person"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
}

// GetCounterpartiesRequest represents request params for listing counterparties
type GetCounterpartiesRequest struct {
	Page   int                 `json:"page" validate:"omitempty,min=1"`
	Limit  int                 `json:"limit" validate:"omitempty,min=1,max=100"`
	Type   *CounterpartyType   `json:"type,omitempty"`
	Status *CounterpartyStatus `json:"status,omitempty"`
	Search *string             `json:"search,omitempty"`
}

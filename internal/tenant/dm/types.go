package dm

import (
	"time"
)

// ---------
// TYPES
// ---------

type DealType struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateDealTypeRequest struct {
	Name    string  `json:"name" validate:"required,min=1,max=255"`
	Comment *string `json:"comment,omitempty"`
}

type UpdateDealTypeRequest struct {
	Name    *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment *string `json:"comment,omitempty"`
}

// ---------
// STATUSES
// ---------

type DealStatus struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Comment   *string   `json:"comment"`
	SortOrder int       `json:"sort_order"`
	Color     *string   `json:"color"` // HEX-код, например "#FF5733"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateDealStatusRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=255"`
	Comment   *string `json:"comment,omitempty"`
	SortOrder int     `json:"sort_order" validate:"min=1"`
	Color     *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

type UpdateDealStatusRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment   *string `json:"comment,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty" validate:"omitempty,min=1"`
	Color     *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// ---------
// DEALS
// ---------

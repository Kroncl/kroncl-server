package hrm

import (
	"time"
)

// ---------
// POSITIONS
// ---------

type Position struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePositionRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=255"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions []string `json:"permissions,omitempty"`
}

type UpdatePositionRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions []string `json:"permissions,omitempty"`
}

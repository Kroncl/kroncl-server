package fm

import "time"

// ---------
// CATEGORIES
// ---------

// TransactionCategoryDirection represents income/expense direction for categories
type TransactionCategoryDirection string

const (
	TransactionCategoryDirectionIncome  TransactionCategoryDirection = "income"
	TransactionCategoryDirectionExpense TransactionCategoryDirection = "expense"
)

// TransactionCategory represents a transaction category
type TransactionCategory struct {
	ID          string                       `json:"id"`
	Name        string                       `json:"name"`
	Description *string                      `json:"description"`
	Direction   TransactionCategoryDirection `json:"direction"`
	System      bool                         `json:"system"`
	Slug        string                       `json:"slug"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
}

// CreateCategoryRequest represents request to create transaction category
type CreateCategoryRequest struct {
	Name        string                       `json:"name" validate:"required,min=1,max=255"`
	Description string                       `json:"description,omitempty" validate:"omitempty,max=1000"`
	Direction   TransactionCategoryDirection `json:"direction" validate:"required,oneof=income expense"`
	System      bool                         `json:"system"` // true только для системных
}

// UpdateCategoryRequest represents request to update transaction category
type UpdateCategoryRequest struct {
	Name        *string                       `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string                       `json:"description,omitempty" validate:"omitempty,max=1000"`
	Direction   *TransactionCategoryDirection `json:"direction,omitempty" validate:"omitempty,oneof=income expense"`
}

// GetCategoriesRequest represents request params for listing categories
type GetCategoriesRequest struct {
	Page      int                           `json:"page" validate:"omitempty,min=1"`
	Limit     int                           `json:"limit" validate:"omitempty,min=1,max=100"`
	Direction *TransactionCategoryDirection `json:"direction,omitempty"`
	Search    *string                       `json:"search,omitempty"`
}

// CategoriesResponse represents paginated response
type CategoriesResponse struct {
	Categories []TransactionCategory `json:"categories"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Pages      int                   `json:"pages"`
}

package fm

import "time"

// ---------
// CREDITS
// ---------

// CreditType represents whether we owe money or money is owed to us
type CreditType string

const (
	CreditTypeDebt   CreditType = "debt"   // мы должны
	CreditTypeCredit CreditType = "credit" // нам должны
)

// CreditStatus represents the status of a credit
type CreditStatus string

const (
	CreditStatusActive CreditStatus = "active"
	CreditStatusClosed CreditStatus = "closed"
)

// Credit represents a credit or loan (базовый тип без контрагента)
type Credit struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Comment      *string                `json:"comment"`
	Type         CreditType             `json:"type"`
	Status       CreditStatus           `json:"status"`
	TotalAmount  int64                  `json:"total_amount"`
	Currency     CurrencyType           `json:"currency"`
	InterestRate float64                `json:"interest_rate"` // годовых в процентах
	StartDate    time.Time              `json:"start_date"`
	EndDate      time.Time              `json:"end_date"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// CreditDetail represents detailed credit view with counterparty data
type CreditDetail struct {
	Credit
	Counterparty *Counterparty `json:"counterparty"`
}

// CreateCreditRequest represents request to create a credit
type CreateCreditRequest struct {
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Comment        string                 `json:"comment,omitempty" validate:"omitempty,max=1000"`
	Type           CreditType             `json:"type" validate:"required,oneof=debt credit"`
	TotalAmount    int64                  `json:"total_amount" validate:"required,gt=0"`
	Currency       CurrencyType           `json:"currency" validate:"required,oneof=RUB"`
	InterestRate   float64                `json:"interest_rate" validate:"min=0,max=100"`
	StartDate      time.Time              `json:"start_date" validate:"required"`
	EndDate        time.Time              `json:"end_date" validate:"required,gtefield=StartDate"`
	CounterpartyID string                 `json:"counterparty_id" validate:"required,uuid"` // обязательная связь
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCreditRequest represents request to update a credit (без статуса)
type UpdateCreditRequest struct {
	Name           *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment        *string                 `json:"comment,omitempty" validate:"omitempty,max=1000"`
	Type           *CreditType             `json:"type,omitempty" validate:"omitempty,oneof=debt credit"`
	TotalAmount    *int64                  `json:"total_amount,omitempty" validate:"omitempty,gt=0"`
	Currency       *CurrencyType           `json:"currency,omitempty" validate:"omitempty,oneof=RUB"`
	InterestRate   *float64                `json:"interest_rate,omitempty" validate:"omitempty,min=0,max=100"`
	StartDate      *time.Time              `json:"start_date,omitempty"`
	EndDate        *time.Time              `json:"end_date,omitempty"`
	CounterpartyID *string                 `json:"counterparty_id,omitempty" validate:"omitempty,uuid"`
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
}

// GetCreditsRequest represents request params for listing credits
type GetCreditsRequest struct {
	Page   int           `json:"page" validate:"omitempty,min=1"`
	Limit  int           `json:"limit" validate:"omitempty,min=1,max=100"`
	Type   *CreditType   `json:"type,omitempty"`
	Status *CreditStatus `json:"status,omitempty"`
	Search *string       `json:"search,omitempty"`
}

// CreditsResponse represents paginated response
type CreditsResponse struct {
	Credits []CreditDetail `json:"credits"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	Pages   int            `json:"pages"`
}

// PayCreditRequest represents request to make a payment towards a credit
type PayCreditRequest struct {
	CreditID   string    `json:"credit_id" validate:"required,uuid"`
	EmployeeID string    `json:"employee_id" validate:"required,uuid"`
	Amount     int64     `json:"amount" validate:"required,gt=0"`
	PaidAt     time.Time `json:"paid_at" validate:"required"`
	Comment    string    `json:"comment,omitempty"`
}

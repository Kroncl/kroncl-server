package fm

import (
	"kroncl-server/internal/tenant/hrm"
	"time"
)

// TransactionStatus represents the state of a financial transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// TransactionDirection represents income/expense direction
type TransactionDirection string

const (
	TransactionDirectionIncome  TransactionDirection = "income"
	TransactionDirectionExpense TransactionDirection = "expense"
)

// CurrencyType represents supported currencies
type CurrencyType string

const (
	CurrencyRUB CurrencyType = "RUB"
	// CurrencyUSD CurrencyType = "USD"
	// CurrencyEUR CurrencyType = "EUR"
	// CurrencyKZT CurrencyType = "KZT"
)

// Transaction represents a financial transaction record
type Transaction struct {
	ID         string                 `json:"id"`
	BaseAmount int64                  `json:"base_amount"` // рубли/тенге/доллары/евро (целое число)
	Currency   CurrencyType           `json:"currency"`
	Direction  TransactionDirection   `json:"direction"`
	Status     TransactionStatus      `json:"status"`
	Comment    *string                `json:"comment"`
	ReverseTo  *string                `json:"reverse_to"`
	CreatedAt  time.Time              `json:"created_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TransactionListItem represents transaction in list views
type TransactionListItem struct {
	Transaction
	EmployeeID        *string `json:"employee_id"`
	EmployeeFirstName *string `json:"employee_first_name"`
	EmployeeLastName  *string `json:"employee_last_name"`
	CategoryID        *string `json:"category_id"`
	CategoryName      *string `json:"category_name"`
}

// TransactionDetail represents detailed transaction view
type TransactionDetail struct {
	TransactionListItem
	Employee *hrm.EmployeeDetail  `json:"employee,omitempty"`
	Category *TransactionCategory `json:"category,omitempty"`
}

// CreateTransactionRequest represents request to create transaction
type CreateTransactionRequest struct {
	BaseAmount int64                  `json:"base_amount" validate:"required,gt=0"` // рубли/тенге/доллары/евро
	Currency   CurrencyType           `json:"currency" validate:"required,oneof=RUB"`
	Direction  TransactionDirection   `json:"direction" validate:"required,oneof=income expense"`
	Comment    string                 `json:"comment,omitempty" validate:"omitempty,max=500"`
	CategoryID string                 `json:"category_id,omitempty" validate:"omitempty,uuid"`
	EmployeeID string                 `json:"employee_id" validate:"uuid"`
	Status     string                 `json:"status,omitempty" validate:"omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// GetTransactionsRequest represents request params for listing
type GetTransactionsRequest struct {
	Page       int                   `json:"page" validate:"omitempty,min=1"`
	Limit      int                   `json:"limit" validate:"omitempty,min=1,max=100"`
	StartDate  *time.Time            `json:"start_date,omitempty"`
	EndDate    *time.Time            `json:"end_date,omitempty"`
	Direction  *TransactionDirection `json:"direction,omitempty"`
	Status     *TransactionStatus    `json:"status,omitempty"`
	CategoryID *string               `json:"category_id,omitempty"`
	EmployeeID *string               `json:"employee_id,omitempty"`
	Search     *string               `json:"search,omitempty"`
}

// TransactionsResponse represents paginated response
type TransactionsResponse struct {
	Transactions []TransactionDetail `json:"transactions"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	Limit        int                 `json:"limit"`
	Pages        int                 `json:"pages"`
}

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

package fm

import (
	"kroncl-server/internal/tenant/hrm"
	"time"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type TransactionDirection string

const (
	TransactionDirectionIncome  TransactionDirection = "income"
	TransactionDirectionExpense TransactionDirection = "expense"
)

type Transaction struct {
	ID         string                 `json:"id"`
	BaseAmount float64                `json:"base_amount"`
	Currency   string                 `json:"currency"`
	Direction  TransactionDirection   `json:"direction"`
	Status     TransactionStatus      `json:"status"`
	Comment    *string                `json:"comment"`
	ReverseTo  *string                `json:"reverse_to"`
	CreatedAt  time.Time              `json:"created_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type TransactionListItem struct {
	Transaction
	EmployeeID        *string `json:"employee_id"`
	EmployeeFirstName *string `json:"employee_first_name"`
	EmployeeLastName  *string `json:"employee_last_name"`
	CategoryID        *string `json:"category_id"`
	CategoryName      *string `json:"category_name"`
}

type TransactionDetail struct {
	TransactionListItem
	Employee *hrm.EmployeeDetail  `json:"employee,omitempty"`
	Category *TransactionCategory `json:"category,omitempty"`
}

type CreateTransactionRequest struct {
	BaseAmount float64                `json:"base_amount" validate:"required,gt=0"`
	Currency   string                 `json:"currency" validate:"required"`
	Direction  TransactionDirection   `json:"direction" validate:"required,oneof=income expense"`
	Comment    string                 `json:"comment,omitempty" validate:"omitempty,max=500"`
	CategoryID string                 `json:"category_id,omitempty" validate:"omitempty,uuid"`
	EmployeeID string                 `json:"employee_id" validate:"uuid"`
	Status     string                 `json:"status,omitempty" validate:"omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

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

type TransactionsResponse struct {
	Transactions []TransactionDetail `json:"transactions"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	Limit        int                 `json:"limit"`
	Pages        int                 `json:"pages"`
}

type CurrencyType string

const (
	CurrencyRUB CurrencyType = "RUB"
	CurrencyUSD CurrencyType = "USD"
	CurrencyEUR CurrencyType = "EUR"
	CurrencyKZT CurrencyType = "KZT"
)

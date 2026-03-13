package hrm

import (
	"kroncl-server/internal/accounts"
	"time"
)

// ---------
// EMPLOYEES
// ---------

type EmployeeStatus string

const (
	EmployeeStatusActive   EmployeeStatus = "active"
	EmployeeStatusInactive EmployeeStatus = "inactive"
)

type Employee struct {
	ID        string         `json:"id"`
	FirstName string         `json:"first_name"`
	LastName  *string        `json:"last_name"`
	Email     *string        `json:"email"`
	Phone     *string        `json:"phone"`
	Status    EmployeeStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type EmployeeListItem struct {
	Employee
	AccountID       *string    `json:"account_id"`
	LinkedAt        *time.Time `json:"linked_at"`
	IsAccountLinked bool       `json:"is_account_linked"`
}

type EmployeeDetail struct {
	EmployeeListItem
	Account *accounts.AccountPublic `json:"account,omitempty"`
}

type CreateEmployeeRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"omitempty,min=2,max=100"`
	Email     string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,min=6,max=50"`
}

type UpdateEmployeeRequest struct {
	FirstName string         `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	LastName  string         `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
	Email     string         `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone     string         `json:"phone,omitempty" validate:"omitempty,min=6,max=50"`
	Status    EmployeeStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

type LinkAccountRequest struct {
	AccountId string `json:"account_id"`
}

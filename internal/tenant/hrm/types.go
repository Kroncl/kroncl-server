package hrm

import (
	"kroncl-server/internal/accounts"
	"time"
)

type EmployeeStatus string

const (
	EmployeeStatusActive   EmployeeStatus = "active"
	EmployeeStatusInactive EmployeeStatus = "inactive"
)

type Employee struct {
	ID        string         `json:"id"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Email     *string        `json:"email,omitempty"`
	Phone     *string        `json:"phone,omitempty"`
	Status    EmployeeStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// EmployeeAccount связь сотрудник ↔ аккаунт
type EmployeeAccount struct {
	ID         string    `json:"id"`
	AccountID  string    `json:"account_id"`
	EmployeeID string    `json:"employee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type EmployeeWithAccount struct {
	Employee
	Account         *accounts.AccountPublic `json:"account,omitempty"`
	AccountID       *string                 `json:"account_id,omitempty"`
	LinkedAt        *time.Time              `json:"linked_at,omitempty"`
	IsAccountLinked bool                    `json:"is_account_linked"`
}

type CreateEmployeeRequest struct {
	FirstName string  `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string  `json:"last_name" validate:"required,min=2,max=100"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,min=6,max=50"`
}

type CreateEmployeeResponse struct {
	Employee EmployeeWithAccount `json:"employee"`
	Message  string              `json:"message"`
}

type UpdateEmployeeRequest struct {
	FirstName *string         `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	LastName  *string         `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
	Email     *string         `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone     *string         `json:"phone,omitempty" validate:"omitempty,min=6,max=50"`
	Status    *EmployeeStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

// UpdateEmployeeResponse ответ при обновлении сотрудника
type UpdateEmployeeResponse struct {
	Employee EmployeeWithAccount `json:"employee"`
	Message  string              `json:"message"`
}

// LinkAccountRequest запрос на привязку аккаунта к сотруднику
type LinkAccountRequest struct {
	AccountID string `json:"account_id" validate:"required,uuid4"`
}

// LinkAccountResponse ответ при привязке аккаунта
type LinkAccountResponse struct {
	Employee        EmployeeWithAccount `json:"employee"`
	EmployeeAccount EmployeeAccount     `json:"employee_account"`
	Message         string              `json:"message"`
}

// UnlinkAccountRequest запрос на отвязку аккаунта
type UnlinkAccountRequest struct {
	AccountID string `json:"account_id" validate:"required,uuid4"`
}

// UnlinkAccountResponse ответ при отвязке аккаунта
type UnlinkAccountResponse struct {
	Employee EmployeeWithAccount `json:"employee"`
	Message  string              `json:"message"`
}

// GetEmployeesRequest запрос на получение списка сотрудников
type GetEmployeesRequest struct {
	Page      int            `json:"page" validate:"min=1"`
	Limit     int            `json:"limit" validate:"min=1,max=100"`
	Search    string         `json:"search,omitempty"`
	Status    EmployeeStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	HasLinked *bool          `json:"has_linked,omitempty"`
}

// GetEmployeesResponse ответ со списком сотрудников
type GetEmployeesResponse struct {
	Employees []EmployeeWithAccount `json:"employees"`
	Total     int                   `json:"total"`
	Page      int                   `json:"page"`
	Limit     int                   `json:"limit"`
	Pages     int                   `json:"pages"`
}

// GetEmployeeResponse ответ при получении сотрудника
type GetEmployeeResponse struct {
	Employee EmployeeWithAccount `json:"employee"`
}

// DeleteEmployeeResponse ответ при удалении сотрудника
type DeleteEmployeeResponse struct {
	Message string `json:"message"`
}

// EmployeeStats статистика по сотрудникам
type EmployeeStats struct {
	TotalEmployees          int `json:"total_employees"`
	ActiveEmployees         int `json:"active_employees"`
	InactiveEmployees       int `json:"inactive_employees"`
	EmployeesWithAccount    int `json:"employees_with_account"`
	EmployeesWithoutAccount int `json:"employees_without_account"`
}

// GetEmployeeStatsResponse ответ со статистикой
type GetEmployeeStatsResponse struct {
	Stats   EmployeeStats `json:"stats"`
	Updated time.Time     `json:"updated"`
}

// SearchAccountsRequest запрос на поиск аккаунтов для привязки
type SearchAccountsRequest struct {
	Search string `json:"search" validate:"required,min=2"` // по email или имени
	Limit  int    `json:"limit" validate:"min=1,max=50"`
}

// SearchAccountsResponse ответ с найденными аккаунтами
type SearchAccountsResponse struct {
	Accounts []accounts.AccountPublic `json:"accounts"`
	Total    int                      `json:"total"`
}

// AccountCandidate кандидат для привязки к сотруднику
type AccountCandidate struct {
	Account          accounts.AccountPublic `json:"account"`
	IsAlreadyLinked  bool                   `json:"is_already_linked"`
	LinkedEmployeeID *string                `json:"linked_employee_id,omitempty"`
}

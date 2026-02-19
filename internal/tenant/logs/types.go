package logs

import (
	"time"
)

// LogStatus represents the status of a log entry
type LogStatus string

const (
	LogStatusSuccess LogStatus = "success"
	LogStatusError   LogStatus = "error"
	LogStatusPending LogStatus = "pending"
)

// Log represents a log entry for user actions
type Log struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`                  // permission key (e.g., fm.transactions.create)
	Status      LogStatus              `json:"status"`               // success, error, pending
	Criticality int                    `json:"criticality"`          // 1-10
	AccountID   string                 `json:"account_id"`           // account ID from public schema
	RequestID   *string                `json:"request_id,omitempty"` // request ID for grouping
	UserAgent   *string                `json:"user_agent,omitempty"` // browser/client
	IP          *string                `json:"ip,omitempty"`         // IP address
	Metadata    map[string]interface{} `json:"metadata"`             // additional details
	CreatedAt   time.Time              `json:"created_at"`
}

// LogListItem represents log in list views
type LogListItem struct {
	Log
	// можно добавить дополнительные поля если нужно
}

// LogDetail represents detailed log view
type LogDetail struct {
	LogListItem
	// можно добавить дополнительные данные при необходимости
}

// GetLogsRequest represents request params for listing logs
type GetLogsRequest struct {
	Page           int        `json:"page" validate:"omitempty,min=1"`
	Limit          int        `json:"limit" validate:"omitempty,min=1,max=100"`
	AccountID      *string    `json:"account_id,omitempty"`
	Key            *string    `json:"key,omitempty"`    // фильтр по ключу (fm.transactions.create)
	Status         *LogStatus `json:"status,omitempty"` // фильтр по статусу
	MinCriticality *int       `json:"min_criticality,omitempty" validate:"omitempty,min=1,max=10"`
	MaxCriticality *int       `json:"max_criticality,omitempty" validate:"omitempty,min=1,max=10"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Search         *string    `json:"search,omitempty"` // поиск по metadata
}

// LogsResponse represents paginated response
type LogsResponse struct {
	Logs  []LogListItem `json:"logs"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Pages int           `json:"pages"`
}

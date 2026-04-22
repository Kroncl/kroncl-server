package logs

import (
	"time"
)

type LogStatus string

const (
	LogStatusSuccess LogStatus = "success"
	LogStatusError   LogStatus = "error"
	LogStatusPending LogStatus = "pending"
)

type Log struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	Status      LogStatus              `json:"status"`
	Criticality int                    `json:"criticality"`
	AccountID   string                 `json:"account_id"`
	RequestID   *string                `json:"request_id,omitempty"`
	UserAgent   *string                `json:"user_agent,omitempty"`
	IP          *string                `json:"ip,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

type LogListItem struct {
	Log
}

type LogDetail struct {
	LogListItem
}

type GetLogsRequest struct {
	Page           int        `json:"page" validate:"omitempty,min=1"`
	Limit          int        `json:"limit" validate:"omitempty,min=1,max=100"`
	AccountID      *string    `json:"account_id,omitempty"`
	Key            *string    `json:"key,omitempty"`
	Status         *LogStatus `json:"status,omitempty"`
	MinCriticality *int       `json:"min_criticality,omitempty" validate:"omitempty,min=1,max=10"`
	MaxCriticality *int       `json:"max_criticality,omitempty" validate:"omitempty,min=1,max=10"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Search         *string    `json:"search,omitempty"`
}

type LogsResponse struct {
	Logs  []LogListItem `json:"logs"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Pages int           `json:"pages"`
}

type LogActivity struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

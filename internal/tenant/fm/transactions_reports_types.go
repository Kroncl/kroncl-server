package fm

import (
	"kroncl-server/internal/core"
	"time"
)

type TransactionsReport struct {
	ID         string    `json:"id"`
	ObjectPath string    `json:"object_path"`
	Comment    *string   `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type GetReportsRequest struct {
	core.PaginationParams
}

type ReportsResponse struct {
	Reports []TransactionsReport `json:"reports"`
	Total   int64                `json:"total"`
	Page    int                  `json:"page"`
	Limit   int                  `json:"limit"`
	Pages   int                  `json:"pages"`
}

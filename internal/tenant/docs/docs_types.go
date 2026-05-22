package docs

import (
	"kroncl-server/internal/core"
	"time"
)

type CreateDocRequest struct {
	ObjectPath string  `json:"object_path"`
	Module     *string `json:"module,omitempty"`
	Type       *string `json:"type,omitempty"`
	Comment    *string `json:"comment,omitempty"`
}

type Doc struct {
	ID         string    `json:"id"`
	ObjectPath string    `json:"object_path"`
	Module     *string   `json:"module"`
	Type       *string   `json:"type"`
	Comment    *string   `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type GetDocsRequest struct {
	core.PaginationParams
	Module *string `json:"module,omitempty"`
	Type   *string `json:"type,omitempty"`
	Search *string `json:"search,omitempty"`
}

type DocsResponse struct {
	Docs       []Doc           `json:"docs"`
	Pagination core.Pagination `json:"pagination"`
}

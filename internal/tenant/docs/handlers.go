package docs

import (
	"net/http"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
)

type Handlers struct {
	service     *Service
	logsService *logs.Service
}

func NewHandlers(service *Service, logsService *logs.Service) *Handlers {
	return &Handlers{
		service:     service,
		logsService: logsService,
	}
}

func (h *Handlers) GetDocs(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	var module, docType, search *string

	if m := r.URL.Query().Get("module"); m != "" {
		module = &m
	}
	if t := r.URL.Query().Get("type"); t != "" {
		docType = &t
	}
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	docs, total, err := h.service.GetDocs(r.Context(), pagination.Offset, pagination.Limit, module, docType, search)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DOCS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, "Failed to get documents")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DOCS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("path", r.URL.Path),
		logs.WithMetadata("filters", map[string]interface{}{
			"module": module,
			"type":   docType,
			"search": search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(docs)),
	)

	response := DocsResponse{
		Docs:       docs,
		Pagination: core.NewPagination(int(total), pagination.Page, pagination.Limit),
	}

	core.SendSuccess(w, response, "Documents retrieved successfully")
}

func (h *Handlers) GetDoc(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	docID := r.PathValue("docId")
	if docID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DOCS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Document ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Document ID is required")
		return
	}

	doc, err := h.service.GetDocByID(r.Context(), docID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DOCS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("doc_id", docID),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendNotFound(w, "Document not found")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DOCS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("doc_id", docID),
		logs.WithMetadata("module", doc.Module),
		logs.WithMetadata("type", doc.Type),
	)

	core.SendSuccess(w, doc, "Document retrieved successfully")
}

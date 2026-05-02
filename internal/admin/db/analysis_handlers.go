package admindb

import (
	"fmt"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.GetSystemStats(r.Context())
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get system stats: %v", err))
		return
	}

	core.SendSuccess(w, response, "System stats.")
}

func (h *Handlers) GetSchemaStats(w http.ResponseWriter, r *http.Request) {
	schemaName := chi.URLParam(r, "schemaName")
	if schemaName == "" {
		core.SendValidationError(w, "schemaName is required")
		return
	}

	response, err := h.service.GetSchemaStats(r.Context(), schemaName)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get schema stats: %v", err))
		return
	}

	core.SendSuccess(w, response, "Schema stats.")
}

func (h *Handlers) GetSchemaTables(w http.ResponseWriter, r *http.Request) {
	schemaName := chi.URLParam(r, "schemaName")
	if schemaName == "" {
		core.SendValidationError(w, "schemaName is required")
		return
	}

	response, err := h.service.GetSchemaTables(r.Context(), schemaName)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get schema tables: %v", err))
		return
	}

	core.SendSuccess(w, response, "Schema tables.")
}

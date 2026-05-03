package admindb

import (
	"fmt"
	"kroncl-server/internal/core"
	"net/http"
	"strconv"
	"time"

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

func (h *Handlers) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// парсим start_date
	var startDate *time.Time
	if sd := query.Get("start_date"); sd != "" {
		t, err := time.Parse(time.RFC3339, sd)
		if err != nil {
			core.SendValidationError(w, "Invalid start_date format, use RFC3339")
			return
		}
		startDate = &t
	}

	// парсим end_date
	var endDate *time.Time
	if ed := query.Get("end_date"); ed != "" {
		t, err := time.Parse(time.RFC3339, ed)
		if err != nil {
			core.SendValidationError(w, "Invalid end_date format, use RFC3339")
			return
		}
		endDate = &t
	}

	// парсим limit
	limit := 100 // дефолт
	if l := query.Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err != nil {
			core.SendValidationError(w, "Invalid limit, must be integer")
			return
		}
		if parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		} else if parsedLimit > 1000 {
			limit = 1000
		}
	}

	response, err := h.service.metricsService.GetMetricsHistory(r.Context(), startDate, endDate, limit)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get metrics history: %v", err))
		return
	}

	core.SendSuccess(w, response, "Metrics history.")
}

func (h *Handlers) GetSchemas(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	search := query.Get("search")
	onlyTenants := query.Get("only_tenants") == "false"

	params := core.GetPaginationParams(r, 20, 100)

	schemas, pagination, err := h.service.GetSchemas(r.Context(), search, onlyTenants, params)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get schemas: %v", err))
		return
	}

	response := map[string]interface{}{
		"schemas":    schemas,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Schemas list.")
}

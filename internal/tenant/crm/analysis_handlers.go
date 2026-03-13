package crm

import (
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"time"
)

// ---------
// ANALYSIS
// ---------

func (h *Handlers) GetClientsSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим параметры дат
	var startDate, endDate *time.Time

	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			startDate = &t
		}
	}

	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			endDate = &t
		}
	}

	summary, err := h.repository.GetClientsSummary(r.Context(), startDate, endDate)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get clients summary: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_ANALYSIS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("date_range", map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		}),
	)

	core.SendSuccess(w, summary, "Clients summary retrieved successfully.")
}

func (h *Handlers) GetGroupedClients(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groupBy := GroupBy(r.URL.Query().Get("group_by"))
	if groupBy == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "group_by parameter is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "group_by parameter is required (source/day/month)")
		return
	}

	// Валидация group_by
	switch groupBy {
	case GroupBySource, GroupByDay, GroupByMonth:
		// valid
	default:
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid group_by value"),
			logs.WithMetadata("group_by", groupBy),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid group_by. Use: source, day, month")
		return
	}

	var startDate, endDate *time.Time
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		t, _ := time.Parse(time.RFC3339, startStr)
		startDate = &t
	}
	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		t, _ := time.Parse(time.RFC3339, endStr)
		endDate = &t
	}

	stats, err := h.repository.GetGroupedClients(r.Context(), groupBy, startDate, endDate)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get grouped clients: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_ANALYSIS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("group_by", groupBy),
		logs.WithMetadata("date_range", map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		}),
	)

	core.SendSuccess(w, stats, "Grouped clients retrieved successfully.")
}

package hrm

import (
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"time"
)

// GetEmployeesSummary возвращает суммарную аналитику по сотрудникам
func (h *Handlers) GetSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим даты из query параметров
	var startDate, endDate *time.Time

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		t, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			startDate = &t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		t, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			endDate = &t
		}
	}

	summary, err := h.repository.GetEmployeesSummary(r.Context(), startDate, endDate)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get employees summary: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_ANALYSIS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("path", r.URL.Path),
	)

	core.SendSuccess(w, summary, "Employees summary retrieved successfully.")
}

// GetEmployeesGrouped возвращает динамику изменения штата сотрудников
func (h *Handlers) AnalyseGrouped(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим даты из query параметров
	var startDate, endDate *time.Time

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		t, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			startDate = &t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		t, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			endDate = &t
		}
	}

	// Парсим group_by параметр
	groupByStr := r.URL.Query().Get("group_by")
	groupBy := GroupByDay
	switch groupByStr {
	case "day":
		groupBy = GroupByDay
	case "month":
		groupBy = GroupByMonth
	case "year":
		groupBy = GroupByYear
	default:
		groupBy = GroupByDay
	}

	stats, err := h.repository.GetEmployeesGrouped(r.Context(), startDate, endDate, groupBy)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get employees grouped: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_ANALYSIS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("path", r.URL.Path),
	)

	core.SendSuccess(w, stats, "Employees grouped stats retrieved successfully.")
}

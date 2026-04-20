package dm

import (
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"
	"time"
)

// GetDealAnalysisSummary возвращает сводку по сделкам
func (h *Handlers) GetAnalysisSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var params GetAnalysisParams

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			params.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			params.EndDate = &t
		}
	}
	if typeID := r.URL.Query().Get("type_id"); typeID != "" {
		params.TypeID = &typeID
	}
	if statusID := r.URL.Query().Get("status_id"); statusID != "" {
		params.StatusID = &statusID
	}
	if clientID := r.URL.Query().Get("client_id"); clientID != "" {
		params.ClientID = &clientID
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		params.EmployeeID = &employeeID
	}

	summary, err := h.repository.GetDealAnalysisSummary(r.Context(), params)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deal analysis summary: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("params", params),
	)

	core.SendSuccess(w, summary, "Deal analysis summary retrieved successfully.")
}

// GetDealsGrouped возвращает распределение сделок по указанной группировке
func (h *Handlers) GetAnalysisGrouped(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	groupBy := GroupBy(r.URL.Query().Get("group_by"))
	validGroupBys := map[GroupBy]bool{
		GroupByType:     true,
		GroupByStatus:   true,
		GroupByEmployee: true,
		GroupByClient:   true,
		GroupByDay:      true,
		GroupByMonth:    true,
		GroupByYear:     true,
	}
	if !validGroupBys[groupBy] {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "invalid group_by parameter"),
			logs.WithMetadata("group_by", groupBy),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid group_by. Allowed: type, status, employee, client, day, month, year")
		return
	}

	var params GetAnalysisParams

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			params.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			params.EndDate = &t
		}
	}
	if typeID := r.URL.Query().Get("type_id"); typeID != "" && groupBy != GroupByType {
		params.TypeID = &typeID
	}
	if statusID := r.URL.Query().Get("status_id"); statusID != "" && groupBy != GroupByStatus {
		params.StatusID = &statusID
	}
	if clientID := r.URL.Query().Get("client_id"); clientID != "" && groupBy != GroupByClient {
		params.ClientID = &clientID
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" && groupBy != GroupByEmployee {
		params.EmployeeID = &employeeID
	}

	var stats []GroupedStats
	var err error

	switch {
	case groupBy == GroupByType:
		stats, err = h.repository.GetDealsGroupedByType(r.Context(), params)
	case groupBy == GroupByStatus:
		stats, err = h.repository.GetDealsGroupedByStatus(r.Context(), params)
	case groupBy == GroupByEmployee:
		stats, err = h.repository.GetDealsGroupedByEmployee(r.Context(), params)
	case groupBy == GroupByClient:
		stats, err = h.repository.GetDealsGroupedByClient(r.Context(), params)
	case groupBy == GroupByDay || groupBy == GroupByMonth || groupBy == GroupByYear:
		stats, err = h.repository.GetDealsGroupedByTime(r.Context(), groupBy, params)
	default:
		core.SendValidationError(w, "Invalid group_by")
		return
	}

	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("group_by", groupBy),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deals grouped: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("group_by", groupBy),
		logs.WithMetadata("params", params),
		logs.WithMetadata("result_count", len(stats)),
	)

	core.SendSuccess(w, stats, fmt.Sprintf("Deals grouped by %s retrieved successfully.", groupBy))
}

// GetDealsFinancialSummary возвращает финансовую сводку по всем сделкам
func (h *Handlers) GetAnalysisFinancialSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var params GetAnalysisParams

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			params.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			params.EndDate = &t
		}
	}

	summary, err := h.repository.GetDealsFinancialSummary(r.Context(), params)
	if err != nil {
		if strings.Contains(err.Error(), "no transactions found") {
			h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
				logs.WithStatus(logs.LogStatusSuccess),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("params", params),
				logs.WithMetadata("result", "empty"),
			)
			core.SendSuccess(w, nil, "No financial data available for the selected period.")
			return
		}
		h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deals financial summary: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_ANALYSIS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("params", params),
	)

	core.SendSuccess(w, summary, "Deals financial summary retrieved successfully.")
}

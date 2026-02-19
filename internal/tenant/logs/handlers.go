package logs

import (
	"fmt"
	"kroncl-server/internal/core"
	"net/http"
	"strconv"
	"time"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// GetLog возвращает один лог по ID
func (h *Handlers) GetLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID лога из URL
	logID := r.PathValue("logId")
	if logID == "" {
		core.SendError(w, http.StatusBadRequest, "Log ID is required.")
		return
	}

	log, err := h.service.GetLogByID(r.Context(), logID)
	if err != nil {
		core.SendNotFound(w, "Log not found.")
		return
	}

	core.SendSuccess(w, log, "Log retrieved successfully.")
}

// GetLogs возвращает список логов с пагинацией и фильтрацией
func (h *Handlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Парсим query параметры в структуру фильтров
	var req GetLogsRequest
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	// Account ID filter
	if accountID := r.URL.Query().Get("account_id"); accountID != "" {
		req.AccountID = &accountID
	}

	// Key filter
	if key := r.URL.Query().Get("key"); key != "" {
		req.Key = &key
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		s := LogStatus(status)
		if s != LogStatusSuccess && s != LogStatusError && s != LogStatusPending {
			core.SendValidationError(w, "Invalid status. Use 'success', 'error', or 'pending'.")
			return
		}
		req.Status = &s
	}

	// Criticality filters
	if minCrit := r.URL.Query().Get("min_criticality"); minCrit != "" {
		val, err := strconv.Atoi(minCrit)
		if err == nil && val >= 1 && val <= 10 {
			req.MinCriticality = &val
		}
	}
	if maxCrit := r.URL.Query().Get("max_criticality"); maxCrit != "" {
		val, err := strconv.Atoi(maxCrit)
		if err == nil && val >= 1 && val <= 10 {
			req.MaxCriticality = &val
		}
	}

	// Date filters
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			req.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			req.EndDate = &t
		}
	}

	// Search filter
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	logs, total, err := h.service.GetLogs(r.Context(), req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get logs: %s", err.Error()))
		return
	}

	response := map[string]interface{}{
		"logs": logs,
		"pagination": core.NewPagination(
			int(total),
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Logs retrieved successfully.")
}

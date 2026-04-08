package logs

import (
	"fmt"
	"kroncl-server/internal/config"
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
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID лога из URL
	logID := r.PathValue("logId")
	if logID == "" {
		h.service.Log(r.Context(), config.PERMISSION_LOGS, accountID,
			WithStatus(LogStatusError),
			WithUserAgent(r.UserAgent()),
			WithMetadata("error", "Log ID is required"),
			WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Log ID is required.")
		return
	}

	log, err := h.service.GetLogByID(r.Context(), logID)
	if err != nil {
		h.service.Log(r.Context(), config.PERMISSION_LOGS, accountID,
			WithStatus(LogStatusError),
			WithUserAgent(r.UserAgent()),
			WithMetadata("error", "Log not found"),
			WithMetadata("path", r.URL.Path),
			WithMetadata("log_id", logID),
		)
		core.SendNotFound(w, "Log not found.")
		return
	}

	h.service.Log(r.Context(), config.PERMISSION_LOGS, accountID,
		WithStatus(LogStatusSuccess),
		WithUserAgent(r.UserAgent()),
		WithMetadata("log_id", logID),
	)

	core.SendSuccess(w, log, "Log retrieved successfully.")
}

// GetLogs возвращает список логов с пагинацией и фильтрацией
func (h *Handlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Парсим query параметры в структуру фильтров
	var req GetLogsRequest
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	// Account ID filter
	if accountIDFilter := r.URL.Query().Get("account_id"); accountIDFilter != "" {
		req.AccountID = &accountIDFilter
	}

	// Key filter
	if key := r.URL.Query().Get("key"); key != "" {
		req.Key = &key
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		s := LogStatus(status)
		if s != LogStatusSuccess && s != LogStatusError && s != LogStatusPending {
			h.service.Log(r.Context(), config.PERMISSION_LOGS, accountID,
				WithStatus(LogStatusError),
				WithUserAgent(r.UserAgent()),
				WithMetadata("error", "Invalid status"),
				WithMetadata("status", status),
				WithMetadata("path", r.URL.Path),
			)
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
		h.service.Log(r.Context(), config.PERMISSION_LOGS, accountID,
			WithStatus(LogStatusError),
			WithUserAgent(r.UserAgent()),
			WithMetadata("error", err.Error()),
			WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get logs: %s", err.Error()))
		return
	}

	// Логируем успешный просмотр логов (мета-логирование)
	h.service.Log(r.Context(), config.PERMISSION_LOGS, accountID,
		WithStatus(LogStatusSuccess),
		WithUserAgent(r.UserAgent()),
		WithMetadata("filters", map[string]interface{}{
			"account_id":      req.AccountID,
			"key":             req.Key,
			"status":          req.Status,
			"min_criticality": req.MinCriticality,
			"max_criticality": req.MaxCriticality,
			"start_date":      req.StartDate,
			"end_date":        req.EndDate,
			"search":          req.Search,
		}),
		WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		WithMetadata("result_count", len(logs)),
	)

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

// ClearLogs очищает все логи компании
func (h *Handlers) ClearLogs(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Логируем начало операции очистки
	h.service.Log(r.Context(), config.PERMISSION_LOGS_CLEAR, accountID,
		WithStatus(LogStatusPending),
		WithUserAgent(r.UserAgent()),
		WithMetadata("action", "clear_logs"),
		WithMetadata("path", r.URL.Path),
	)

	err := h.service.сlearLogs(r.Context())
	if err != nil {
		// Логируем ошибку
		h.service.Log(r.Context(), config.PERMISSION_LOGS_CLEAR, accountID,
			WithStatus(LogStatusError),
			WithUserAgent(r.UserAgent()),
			WithMetadata("error", err.Error()),
			WithMetadata("action", "clear_logs"),
			WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to clear logs: %s", err.Error()))
		return
	}

	// Логируем успешную очистку
	h.service.Log(r.Context(), config.PERMISSION_LOGS_CLEAR, accountID,
		WithStatus(LogStatusSuccess),
		WithUserAgent(r.UserAgent()),
		WithMetadata("action", "clear_logs"),
		WithMetadata("path", r.URL.Path),
	)

	core.SendSuccess(w, nil, "Logs cleared successfully.")
}

// handlers.go - хэндлер для оптимизации
// OptimizeLogs удаляет логи, которые хранятся дольше оптимального периода
func (h *Handlers) OptimizeLogs(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Логируем начало операции оптимизации
	h.service.Log(r.Context(), config.PERMISSION_LOGS_OPTIMIZE, accountID,
		WithStatus(LogStatusPending),
		WithUserAgent(r.UserAgent()),
		WithMetadata("action", "optimize_logs"),
		WithMetadata("path", r.URL.Path),
	)

	err := h.service.optimizeLogs(r.Context())
	if err != nil {
		// Логируем ошибку
		h.service.Log(r.Context(), config.PERMISSION_LOGS_OPTIMIZE, accountID,
			WithStatus(LogStatusError),
			WithUserAgent(r.UserAgent()),
			WithMetadata("error", err.Error()),
			WithMetadata("action", "optimize_logs"),
			WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to optimize logs: %s", err.Error()))
		return
	}

	// Логируем успешную оптимизацию
	h.service.Log(r.Context(), config.PERMISSION_LOGS_OPTIMIZE, accountID,
		WithStatus(LogStatusSuccess),
		WithUserAgent(r.UserAgent()),
		WithMetadata("action", "optimize_logs"),
		WithMetadata("path", r.URL.Path),
	)

	core.SendSuccess(w, nil, "Logs optimized successfully.")
}

// handlers.go
// GetLogsActivity возвращает активность по дням
func (h *Handlers) GetLogsActivity(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим параметры дат
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

	activities, err := h.service.GetLogsActivity(r.Context(), startDate, endDate)
	if err != nil {
		h.service.Log(r.Context(), config.PERMISSION_LOGS_ACTIVITY, accountID,
			WithStatus(LogStatusError),
			WithUserAgent(r.UserAgent()),
			WithMetadata("error", err.Error()),
			WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get logs activity: %s", err.Error()))
		return
	}

	h.service.Log(r.Context(), config.PERMISSION_LOGS_ACTIVITY, accountID,
		WithStatus(LogStatusSuccess),
		WithUserAgent(r.UserAgent()),
		WithMetadata("path", r.URL.Path),
	)

	core.SendSuccess(w, activities, "Logs activity retrieved successfully.")
}

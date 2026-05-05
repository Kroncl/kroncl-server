package adminclientele

import (
	"fmt"
	"kroncl-server/internal/core"
	"net/http"
	"strconv"
	"time"
)

func (h *Handlers) GetClienteleStats(w http.ResponseWriter, r *http.Request) {
	// Получаем последний снапшот из истории
	limit := 1
	response, err := h.service.metricsService.GetClienteleMetricsHistory(r.Context(), nil, nil, limit)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get clientele stats: %v", err))
		return
	}

	if len(response) == 0 {
		core.SendSuccess(w, nil, "No clientele metrics available yet")
		return
	}

	core.SendSuccess(w, response[0], "Clientele statistics.")
}

func (h *Handlers) GetClienteleHistory(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.service.metricsService.GetClienteleMetricsHistory(r.Context(), startDate, endDate, limit)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get clientele history: %v", err))
		return
	}

	core.SendSuccess(w, response, "Clientele metrics history.")
}

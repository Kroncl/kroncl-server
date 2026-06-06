package fm

import (
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"time"
)

func (h *Handlers) GetForecastTimeline(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ForecastRequest

	// Парсим query параметры
	if method := r.URL.Query().Get("method"); method != "" {
		req.Method = ForecastMethod(method)
	} else {
		req.Method = ForecastMethodTheta // дефолт
	}

	if horizon := r.URL.Query().Get("horizon"); horizon != "" {
		fmt.Sscanf(horizon, "%d", &req.Horizon)
	}

	// Опциональные даты
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := parseTime(startDate)
		if err == nil {
			req.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := parseTime(endDate)
		if err == nil {
			req.EndDate = &t
		}
	}

	// Вызываем метод прогноза
	var response *ForecastResponse
	var err error

	switch req.Method {
	case ForecastMethodTheta:
		response, err = h.repository.ThetaForecast(r.Context(), req)
	default:
		h.logsService.Log(r.Context(), config.PERMISSION_FM_FORECAST_TIMELINE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid forecast method"),
			logs.WithMetadata("method", req.Method),
		)
		core.SendValidationError(w, "Invalid forecast method. Use 'theta'.")
		return
	}

	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_FORECAST_TIMELINE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)
		core.SendInternalError(w, fmt.Sprintf("Forecast failed: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_FORECAST_TIMELINE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("method", response.Method),
		logs.WithMetadata("data_points", response.DataPoints),
		logs.WithMetadata("horizon", response.Horizon),
		logs.WithMetadata("confidence", response.Confidence),
	)

	core.SendSuccess(w, response, "Forecast generated successfully.")
}

func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}

func (h *Handlers) GetForecastSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ForecastSummaryRequest

	if method := r.URL.Query().Get("method"); method != "" {
		req.Method = ForecastMethod(method)
	} else {
		req.Method = ForecastMethodTheta
	}

	if horizon := r.URL.Query().Get("horizon"); horizon != "" {
		fmt.Sscanf(horizon, "%d", &req.Horizon)
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := parseTime(startDate)
		if err == nil {
			req.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := parseTime(endDate)
		if err == nil {
			req.EndDate = &t
		}
	}

	response, err := h.repository.ThetaForecastSummary(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_FORECAST_SUMMARY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)
		core.SendInternalError(w, fmt.Sprintf("Forecast summary failed: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_FORECAST_SUMMARY, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("method", response.Method),
		logs.WithMetadata("horizon", response.Horizon),
		logs.WithMetadata("predicted_balance", response.PredictedBalance),
	)

	core.SendSuccess(w, response, "Forecast summary generated successfully.")
}

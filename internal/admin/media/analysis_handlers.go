package adminmedia

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"kroncl-server/internal/core"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetSystemStats(r.Context())
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get system stats: %v", err))
		return
	}

	core.SendSuccess(w, stats, "System media stats.")
}

func (h *Handlers) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var startDate *time.Time
	if sd := query.Get("start_date"); sd != "" {
		t, err := time.Parse(time.RFC3339, sd)
		if err != nil {
			core.SendValidationError(w, "Invalid start_date format, use RFC3339")
			return
		}
		startDate = &t
	}

	var endDate *time.Time
	if ed := query.Get("end_date"); ed != "" {
		t, err := time.Parse(time.RFC3339, ed)
		if err != nil {
			core.SendValidationError(w, "Invalid end_date format, use RFC3339")
			return
		}
		endDate = &t
	}

	limit := 100
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

	metrics, err := h.service.GetMetricsHistory(r.Context(), startDate, endDate, limit)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get metrics history: %v", err))
		return
	}

	core.SendSuccess(w, metrics, "Media metrics history.")
}

func (h *Handlers) GetBuckets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	search := query.Get("search")
	params := core.GetPaginationParams(r, 20, 100)

	response, err := h.service.GetBuckets(r.Context(), search, params)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get buckets: %v", err))
		return
	}

	core.SendSuccess(w, response, "Buckets list.")
}

func (h *Handlers) GetBucket(w http.ResponseWriter, r *http.Request) {
	bucketName := chi.URLParam(r, "bucketId")
	if bucketName == "" {
		core.SendValidationError(w, "bucketId is required")
		return
	}

	bucket, err := h.service.GetBucket(r.Context(), bucketName)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get bucket: %v", err))
		return
	}

	core.SendSuccess(w, bucket, "Bucket info.")
}

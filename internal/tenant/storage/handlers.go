package storage

import (
	"kroncl-server/internal/core"
	"net/http"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) GetStorageSummary(w http.ResponseWriter, r *http.Request) {
	companyID, ok := core.GetCompanyIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusBadRequest, "Company context not found")
		return
	}

	dbStatus, err := h.service.Db.GetStorageStatus(r.Context(), companyID)
	if err != nil {
		core.SendInternalError(w, "Failed to get database storage status")
		return
	}

	bucketInfo, err := h.service.Media.GetBucketStatus(r.Context(), companyID)
	if err != nil {
		core.SendInternalError(w, "Failed to get media bucket status")
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"database": dbStatus,
		"media":    bucketInfo,
	}, "Storage summary retrieved successfully")
}

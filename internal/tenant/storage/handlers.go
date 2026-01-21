package storage

import (
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// получение хранилища (полный объект)
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем company_id из URL параметра
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	storage, err := h.service.repository.GetStorageByCompanyID(r.Context(), companyID)
	if err != nil {
		core.SendValidationError(w, "Company storage was not initialized.")
		return
	}

	core.SendSuccess(w, storage, "Company storage retrieved successfully.")
}

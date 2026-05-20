package storagedb

import (
	"fmt"
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

// получение ресурсов хранилища
func (h *Handlers) GetSources(w http.ResponseWriter, r *http.Request) {
	// Получаем company_id из URL параметра
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	storage, err := h.service.repository.GetStorageByCompanyID(r.Context(), companyID)
	if err != nil {
		core.SendValidationError(w, "Company DB storage was not initialized.")
		return
	}

	sources, err := h.service.repository.GetStorageSources(r.Context(), storage.SchemaName)
	if err != nil {
		core.SendValidationError(w, fmt.Sprintf("Failed get company DB storage sources stat: %s", err.Error()))
		return
	}

	core.SendSuccess(w, sources, "Company DB storage sources retrieved successfully.")
}

// получение хранилища (полный объект)
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	// Получаем company_id из URL параметра
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	storage, err := h.service.repository.GetStorageByCompanyID(r.Context(), companyID)
	if err != nil {
		core.SendValidationError(w, "Company DB storage was not initialized.")
		return
	}

	core.SendSuccess(w, storage, "Company DB storage retrieved successfully.")
}

func (h *Handlers) GetByModules(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	result, err := h.service.GetStorageByModules(r.Context(), companyID)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	core.SendSuccess(w, result, "Module DB storage statistics retrieved successfully.")
}

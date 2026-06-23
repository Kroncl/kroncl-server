package currency

import (
	"kroncl-server/internal/core"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) GetAll(w http.ResponseWriter, r *http.Request) {
	var codes []string
	if codesParam := r.URL.Query().Get("ids"); codesParam != "" {
		codes = strings.Split(codesParam, ",")
	}

	currencies, err := h.service.GetAll(r.Context(), codes)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, "Failed to get currencies")
		return
	}

	core.SendSuccess(w, currencies, "Currencies retrieved successfully")
}

func (h *Handlers) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		core.SendError(w, http.StatusBadRequest, "Currency ID is required")
		return
	}

	currency, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		core.SendError(w, http.StatusNotFound, "Currency not found")
		return
	}

	core.SendSuccess(w, currency, "Currency retrieved successfully")
}

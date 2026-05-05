package admincompanies

import (
	"fmt"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetAllCompanies(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	params := core.GetDefaultPaginationParams(r)

	companies, pagination, err := h.service.GetAllCompanies(r.Context(), search, params)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get companies: %v", err))
		return
	}

	response := map[string]interface{}{
		"companies":  companies,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Companies list.")
}

func (h *Handlers) GetCompanyByID(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "companyId")
	if companyID == "" {
		core.SendValidationError(w, "companyId is required")
		return
	}

	company, err := h.service.GetCompanyByID(r.Context(), companyID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get company: %v", err))
		return
	}

	core.SendSuccess(w, company, "Company details.")
}

package adminpartners

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/core"
	"kroncl-server/internal/public"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetAllPartners(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var status *string
	if s := query.Get("status"); s != "" {
		status = &s
	}

	search := query.Get("search")
	params := core.GetDefaultPaginationParams(r)

	partners, pagination, err := h.service.GetAllPartners(r.Context(), status, search, params)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get partners: %v", err))
		return
	}

	response := map[string]interface{}{
		"partners":   partners,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Partners list.")
}

func (h *Handlers) GetPartnerByID(w http.ResponseWriter, r *http.Request) {
	partnerID := chi.URLParam(r, "partnerId")
	if partnerID == "" {
		core.SendValidationError(w, "partnerId is required")
		return
	}

	partner, err := h.service.GetPartnerByID(r.Context(), partnerID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get partner: %v", err))
		return
	}

	core.SendSuccess(w, partner, "Partner details.")
}

func (h *Handlers) UpdatePartner(w http.ResponseWriter, r *http.Request) {
	partnerID := chi.URLParam(r, "partnerId")
	if partnerID == "" {
		core.SendValidationError(w, "partnerId is required")
		return
	}

	var req public.UpdateIncomingPartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	partner, err := h.service.UpdatePartner(r.Context(), partnerID, &req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to update partner: %v", err))
		return
	}

	core.SendSuccess(w, partner, "Partner updated.")
}

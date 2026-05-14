package adminpricing

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetPromocodes(w http.ResponseWriter, r *http.Request) {
	params := core.GetDefaultPaginationParams(r)

	promocodes, pagination, err := h.service.GetPromocodes(r.Context(), params)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get promocodes: %v", err))
		return
	}

	response := map[string]interface{}{
		"promocodes": promocodes,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Promocodes retrieved successfully")
}

func (h *Handlers) GetPromocodeByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "promocodeId")
	if id == "" {
		core.SendValidationError(w, "Promocode ID is required")
		return
	}

	promocode, err := h.service.GetPromocodeByID(r.Context(), id)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Promocode not found: %v", err))
		return
	}

	core.SendSuccess(w, promocode, "Promocode retrieved successfully")
}

func (h *Handlers) GetPromocodeByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "promocode")
	if code == "" {
		core.SendValidationError(w, "Promocode code is required")
		return
	}

	promocode, err := h.service.GetPromocodeByCode(r.Context(), code)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Promocode not found: %v", err))
		return
	}

	core.SendSuccess(w, promocode, "Promocode retrieved successfully")
}

func (h *Handlers) CreatePromocode(w http.ResponseWriter, r *http.Request) {
	var req CreatePromocodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	promocode, err := h.service.CreatePromocode(r.Context(), &req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to create promocode: %v", err))
		return
	}

	core.SendSuccess(w, promocode, "Promocode created successfully")
}

func (h *Handlers) UpdatePromocode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "promocodeId")
	if id == "" {
		core.SendValidationError(w, "Promocode ID is required")
		return
	}

	var req UpdatePromocodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	promocode, err := h.service.UpdatePromocode(r.Context(), id, &req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to update promocode: %v", err))
		return
	}

	core.SendSuccess(w, promocode, "Promocode updated successfully")
}

func (h *Handlers) DeletePromocode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "promocodeId")
	if id == "" {
		core.SendValidationError(w, "Promocode ID is required")
		return
	}

	err := h.service.DeletePromocode(r.Context(), id)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to delete promocode: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Promocode deleted successfully")
}

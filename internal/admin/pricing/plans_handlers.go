package adminpricing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"kroncl-server/internal/core"
)

func (h *Handlers) GetPlans(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	plans, total, err := h.service.GetPlans(r.Context(), page, limit, search)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get plans: %v", err))
		return
	}

	pages := total / limit
	if total%limit > 0 {
		pages++
	}

	response := map[string]interface{}{
		"plans": plans,
		"pagination": map[string]interface{}{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		},
	}

	core.SendSuccess(w, response, "Pricing plans retrieved successfully")
}

func (h *Handlers) GetPlan(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("planCode")
	if code == "" {
		core.SendValidationError(w, "Plan code is required")
		return
	}

	plan, err := h.service.GetPlanByCode(r.Context(), code)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Plan not found: %v", err))
		return
	}

	core.SendSuccess(w, plan, "Plan retrieved successfully")
}

func (h *Handlers) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("planCode")
	if code == "" {
		core.SendValidationError(w, "Plan code is required")
		return
	}

	var req UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	plan, err := h.service.UpdatePlan(r.Context(), code, req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to update plan: %v", err))
		return
	}

	core.SendSuccess(w, plan, "Plan updated successfully")
}

package adminaccounts

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	params := core.GetDefaultPaginationParams(r)

	accounts, pagination, err := h.service.GetAllAccounts(r.Context(), search, params)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get accounts: %v", err))
		return
	}

	response := map[string]interface{}{
		"accounts":   accounts,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Accounts list.")
}

func (h *Handlers) GetUserStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetUserStats(r.Context())
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get user stats: %v", err))
		return
	}

	core.SendSuccess(w, stats, "User statistics.")
}

func (h *Handlers) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		core.SendValidationError(w, "accountId is required")
		return
	}

	account, err := h.service.accountsService.GetByID(r.Context(), accountID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get account: %v", err))
		return
	}

	core.SendSuccess(w, account, "Account details.")
}

func (h *Handlers) PromoteToAdmin(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		core.SendValidationError(w, "accountId is required")
		return
	}

	caller, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Unauthorized")
		return
	}

	var req PromoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	if req.Level < config.ADMIN_LEVEL_MIN || req.Level > config.ADMIN_LEVEL_MAX {
		core.SendValidationError(w, fmt.Sprintf("Level must be between %d and %d", config.ADMIN_LEVEL_MIN, config.ADMIN_LEVEL_MAX))
		return
	}

	err := h.service.PromoteToAdmin(r.Context(), caller.UserID, accountID, req.Level)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to promote: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Account promoted to admin successfully")
}

func (h *Handlers) DemoteFromAdmin(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		core.SendValidationError(w, "accountId is required")
		return
	}

	caller, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Unauthorized")
		return
	}

	err := h.service.DemoteFromAdmin(r.Context(), caller.UserID, accountID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to demote: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Account demoted from admin successfully")
}

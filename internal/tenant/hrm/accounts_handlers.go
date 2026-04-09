package hrm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// GetAccountSettings возвращает настройки аккаунта
func (h *Handlers) GetAccountSettings(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Company ID required")
		return
	}

	targetAccountID := chi.URLParam(r, "accountId")
	if targetAccountID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Account ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Account ID required")
		return
	}

	// Проверяем принадлежность к компании
	member, err := h.repository.companiesService.GetUserCompanyById(r.Context(), accountID, companyID)
	if err != nil || member == nil {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company not found or access denied"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendNotFound(w, "Company not found or access denied")
		return
	}

	settings, err := h.repository.GetAccountSettings(r.Context(), targetAccountID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get account settings: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("path", r.URL.Path),
		logs.WithMetadata("target_account_id", targetAccountID),
	)

	core.SendSuccess(w, settings, "Account settings retrieved successfully.")
}

// UpdateAccountSettings обновляет настройки аккаунта
func (h *Handlers) UpdateAccountSettings(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Company ID required")
		return
	}

	targetAccountID := chi.URLParam(r, "accountId")
	if targetAccountID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Account ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Account ID required")
		return
	}

	// Проверяем принадлежность к компании
	member, err := h.repository.companiesService.GetUserCompanyById(r.Context(), accountID, companyID)
	if err != nil || member == nil {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company not found or access denied"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendNotFound(w, "Company not found or access denied")
		return
	}

	var req UpdateAccountSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	settings, err := h.repository.UpsertAccountSettings(r.Context(), targetAccountID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "appears in both"):
			h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, errorMsg)
		case strings.Contains(errorMsg, "invalid"):
			h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update account settings: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("path", r.URL.Path),
		logs.WithMetadata("target_account_id", targetAccountID),
	)

	core.SendSuccess(w, settings, "Account settings updated successfully.")
}

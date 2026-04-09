package hrm

import (
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetAccountPermissions возвращает все разрешения целевого аккаунта в компании
func (h *Handlers) GetAccountPermissions(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company not found or access denied"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendNotFound(w, "Company not found or access denied")
		return
	}

	permissions, err := h.repository.GetAccountPermissions(r.Context(), companyID, targetAccountID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get account permissions: %s", err.Error()))
		return
	}

	// Преобразуем map в список для удобства клиента
	type PermissionItem struct {
		Code string `json:"code"`
	}

	items := make([]PermissionItem, 0, len(permissions))
	for code := range permissions {
		items = append(items, PermissionItem{Code: code})
	}

	h.logsService.Log(r.Context(), config.PERMISSION_ACCOUNTS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("path", r.URL.Path),
		logs.WithMetadata("target_account_id", targetAccountID),
		logs.WithMetadata("permissions_count", len(items)),
	)

	core.SendSuccess(w, items, "Account permissions retrieved successfully.")
}

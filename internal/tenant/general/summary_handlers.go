package tenantgeneral

import (
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
)

func (h *Handlers) GetCompanySummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetCurrency := r.URL.Query().Get("currency")
	if targetCurrency == "" {
		targetCurrency = "RUB"
	}

	summary, err := h.service.GetCompanySummary(r.Context(), targetCurrency)

	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_COMPANY_SUMMARY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get company summary: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_COMPANY_SUMMARY, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("path", r.URL.Path),
	)

	core.SendSuccess(w, summary, "Company summary retrieved successfully.")
}

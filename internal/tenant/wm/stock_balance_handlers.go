package wm

import (
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
)

func (h *Handlers) GetStockBalance(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	unitID := r.URL.Query().Get("unit_id")
	var unitIDPtr *string
	if unitID != "" {
		unitIDPtr = &unitID
	}

	balances, err := h.repository.GetStockBalance(r.Context(), unitIDPtr)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get stock balance: %s", err.Error()))
		return
	}

	core.SendSuccess(w, balances, "Stock balance retrieved successfully.")
}

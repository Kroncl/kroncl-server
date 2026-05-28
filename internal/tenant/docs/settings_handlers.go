package docs

import (
	"encoding/json"
	"net/http"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
)

func (h *Handlers) GetSettings(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DOCS_SETTINGS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)
		core.SendInternalError(w, "Failed to get document settings")
		return
	}

	core.SendSuccess(w, settings, "Document settings retrieved successfully")
}

func (h *Handlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateDocsSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DOCS_SETTINGS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	settings, err := h.service.UpdateSettings(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DOCS_SETTINGS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)
		core.SendInternalError(w, "Failed to update document settings")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DOCS_SETTINGS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
	)

	core.SendSuccess(w, settings, "Document settings updated successfully")
}

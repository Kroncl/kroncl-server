package crm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
)

func (h *Handlers) GenerateFullReport(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Types   []string `json:"types"`
		Comment *string  `json:"comment,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_REPORT, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Types) == 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_REPORT, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Types array is required"),
		)
		core.SendError(w, http.StatusBadRequest, "At least one report type is required")
		return
	}

	opts := FullReportOptions{
		Types:   req.Types,
		Comment: req.Comment,
	}

	doc, err := h.repository.GenerateFullReport(r.Context(), opts)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_REPORT, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("types", req.Types),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to generate report: %s", err.Error()))
		return
	}

	presignedURL, err := h.repository.mediaService.GeneratePresignedURL(r.Context(), doc.ObjectPath, 1*time.Hour)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_REPORT, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)
		core.SendInternalError(w, "Failed to generate download URL")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_REPORT, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("doc_id", doc.ID),
		logs.WithMetadata("types", req.Types),
	)

	core.SendSuccess(w, map[string]interface{}{
		"download_url": presignedURL,
		"doc":          doc,
	}, "Full report generated successfully")
}

package dm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
)

func (h *Handlers) GenerateDealInvoice(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dealID := r.PathValue("dealId")
	if dealID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required.")
		return
	}

	var req GenerateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if len(req.Positions) == 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "At least one position is required"),
		)
		core.SendError(w, http.StatusBadRequest, "At least one position is required")
		return
	}

	if req.TotalAmount <= 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Total amount must be greater than 0"),
		)
		core.SendError(w, http.StatusBadRequest, "Total amount must be greater than 0")
		return
	}

	doc, err := h.repository.GenerateDealInvoice(r.Context(), dealID, req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("deal_id", req.DealID),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to generate invoice: %s", err.Error()))
		return
	}

	presignedURL, err := h.repository.mediaService.GeneratePresignedURL(r.Context(), doc.ObjectPath, 1*time.Hour)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)
		core.SendInternalError(w, "Failed to generate download URL")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("doc_id", doc.ID),
		logs.WithMetadata("deal_id", req.DealID),
	)

	core.SendSuccess(w, map[string]interface{}{
		"download_url": presignedURL,
		"doc":          doc,
	}, "Invoice generated successfully")
}

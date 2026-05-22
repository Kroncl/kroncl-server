package dm

import (
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
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required")
		return
	}

	// var req struct {
	// 	Comment *string `json:"comment,omitempty"`
	// }

	// Парсим тело запроса (опционально)
	if r.Body != http.NoBody {
		// можно распарсить, если нужно
	}

	comment := r.URL.Query().Get("comment")
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	doc, err := h.repository.GenerateDealInvoice(r.Context(), dealID, commentPtr)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_INVOICE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("deal_id", dealID),
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
		logs.WithMetadata("deal_id", dealID),
	)

	core.SendSuccess(w, map[string]interface{}{
		"download_url": presignedURL,
		"doc":          doc,
	}, "Invoice generated successfully")
}

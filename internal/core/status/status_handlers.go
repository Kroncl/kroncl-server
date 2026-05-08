package corestatus

import (
	"kroncl-server/internal/core"
	"net/http"
	"strconv"
)

func (h *Handlers) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	status, err := h.service.GetFullSystemStatus(r.Context(), days)
	if err != nil {
		core.SendInternalError(w, "Failed to get system status")
		return
	}

	core.SendSuccess(w, status, "System status retrieved")
}

package adminauth

import (
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"
)

func (h *Handlers) CheckAdmin(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// reverify admin
	isAdmin, adminLevel, err := h.service.GetAdminStatus(r.Context(), user.UserID)
	if err != nil {
		core.SendInternalError(w, "Failed to verify admin status")
		return
	}

	if !isAdmin {
		core.SendError(w, http.StatusForbidden, "Admin access required")
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"is_admin":    true,
		"admin_level": adminLevel,
		"account_id":  user.UserID,
	}, "Admin access confirmed")
}

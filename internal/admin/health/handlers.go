package adminhealth

import (
	"kroncl-server/internal/core"
	"net/http"
)

func SendResult(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	core.SendSuccess(w, accountID, "Welcome home.")
}

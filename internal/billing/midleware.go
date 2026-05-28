package billing

import (
	"net/http"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
)

func BillingRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.IsBillingEnabled() {
			core.SendError(w, http.StatusServiceUnavailable, "The service currently operates without payment")
			return
		}
		next.ServeHTTP(w, r)
	})
}

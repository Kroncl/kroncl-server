// internal/permissioner/middleware.go
package permissioner

import (
	"kroncl-server/internal/core"
	"log"
	"net/http"
)

func RequirePermission(permService *Service, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Используем core helper'ы
			userID, ok := core.GetUserIDFromContext(r.Context())
			log.Printf("RequirePermission: userID=%s, ok=%v", userID, ok)
			if !ok {
				http.Error(w, "User not authenticated", http.StatusUnauthorized)
				return
			}

			companyID, ok := core.GetCompanyIDFromContext(r.Context())
			log.Printf("RequirePermission: companyID=%s, ok=%v", companyID, ok)
			if !ok {
				http.Error(w, "Company not found in context", http.StatusBadRequest)
				return
			}

			hasPerm, err := permService.Has(r.Context(), userID, companyID, permission)
			if err != nil {
				http.Error(w, "Permission check failed", http.StatusInternalServerError)
				return
			}

			if !hasPerm {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

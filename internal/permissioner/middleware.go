// internal/permissioner/middleware.go
package permissioner

import (
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/storage"
	"log"
	"net/http"
)

type PermissionDeps struct {
	PermService    *Service
	StorageService *storage.Service
}

func RequirePermission(deps *PermissionDeps, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := core.GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "User not authenticated", http.StatusUnauthorized)
				return
			}

			companyID, ok := core.GetCompanyIDFromContext(r.Context())
			if !ok {
				http.Error(w, "Company not found in context", http.StatusBadRequest)
				return
			}

			tenantPool, ok := deps.StorageService.GetTenantPoolFromRequest(r)
			if !ok {
				http.Error(w, "Failed to get tenant connection", http.StatusInternalServerError)
				return
			}

			result, err := deps.PermService.CheckPermissionDetailed(r.Context(), tenantPool, companyID, userID, permission)
			if err != nil {
				log.Printf("Permission check error for user %s, company %s, permission %s: %v", userID, companyID, permission, err)
				http.Error(w, "Permission check failed", http.StatusInternalServerError)
				return
			}

			if !result.Allowed {
				log.Printf("Permission denied for user %s, company %s: %s", userID, companyID, result.Reason)
				http.Error(w, result.Reason, http.StatusForbidden)
				return
			}

			log.Printf("Permission granted for user %s, company %s: %s", userID, companyID, result.Reason)
			next.ServeHTTP(w, r)
		})
	}
}

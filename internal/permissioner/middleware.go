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
			_ = userID // пока не используем, но для логирования пригодится

			companyID, ok := core.GetCompanyIDFromContext(r.Context())
			if !ok {
				http.Error(w, "Company not found in context", http.StatusBadRequest)
				return
			}

			// Проверяем права через сервис (без tenantPool)
			hasPerm, err := deps.PermService.CheckPermission(r.Context(), companyID, permission)
			if err != nil {
				log.Printf("Permission check error: %v", err)
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

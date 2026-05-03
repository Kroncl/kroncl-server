package adminauth

import (
	"context"
	"kroncl-server/internal/auth"
	"net/http"
)

type contextKey string

const adminContextKey contextKey = "admin"

func (s *Service) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		isAdmin, err := s.IsAdmin(r.Context(), user.UserID)
		if err != nil {
			http.Error(w, "Failed to verify admin status", http.StatusInternalServerError)
			return
		}

		if !isAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), adminContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) RequireAdminLevel(requiredLevel int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			adminLevel, err := s.GetAdminLevel(r.Context(), user.UserID)
			if err != nil {
				http.Error(w, "Failed to verify admin level", http.StatusInternalServerError)
				return
			}

			if adminLevel == 0 {
				http.Error(w, "Admin access required", http.StatusForbidden)
				return
			}

			if adminLevel < requiredLevel {
				http.Error(w, "Insufficient admin level", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), adminContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAdminFromContext(ctx context.Context) (*auth.AccessClaims, bool) {
	admin, ok := ctx.Value(adminContextKey).(*auth.AccessClaims)
	return admin, ok
}

package adminauth

import (
	"context"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/config"
	"net/http"
)

type contextKey string

const adminContextKey contextKey = "admin"
const adminKeywordVerified contextKey = "admin_keyword_verified"

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

// RequireAdminKeyword требует подтверждения админского ключевого слова
func (s *Service) RequireAdminKeyword(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что пользователь уже админ
		_, ok := GetAdminFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем заголовок X-Admin-Keyword
		providedKeyword := r.Header.Get(config.ADMIN_KEYWORD_HEADER)
		if providedKeyword == "" {
			http.Error(w, "Admin keyword required", http.StatusForbidden)
			return
		}

		expectedKeyword := config.GetAdminKeyword()
		if expectedKeyword == "" {
			http.Error(w, "Admin keyword not configured", http.StatusInternalServerError)
			return
		}

		if providedKeyword != expectedKeyword {
			http.Error(w, "Invalid admin keyword", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), adminKeywordVerified, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) RequireAdminKeywordLevel(requiredLevel int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Сначала проверяем уровень админа
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

			// Проверяем ключевое слово
			providedKeyword := r.Header.Get(config.ADMIN_KEYWORD_HEADER)
			if providedKeyword == "" {
				http.Error(w, "Admin keyword required", http.StatusForbidden)
				return
			}

			expectedKeyword := config.GetAdminKeyword()
			if expectedKeyword == "" {
				http.Error(w, "Admin keyword not configured", http.StatusInternalServerError)
				return
			}

			if providedKeyword != expectedKeyword {
				http.Error(w, "Invalid admin keyword", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), adminContextKey, user)
			ctx = context.WithValue(ctx, adminKeywordVerified, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAdminKeywordVerified(ctx context.Context) bool {
	verified, ok := ctx.Value(adminKeywordVerified).(bool)
	return ok && verified
}

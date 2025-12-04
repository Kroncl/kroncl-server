package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware проверяет JWT токен в заголовке Authorization
func (s *JWTService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Валидируем access токен (ИСПРАВЛЕНО: используем ValidateAccessToken)
		claims, err := s.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Добавляем claims в контекст
		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext получает пользователя из контекста
func GetUserFromContext(ctx context.Context) (*AccessClaims, bool) {
	user, ok := ctx.Value(userContextKey).(*AccessClaims)
	return user, ok
}

// RequireAuth обертка для маршрутов, требующих аутентификации
func (s *JWTService) RequireAuth(next http.Handler) http.Handler {
	return s.AuthMiddleware(next)
}

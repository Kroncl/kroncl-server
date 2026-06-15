package auth

import (
	"context"
	"kroncl-server/internal/config"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const userContextKey contextKey = "user"

// ApiKeyValidator — интерфейс для проверки API-ключей (реализуется accounts.Service)
type ApiKeyValidator interface {
	ValidateApiKey(ctx context.Context, rawKey string) (*ApiKeyInfo, error)
	UpdateApiKeyLastUsed(ctx context.Context, keyID string) error
}

type ApiKeyInfo struct {
	ID            string
	AccountID     string
	DailyRequests int
	ExpiresAt     *time.Time
	RevokedAt     *time.Time
}

func (s *JWTService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		claims, err := s.ValidateAccessToken(tokenString)
		if err == nil {
			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Если не JWT и есть валидатор API-ключей — пробуем как API-ключ
		if s.apiKeyValidator != nil && strings.HasPrefix(tokenString, config.API_KEY_PREFIX) {
			apiKey, err := s.apiKeyValidator.ValidateApiKey(r.Context(), tokenString)
			if err != nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Обновляем last_used_at асинхронно
			go s.apiKeyValidator.UpdateApiKeyLastUsed(context.Background(), apiKey.ID)

			// Создаём JWT с ApiKeyID
			jwtToken, err := s.GenerateApiAccessToken(apiKey.AccountID, apiKey.ID)
			if err != nil {
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)
				return
			}

			// Валидируем созданный JWT и кладём в контекст
			claims, err := s.ValidateAccessToken(jwtToken)
			if err != nil {
				http.Error(w, "Failed to validate generated token", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
	})
}

func GetUserFromContext(ctx context.Context) (*AccessClaims, bool) {
	user, ok := ctx.Value(userContextKey).(*AccessClaims)
	return user, ok
}

func (s *JWTService) RequireAuth(next http.Handler) http.Handler {
	return s.AuthMiddleware(next)
}

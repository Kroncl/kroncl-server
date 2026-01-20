package core

import (
	"context"
	"kroncl-server/internal/auth"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	CompanyIDKey contextKey = "company_id"
)

func GetCompanyIDFromContext(ctx context.Context) (string, bool) {
	companyID, ok := ctx.Value(CompanyIDKey).(string)
	return companyID, ok
}

func SetCompanyIDInContext(ctx context.Context, companyID string) context.Context {
	return context.WithValue(ctx, CompanyIDKey, companyID)
}

// GetUserIDFromContext получает user_id из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	// Пробуем из auth пакета
	if user, ok := auth.GetUserFromContext(ctx); ok {
		return user.UserID, true
	}
	return "", false
}

// GetClaimsFromContext получает полные claims
func GetClaimsFromContext(ctx context.Context) (*auth.AccessClaims, bool) {
	return auth.GetUserFromContext(ctx)
}

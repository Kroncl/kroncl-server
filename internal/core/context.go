package core

import (
	"context"
	"kroncl-server/internal/auth"
)

// GetUserIDFromContext получает user_id из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	// Пробуем из auth пакета
	if user, ok := auth.GetUserFromContext(ctx); ok {
		return user.ID, true
	}
	return "", false
}

// GetCompanyIDFromContext получает company_id из контекста
func GetCompanyIDFromContext(ctx context.Context) (string, bool) {
	if val := ctx.Value("company_id"); val != nil {
		if companyID, ok := val.(string); ok {
			return companyID, true
		}
	}
	return "", false
}

// GetClaimsFromContext получает полные claims
func GetClaimsFromContext(ctx context.Context) (*auth.AccessClaims, bool) {
	return auth.GetUserFromContext(ctx)
}

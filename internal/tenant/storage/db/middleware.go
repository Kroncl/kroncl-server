package storagedb

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const tenantPoolKey contextKey = "tenant_pool"

func (s *Service) GetTenantPoolFromContext(ctx context.Context) (*pgxpool.Pool, error) {
	companyID, ok := core.GetCompanyIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("company ID not found in context")
	}

	return s.GetTenantPool(ctx, companyID)
}

func (s *Service) TenantPoolMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := core.GetCompanyIDFromContext(r.Context())
		if !ok {
			core.SendError(w, http.StatusBadRequest, "Company context not found")
			return
		}

		// Получаем или создаём пул
		tenantPool, err := s.GetTenantPool(r.Context(), companyID)
		if err != nil {
			core.SendError(w, http.StatusInternalServerError, "Failed to get company storage")
			return
		}

		// Проверяем, что пул жив
		if err := tenantPool.Ping(r.Context()); err != nil {
			// Удаляем битый пул из кэша
			s.tenantPools.Delete(companyID)
			core.SendError(w, http.StatusInternalServerError, "Storage connection lost")
			return
		}

		// Добавляем в контекст
		ctx := context.WithValue(r.Context(), tenantPoolKey, tenantPool)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) GetTenantPoolFromRequest(r *http.Request) (*pgxpool.Pool, bool) {
	pool, ok := r.Context().Value(tenantPoolKey).(*pgxpool.Pool)
	return pool, ok
}

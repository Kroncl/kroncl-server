package storagemedia

import (
	"context"
	"net/http"

	"kroncl-server/internal/core"
)

type contextKey string

const tenantBucketKey contextKey = "tenant_bucket"

func (s *Service) TenantBucketMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := core.GetCompanyIDFromContext(r.Context())
		if !ok {
			core.SendError(w, http.StatusBadRequest, "Company context not found")
			return
		}

		bucketName := s.GetTenantBucketName(companyID)
		ctx := context.WithValue(r.Context(), tenantBucketKey, bucketName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) GetTenantBucketName(tenantID string) string {
	return "tenant-" + tenantID
}

func (s *Service) GetBucketFromContext(ctx context.Context) (string, bool) {
	bucket, ok := ctx.Value(tenantBucketKey).(string)
	return bucket, ok
}

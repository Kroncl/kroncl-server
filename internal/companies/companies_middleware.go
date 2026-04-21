package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CompanyMembership(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			companyID := chi.URLParam(r, "id")

			if companyID == "" {
				core.SendValidationError(w, "Company ID is required")
				return
			}

			if _, err := uuid.Parse(companyID); err != nil {
				core.SendValidationError(w, "Invalid company ID format")
				return
			}

			userID, ok := core.GetUserIDFromContext(r.Context())

			if !ok {
				core.SendUnauthorized(w, "User not authenticated")
				return
			}

			isMember, err := checkCompanyMembership(r.Context(), pool, companyID, userID)
			if err != nil {
				core.SendInternalError(w, fmt.Sprintf("Failed to check membership: %v", err))
				return
			}

			if !isMember {
				core.SendForbidden(w, "Not a member of this company")
				return
			}

			ctx := core.SetCompanyIDInContext(r.Context(), companyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func checkCompanyMembership(ctx context.Context, pool *pgxpool.Pool, companyID, userID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM companies c
			JOIN company_accounts ca ON c.id = ca.company_id
			WHERE c.id = $1 AND ca.account_id = $2
		)
	`

	err := pool.QueryRow(ctx, query, companyID, userID).Scan(&exists)
	return exists, err
}

func checkCompanyExists(ctx context.Context, pool *pgxpool.Pool, companyID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`

	err := pool.QueryRow(ctx, query, companyID).Scan(&exists)
	return exists, err
}

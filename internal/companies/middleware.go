package companies

import (
	"context"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CompanyMembership(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			companyID := chi.URLParam(r, "id")
			if companyID == "" {
				http.Error(w, "Company ID is required", http.StatusBadRequest)
				return
			}

			// Получаем user_id из контекста
			userID, ok := core.GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "User not authenticated", http.StatusUnauthorized)
				return
			}

			// Проверяем что компания существует И пользователь в ней
			isMember, err := checkCompanyMembership(r.Context(), pool, companyID, userID)
			if err != nil {
				http.Error(w, "Failed to check membership", http.StatusInternalServerError)
				return
			}

			if !isMember {
				http.Error(w, "Not a member of this company", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), "company_id", companyID)
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

// checkCompanyExists проверяет существование компании
func checkCompanyExists(ctx context.Context, pool *pgxpool.Pool, companyID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`

	err := pool.QueryRow(ctx, query, companyID).Scan(&exists)
	return exists, err
}

package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CompanyMembership(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			companyID := chi.URLParam(r, "id")
			log.Printf("CompanyMembership DEBUG: companyID from URL = '%s'", companyID)

			if companyID == "" {
				core.SendValidationError(w, "Company ID is required")
				return
			}

			// Получаем user_id из контекста
			userID, ok := core.GetUserIDFromContext(r.Context())
			log.Printf("CompanyMembership DEBUG: userID from context = '%s', ok = %v", userID, ok)

			if !ok {
				core.SendUnauthorized(w, "User not authenticated")
				return
			}

			log.Printf("CompanyMembership DEBUG: Checking membership for userID='%s', companyID='%s'", userID, companyID)

			// Проверяем что компания существует И пользователь в ней
			isMember, err := checkCompanyMembership(r.Context(), pool, companyID, userID)
			if err != nil {
				log.Printf("CompanyMembership ERROR: %v", err)
				core.SendInternalError(w, fmt.Sprintf("Failed to check membership: %v", err))
				return
			}

			log.Printf("CompanyMembership DEBUG: isMember = %v", isMember)

			if !isMember {
				core.SendUnauthorized(w, "Not a member of this company")
				return
			}

			// Добавляем company_id в контекст используя core.SetCompanyIDInContext
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

// checkCompanyExists проверяет существование компании
func checkCompanyExists(ctx context.Context, pool *pgxpool.Pool, companyID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`

	err := pool.QueryRow(ctx, query, companyID).Scan(&exists)
	return exists, err
}

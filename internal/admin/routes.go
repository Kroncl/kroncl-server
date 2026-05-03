package admin

import (
	adminaccounts "kroncl-server/internal/admin/accounts"
	adminauth "kroncl-server/internal/admin/auth"
	admindb "kroncl-server/internal/admin/db"
	adminhealth "kroncl-server/internal/admin/health"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/config"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type Deps struct {
	JWTService            *auth.JWTService
	AdminAuthService      *adminauth.Service
	AdminAuthHandlers     *adminauth.Handlers
	AdminDbHandlers       *admindb.Handlers
	AdminAccountsService  *adminaccounts.Service
	AdminAccountsHandlers *adminaccounts.Handlers
}

func NewRoutes(deps Deps) chi.Router {
	r := chi.NewRouter()

	r.Use(deps.JWTService.RequireAuth)
	r.Use(httprate.LimitByIP(config.RATE_LIMIT_PRIVATE_ROUTES_PER_MINUTE, 1*time.Minute))

	r.Group(func(r chi.Router) {
		r.Use(deps.AdminAuthService.RequireAdmin)
		r.Get("/health", adminhealth.SendResult)
		r.Get("/check", deps.AdminAuthHandlers.CheckAdmin)

		// db-condition
		r.Route("/db", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_1))

			r.Get("/sys", deps.AdminDbHandlers.GetSystemStats)
			r.Get("/history", deps.AdminDbHandlers.GetMetricsHistory)
			r.Get("/schemas", deps.AdminDbHandlers.GetSchemas)

			r.Route("/{schemaName}", func(r chi.Router) {
				r.Get("/sys", deps.AdminDbHandlers.GetSchemaStats)
				r.Get("/tables", deps.AdminDbHandlers.GetSchemaTables)
			})
		})

		// accounts-base
		r.Route("/accounts", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_2))

			r.Get("/", deps.AdminAccountsHandlers.GetAllAccounts)
			r.Get("/stats", deps.AdminAccountsHandlers.GetUserStats)

			r.Route("/{accountId}", func(r chi.Router) {
				r.Use(deps.AdminAuthService.RequireAdminKeyword)

				r.Get("/", deps.AdminAccountsHandlers.GetAccountByID)
			})
		})
	})

	return r
}

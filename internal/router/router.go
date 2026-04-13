package router

import (
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/di"
	"kroncl-server/internal/metrics"
	"kroncl-server/internal/permissioner"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(cfg *config.Config, container *di.Container) chi.Router {
	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(core.BaseResponse)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// prometheus
	r.Use(metrics.MetricsMiddleware())
	r.With(
		middleware.AllowContentEncoding("identity"),
		metrics.PrometheusIPWhitelist,
	).Get("/metrics", promhttp.Handler().ServeHTTP)

	// API роуты
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", core.HealthCheck)

		// Public routes
		// account actions
		r.Route("/account", func(r chi.Router) {
			// rate limiter
			r.Use(httprate.LimitByIP(config.RATE_LIMIT_PUBLIC_ROUTES_PER_MINUTE, 1*time.Minute))

			r.Post("/reg", container.AccountsHandlers.Register)
			r.Get("/check-email-unique", container.AccountsHandlers.CheckEmailUnique)
			r.Post("/auth", container.AccountsHandlers.Login)
			r.Post("/fingerprints/auth", container.AccountsHandlers.LoginWithFingerprint)
			r.Post("/refresh", container.AccountsHandlers.Refresh)

			// account [protected]
			r.Group(func(r chi.Router) {
				r.Use(container.JWTService.RequireAuth)

				// rate limiter
				r.Use(httprate.LimitByIP(config.RATE_LIMIT_PRIVATE_ROUTES_PER_MINUTE, 1*time.Minute))

				r.Get("/", container.AccountsHandlers.GetProfile)
				r.Patch("/", container.AccountsHandlers.Update)
				r.Post("/confirm", container.AccountsHandlers.ConfirmEmail)
				r.Post("/confirm/resend", container.AccountsHandlers.ResendConfirmationCode)
				r.Post("/log-out", container.AccountsHandlers.Logout)

				// Accounts -> companies invitations [protected]
				r.Route("/invitations", func(r chi.Router) {
					r.Get("/", container.AccountsHandlers.GetAccountInvitations)

					r.Route("/{invitationId}", func(r chi.Router) {
						r.Post("/accept", container.AccountsHandlers.AcceptAccountInvitation)
						r.Post("/reject", container.AccountsHandlers.RejectAccountInvitation)
					})
				})

				// Account -> fingerprints
				r.Route("/fingerprints", func(r chi.Router) {
					r.Get("/", container.AccountsHandlers.GetFingerprints)
					r.Post("/", container.AccountsHandlers.CreateFingerprint)

					r.Route("/{fingerprintId}", func(r chi.Router) {
						r.Post("/revoke", container.AccountsHandlers.RevokeFingerprint)
					})
				})
			})
		})

		// pricing-plans actions
		r.Route("/plans", func(r chi.Router) {
			// rate limiter
			r.Use(httprate.LimitByIP(config.RATE_LIMIT_PUBLIC_ROUTES_PER_MINUTE, 1*time.Minute))

			r.Get("/", container.PricingHandlers.GetPlans)
			r.Get("/{code}", container.PricingHandlers.GetPlanByCode)
		})

		// company permissions + plan lvl mapping (config for all)
		r.Route("/permissions", func(r chi.Router) {
			// rate limiter
			r.Use(httprate.LimitByIP(config.RATE_LIMIT_PUBLIC_ROUTES_PER_MINUTE, 1*time.Minute))

			r.Get("/", container.CompaniesHandlers.GetPlatformPermissions)
		})

		// Protected routes (require auth)
		r.Group(func(r chi.Router) {
			r.Use(container.JWTService.RequireAuth)

			// rate limiter
			r.Use(httprate.LimitByIP(config.RATE_LIMIT_PRIVATE_ROUTES_PER_MINUTE, 1*time.Minute))

			r.Route("/media", func(r chi.Router) {
				r.Post("/upload", container.MediaHandlers.UploadFile)
				r.Get("/{fileId}", container.MediaHandlers.GetFile)
			})

			// Search for public accounts to invite to the company
			r.Route("/accounts", func(r chi.Router) {
				r.Get("/", container.AccountsHandlers.GetPublicAccounts)
			})

			// Companies protected routes
			r.Route("/companies", func(r chi.Router) {
				// Company creation
				r.Post("/", container.CompaniesHandlers.Create)
				r.Get("/my", container.CompaniesHandlers.GetUserCompanies)
				r.Get("/check-slug-unique", container.CompaniesHandlers.CheckSlugUnique)

				// Specific company routes
				r.Route("/{id}", func(r chi.Router) {
					// Company context + basic access check
					r.Use(companies.CompanyMembership(container.DB))
					// Tenant pool middleware
					r.Use(container.StorageService.TenantPoolMiddleware)

					// Company permissions
					r.Get("/permissions", container.CompaniesHandlers.GetCompanyPermissions)

					// Company pricing
					r.Route("/pricing", func(r chi.Router) {
						r.Get("/", container.CompaniesHandlers.GetCompanyPricingPlan) // текущий план+остаток
						r.With(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_PRICING_MIGRATE)).
							Post("/migrate", container.CompaniesHandlers.MigratePricingPlan) // смена плана
						r.With(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_PRICING_TRANSACTIONS)).
							Get("/transactions", container.CompaniesHandlers.GetCompanyPricingTransactions) // получение операций
						r.With(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_PRICING_MIGRATE)).
							Post("/transactions/{transactionId}/revoke", container.CompaniesHandlers.RevokePricingTransaction) // отмена транзакции
					})

					r.Get("/", container.CompaniesHandlers.GetUserCompanyById)
					r.With(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_COMPANY_UPDATE)).Patch("/", container.CompaniesHandlers.Update)

					// Company storage
					r.Route("/storage", func(r chi.Router) {
						r.Get("/", container.StorageHandlers.Get)
						r.With(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_STORAGE_SOURCES)).Get("/sources", container.StorageHandlers.GetSources)
					})

					// Company accounts (hrm part)
					r.Route("/accounts", func(r chi.Router) {
						r.Use(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_ACCOUNTS))
						r.Get("/", container.CompaniesHandlers.GetCompanyMembers)
						r.Get("/{accountId}", container.CompaniesHandlers.GetCompanyMember)

						r.Route("/invitations", func(r chi.Router) {
							r.Use(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_ACCOUNTS_INVITATIONS))

							r.Get("/", container.CompaniesHandlers.GetCompanyInvitations)
							r.With(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_ACCOUNTS_INVITATIONS_CREATE)).
								Post("/", container.CompaniesHandlers.CreateCompanyInvitation)
							r.With(permissioner.RequirePermission(container.PermissionDeps, config.PERMISSION_ACCOUNTS_INVITATIONS_REVOKE)).
								Delete("/{invitationId}", container.CompaniesHandlers.RevokeInvitation)
						})
					})

					// encapsulated modules
					r.Route("/modules", func(r chi.Router) {
						container.TenantRoutes.Register(r, container.PermissionDeps)
					})
				})
			})
		})
	})

	return r
}

package router

import (
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/di"
	"kroncl-server/internal/permissioner"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	// API роуты
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", core.HealthCheck)

		// Public routes
		r.Route("/account", func(r chi.Router) {
			r.Post("/reg", container.AccountsHandlers.Register)
			r.Get("/check-email-unique", container.AccountsHandlers.CheckEmailUnique)
			r.Post("/auth", container.AccountsHandlers.Login)
			r.Post("/refresh", container.AccountsHandlers.Refresh)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(container.JWTService.RequireAuth)
				r.Get("/", container.AccountsHandlers.GetProfile)
				r.Patch("/", container.AccountsHandlers.Update)
				r.Post("/confirm", container.AccountsHandlers.ConfirmEmail)
				r.Post("/confirm/resend", container.AccountsHandlers.ResendConfirmationCode)
			})

			// Accounts -> companies invitations [protected]
			r.Route("/invitations", func(r chi.Router) {
				r.Use(container.JWTService.RequireAuth)

				r.Get("/", container.AccountsHandlers.GetAccountInvitations)

				r.Route("/{invitationId}", func(r chi.Router) {
					r.Post("/accept", container.AccountsHandlers.AcceptAccountInvitation)
					r.Post("/reject", container.AccountsHandlers.RejectAccountInvitation)
				})
			})
		})

		// Protected routes (require auth)
		r.Group(func(r chi.Router) {
			r.Use(container.JWTService.RequireAuth)

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
					// Company context + access check
					r.Use(companies.CompanyMembership(container.DB))
					// Tenant pool middleware
					r.Use(container.StorageService.TenantPoolMiddleware)

					r.Get("/", container.CompaniesHandlers.GetUserCompanyById)
					r.With(permissioner.RequirePermission(container.PermissionService, "company.update")).Patch("/", container.CompaniesHandlers.Update)

					// Company storage
					r.Route("/storage", func(r chi.Router) {
						r.Get("/", container.StorageHandlers.Get)
						r.With(permissioner.RequirePermission(container.PermissionService, "storage.sources")).Get("/sources", container.StorageHandlers.GetSources)
					})

					// Company accounts (hrm part)
					r.Route("/accounts", func(r chi.Router) {
						r.Use(permissioner.RequirePermission(container.PermissionService, "accounts"))
						r.Get("/", container.CompaniesHandlers.GetCompanyMembers)
						r.Get("/{accountId}", container.CompaniesHandlers.GetCompanyMember)

						r.Route("/invitations", func(r chi.Router) {
							r.Use(permissioner.RequirePermission(container.PermissionService, "accounts.invitations"))

							r.Get("/", container.CompaniesHandlers.GetCompanyInvitations)
							r.With(permissioner.RequirePermission(container.PermissionService, "accounts.invitations.create")).Post("/", container.CompaniesHandlers.CreateCompanyInvitation)
							r.With(permissioner.RequirePermission(container.PermissionService, "accounts.invitations.revoke")).Delete("/{invitationId}", container.CompaniesHandlers.RevokeInvitation)
						})
					})

					// encapsulated modules
					r.Route("/modules", func(r chi.Router) {
						container.TenantRoutes.Register(r)
					})
				})
			})
		})
	})

	return r
}

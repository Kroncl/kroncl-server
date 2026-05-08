package admin

import (
	adminaccounts "kroncl-server/internal/admin/accounts"
	adminauth "kroncl-server/internal/admin/auth"
	adminclientele "kroncl-server/internal/admin/clientele"
	admincompanies "kroncl-server/internal/admin/companies"
	admindb "kroncl-server/internal/admin/db"
	adminhealth "kroncl-server/internal/admin/health"
	adminpartners "kroncl-server/internal/admin/partners"
	adminserver "kroncl-server/internal/admin/server"
	adminsupport "kroncl-server/internal/admin/support"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/config"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type Deps struct {
	JWTService             *auth.JWTService
	AdminAuthService       *adminauth.Service
	AdminAuthHandlers      *adminauth.Handlers
	AdminDbHandlers        *admindb.Handlers
	AdminAccountsHandlers  *adminaccounts.Handlers
	AdminClienteleHandlers *adminclientele.Handlers
	AdminCompaniesHandlers *admincompanies.Handlers
	AdminSupportHandlers   *adminsupport.Handlers
	AdminPartnersHandlers  *adminpartners.Handlers
	AdminServerHandlers    *adminserver.Handlers
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

			r.Route("/schemas", func(r chi.Router) {
				r.Get("/", deps.AdminDbHandlers.GetSchemas)

				r.Route("/{schemaName}", func(r chi.Router) {
					r.Get("/sys", deps.AdminDbHandlers.GetSchemaStats)
					r.Get("/tables", deps.AdminDbHandlers.GetSchemaTables)

					// критичные действия [max level + keyword]
					r.Group(func(r chi.Router) {
						r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_MAX))
						r.Use(deps.AdminAuthService.RequireAdminKeyword)

						r.Post("/migrate", deps.AdminDbHandlers.MigrateTenant)
					})
				})

				// критичные действия [max level + keyword]
				r.Group(func(r chi.Router) {
					r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_MAX))
					r.Use(deps.AdminAuthService.RequireAdminKeyword)

					r.Post("/up", deps.AdminDbHandlers.MigrateAllTenants)
				})
			})
		})

		// server-condition
		r.Route("/server", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_1))

			r.Get("/sys", deps.AdminServerHandlers.GetServerStats)
			r.Get("/history", deps.AdminServerHandlers.GetServerHistory)
		})

		// accounts-base
		r.Route("/accounts", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_2))

			r.Get("/", deps.AdminAccountsHandlers.GetAllAccounts)
			r.Get("/stats", deps.AdminAccountsHandlers.GetUserStats)

			r.Route("/{accountId}", func(r chi.Router) {
				r.Get("/", deps.AdminAccountsHandlers.GetAccountByID)

				// критичные действия [max level + keyword]
				r.Group(func(r chi.Router) {
					r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_MAX))
					r.Use(deps.AdminAuthService.RequireAdminKeyword)

					r.Post("/promote-admin", deps.AdminAccountsHandlers.PromoteToAdmin)
					r.Post("/demote-admin", deps.AdminAccountsHandlers.DemoteFromAdmin)
				})
			})
		})

		r.Route("/companies", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_3))

			r.Get("/", deps.AdminCompaniesHandlers.GetAllCompanies)

			r.Route("/{companyId}", func(r chi.Router) {
				r.Get("/", deps.AdminCompaniesHandlers.GetCompanyByID)
				r.Get("/accounts", deps.AdminCompaniesHandlers.GetCompanyAccounts)

				r.Route("/pricing", func(r chi.Router) {
					r.Get("/", deps.AdminCompaniesHandlers.GetCompanyPricingPlan)
				})
			})
		})

		r.Route("/clientele", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_4))

			r.Get("/stats", deps.AdminClienteleHandlers.GetClienteleStats)
			r.Get("/history", deps.AdminClienteleHandlers.GetClienteleHistory)
		})

		// tickets
		r.Route("/tickets", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_4))

			r.Get("/", deps.AdminSupportHandlers.GetAllTickets)

			r.Route("/{ticketId}", func(r chi.Router) {
				r.Get("/", deps.AdminSupportHandlers.GetTicketByID)
				r.Post("/assign", deps.AdminSupportHandlers.AssignTicket)
				r.Post("/unassign", deps.AdminSupportHandlers.UnassignTicket)
				r.Post("/close", deps.AdminSupportHandlers.CloseTicket)

				r.Route("/messages", func(r chi.Router) {
					r.Get("/", deps.AdminSupportHandlers.GetTicketMessages)
					r.Post("/", deps.AdminSupportHandlers.CreateAdminMessage)

					r.Route("/{messageId}", func(r chi.Router) {
						r.Patch("/", deps.AdminSupportHandlers.UpdateAdminMessage)
						r.Delete("/", deps.AdminSupportHandlers.DeleteAdminMessage)
					})
				})
			})
		})

		// partners
		r.Route("/partners", func(r chi.Router) {
			r.Use(deps.AdminAuthService.RequireAdminLevel(config.ADMIN_LEVEL_4))

			r.Get("/", deps.AdminPartnersHandlers.GetAllPartners)

			r.Route("/{partnerId}", func(r chi.Router) {
				r.Get("/", deps.AdminPartnersHandlers.GetPartnerByID)
				r.Patch("/", deps.AdminPartnersHandlers.UpdatePartner)
			})
		})
	})

	return r
}

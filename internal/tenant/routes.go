package tenant

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/permissioner"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Routes struct {
	storageService    *storage.Service
	permissionService *permissioner.Service
	accountsService   *accounts.Service
	companiesService  *companies.Service
}

func NewRoutes(
	storageService *storage.Service,
	permissionService *permissioner.Service,
	accountsService *accounts.Service,
	companiesService *companies.Service,
) *Routes {
	return &Routes{
		storageService:    storageService,
		permissionService: permissionService,
		accountsService:   accountsService,
		companiesService:  companiesService,
	}
}

func (rt *Routes) Register(r chi.Router) {
	// accounts -> employees actions
	r.Route("/accounts", func(r chi.Router) {
		r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_ACCOUNTS_DELETE)).
			Delete("/{accountId}", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
				return h.RemoveEmployeeAccount
			}))
	})

	// HRM module
	r.Route("/hrm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_HRM))

		// employees
		r.Route("/employees", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_HRM_EMPLOYEES))

			r.Get("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
				return h.GetEmployees
			}))
			r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_HRM_EMPLOYEES_CREATE)).
				Post("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
					return h.CreateEmployee
				}))
			r.Route("/{employeeId}", func(r chi.Router) {
				r.Get("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
					return h.GetEmployee
				}))

				// обновление
				r.Group(func(r chi.Router) {
					r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_HRM_EMPLOYEES_UPDATE))

					r.Post("/deactivate", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
						return h.DeactivateEmployee
					}))
					r.Post("/activate", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
						return h.ActivateEmployee
					}))
					r.Patch("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
						return h.UpdateEmployee
					}))
					r.Post("/link-account", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
						return h.LinkAccountEmployee
					}))
					r.Post("/unlink-account", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
						return h.UnlinkAccountEmployee
					}))
				})
			})
		})
	})

	// FM module
	r.Route("/fm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM))

		// transactions
		r.Route("/transactions", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_TRANSACTIONS))

			r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetTransactions
			}))

			r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_TRANSACTIONS_CREATE)).
				Post("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
					return h.CreateTransaction
				}))

			// NO update or delete action
			// for specific transaction
			r.Route("/{transactionId}", func(r chi.Router) {
				r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetTransaction
				}))

				// create reverse transaction
				r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_TRANSACTIONS_REVERSE)).
					Post("/reverse", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.CreateReverseTransaction
					}))
			})

			// transactions categories
			r.Route("/categories", func(r chi.Router) {
				r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES))

				r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetCategories
				}))
				r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE)).
					Post("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.CreateCategory
					}))
				r.Route("/{categoryId}", func(r chi.Router) {
					r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.GetCategory
					}))
					r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE)).
						Patch("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
							return h.UpdateCategory
						}))
					r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE)).
						Delete("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
							return h.DeleteCategory
						}))
				})
			})
		})

		// analysis
		r.Route("/analysis", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_ANALYSIS))

			r.Get("/summary", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetAnalysisSummary
			}))
			r.Get("/grouped", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetGroupedTransactions
			}))
		})
	})
}

// мидлвары для инъекции зависимостей
// тенантских модулей
func (rt *Routes) withHRMHandlers(factory func(*hrm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantPool, ok := rt.storageService.GetTenantPoolFromRequest(r)
		if !ok {
			core.SendError(w, http.StatusInternalServerError, "Error getting a storage connection.")
			return
		}

		// Создаем хэндлеры
		repo := hrm.NewRepository(tenantPool, rt.accountsService, rt.companiesService)
		handlers := hrm.NewHandlers(repo)

		// Вызываем целевой обработчик через фабрику
		handler := factory(handlers)
		handler(w, r)
	}
}

func (rt *Routes) withFMHandlers(factory func(*fm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantPool, ok := rt.storageService.GetTenantPoolFromRequest(r)
		if !ok {
			core.SendError(w, http.StatusInternalServerError, "Error getting a storage connection.")
			return
		}

		hrmRepo := hrm.NewRepository(tenantPool, rt.accountsService, rt.companiesService)
		fmRepo := fm.NewRepository(tenantPool, hrmRepo)
		fmHandlers := fm.NewHandlers(fmRepo)

		// Вызываем целевой обработчик через фабрику
		handler := factory(fmHandlers)
		handler(w, r)
	}
}

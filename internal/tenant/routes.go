package tenant

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/permissioner"
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/logs"
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

	// logs tech actions
	r.Route("/logs", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_LOGS))

		r.Get("/", rt.withLogsHandlers(func(h *logs.Handlers) http.HandlerFunc {
			return h.GetLogs
		}))
		r.Get("/{logId}", rt.withLogsHandlers(func(h *logs.Handlers) http.HandlerFunc {
			return h.GetLog
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

		// counterparties
		r.Route("/counterparties", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_COUNTERPARTIES))

			r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetCounterparties
			}))
			r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_COUNTERPARTIES_CREATE)).
				Post("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
					return h.CreateCounterparty
				}))
			r.Route("/{counterpartyId}", func(r chi.Router) {
				r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetCounterparty
				}))

				// [update counterparty] no hard delete!
				r.Group(func(r chi.Router) {
					r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_COUNTERPARTIES_UPDATE))

					r.Patch("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.UpdateCounterparty
					}))
					r.Post("/deactivate", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.DeactivateCounterparty
					}))
					r.Post("/activate", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.ActivateCounterparty
					}))
				})
			})
		})

		// credits
		r.Route("/credits", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_COUNTERPARTIES))

			r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetCredits
			}))
			r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_CREDITS_CREATE)).
				Post("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
					return h.CreateCredit
				}))
			r.Route("/{creditId}", func(r chi.Router) {
				r.Get("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetCredit
				}))
				r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_CREDITS_TRANSACTIONS)).
					Get("/transactions", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.GetCreditTransactions
					}))
				r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_CREDITS_PAY)).
					Post("/pay", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.PayCredit
					}))

				// [update credit] no hard delete!
				r.Group(func(r chi.Router) {
					r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_FM_COUNTERPARTIES_UPDATE))

					r.Patch("/", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.UpdateCredit
					}))
					r.Post("/deactivate", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.DeactivateCredit
					}))
					r.Post("/activate", rt.withFMHandlers(func(h *fm.Handlers) http.HandlerFunc {
						return h.ActivateCredit
					}))
				})
			})
		})
	})

	// CRM module
	r.Route("/crm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_CRM))

		// sources
		r.Route("/sources", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_CRM_SOURCES))

			r.Get("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
				return h.GetClientSources
			}))
			r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_CRM_SOURCES_CREATE)).
				Post("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
					return h.CreateClientSource
				}))
			r.Route("/{sourceId}", func(r chi.Router) {
				r.Get("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
					return h.GetClientSource
				}))

				// [update source] no hard delete!
				r.Group(func(r chi.Router) {
					r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_CRM_SOURCES_UPDATE))

					r.Patch("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
						return h.UpdateClientSource
					}))
					r.Post("/deactivate", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
						return h.DeactivateClientSource
					}))
					r.Post("/activate", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
						return h.ActivateClientSource
					}))
				})
			})
		})

		// clients
		// r.Route("/clients", func(r chi.Router) {
		// 	r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_CRM_CLIENTS))

		// 	r.Get("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
		// 		return h.GetClients
		// 	}))
		// 	r.With(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_CRM_CLIENTS_CREATE)).
		// 		Post("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
		// 			return h.CreateClient
		// 		}))
		// 	r.Route("/{clientId}", func(r chi.Router) {
		// 		r.Get("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
		// 			return h.GetClient
		// 		}))

		// 		// [update client] no hard delete!
		// 		r.Group(func(r chi.Router) {
		// 			r.Use(permissioner.RequirePermission(rt.permissionService, config.PERMISSION_CRM_CLIENTS_UPDATE))

		// 			r.Patch("/", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
		// 				return h.UpdateClient
		// 			}))
		// 			r.Post("/deactivate", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
		// 				return h.DeactivateClient
		// 			}))
		// 			r.Post("/activate", rt.withCrmHandlers(func(h *crm.Handlers) http.HandlerFunc {
		// 				return h.ActivateClient
		// 			}))
		// 		})
		// 	})
		// })
	})
}

// -------
// INJECTION
// -------

func (rt *Routes) withLogsHandlers(factory func(*logs.Handlers) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantPool, ok := rt.storageService.GetTenantPoolFromRequest(r)
		if !ok {
			core.SendError(w, http.StatusInternalServerError, "Error getting a storage connection.")
			return
		}

		logsService := logs.NewService(tenantPool)
		logsHandlers := logs.NewHandlers(logsService)

		// Вызываем целевой обработчик через фабрику
		handler := factory(logsHandlers)
		handler(w, r)
	}
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
		logsService := logs.NewService(tenantPool)
		handlers := hrm.NewHandlers(repo, logsService)

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
		logsService := logs.NewService(tenantPool)
		fmHandlers := fm.NewHandlers(fmRepo, logsService)

		// Вызываем целевой обработчик через фабрику
		handler := factory(fmHandlers)
		handler(w, r)
	}
}

func (rt *Routes) withCrmHandlers(factory func(*crm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantPool, ok := rt.storageService.GetTenantPoolFromRequest(r)
		if !ok {
			core.SendError(w, http.StatusInternalServerError, "Error getting a storage connection.")
			return
		}

		crmRepo := crm.NewRepository(tenantPool)
		logsService := logs.NewService(tenantPool)
		crmHandlers := crm.NewHandlers(crmRepo, logsService)

		// Вызываем целевой обработчик через фабрику
		handler := factory(crmHandlers)
		handler(w, r)
	}
}

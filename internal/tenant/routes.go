package tenant

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	"kroncl-server/internal/permissioner"
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/dm"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/logs"
	"kroncl-server/internal/tenant/storage"
	"kroncl-server/internal/tenant/support"
	"kroncl-server/internal/tenant/wm"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Routes struct {
	publicPool       *pgxpool.Pool
	storageService   *storage.Service
	accountsService  *accounts.Service
	companiesService *companies.Service
}

func NewRoutes(
	publicPool *pgxpool.Pool,
	storageService *storage.Service,
	accountsService *accounts.Service,
	companiesService *companies.Service,
) *Routes {
	return &Routes{
		publicPool:       publicPool,
		storageService:   storageService,
		accountsService:  accountsService,
		companiesService: companiesService,
	}
}

func (rt *Routes) Register(r chi.Router, permDeps *permissioner.PermissionDeps) {
	// accounts -> employees actions + account settings
	// корявенько получилось в плане /modules/accounts и просто /accounts эп,
	// но пока похуй
	r.Route("/accounts/{accountId}", func(r chi.Router) {
		r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_ACCOUNTS_SETTINGS))

		r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_ACCOUNTS_DELETE)).
			Delete("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
				return h.RemoveEmployeeAccount
			}))

		// цепочка переопределения разрешений
		r.Get("/permissions", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
			return h.GetAccountPermissions
		}))

		// настройки аккаунта в компании
		r.Route("/settings", func(r chi.Router) {
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_ACCOUNTS_SETTINGS))

			r.Get("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
				return h.GetAccountSettings
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_ACCOUNTS_SETTINGS_UPDATE)).
				Patch("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
					return h.UpdateAccountSettings
				}))
		})
	})

	// logs tech actions
	r.Route("/logs", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_LOGS))

		r.Get("/", rt.logs(func(h *logs.Handlers) http.HandlerFunc {
			return h.GetLogs
		}))
		r.Get("/{logId}", rt.logs(func(h *logs.Handlers) http.HandlerFunc {
			return h.GetLog
		}))

		// clear
		// - hard clean all logs
		r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_LOGS_CLEAR)).
			Post("/clear", rt.logs(func(h *logs.Handlers) http.HandlerFunc {
				return h.ClearLogs
			}))

		// optimize
		// - clean old
		r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_LOGS_OPTIMIZE)).
			Post("/optimize", rt.logs(func(h *logs.Handlers) http.HandlerFunc {
				return h.OptimizeLogs
			}))

		// activity
		r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_LOGS_ACTIVITY)).
			Get("/activity", rt.logs(func(h *logs.Handlers) http.HandlerFunc {
				return h.GetLogsActivity
			}))
	})

	// support tech actions
	r.Route("/support/tickets", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_SUPPORT_TICKETS))

		r.Get("/", rt.support(func(h *support.Handlers) http.HandlerFunc {
			return h.GetTickets
		}))
		r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_SUPPORT_TICKETS_CREATE)).
			Post("/", rt.support(func(h *support.Handlers) http.HandlerFunc {
				return h.CreateTicket
			}))
		r.Route("/{ticketId}", func(r chi.Router) {
			r.Get("/", rt.support(func(h *support.Handlers) http.HandlerFunc {
				return h.GetTicketByID
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_SUPPORT_TICKETS_UPDATE)).
				Patch("/", rt.support(func(h *support.Handlers) http.HandlerFunc {
					return h.UpdateTicketStatus
				}))

			r.Route("/messages", func(r chi.Router) {
				r.Post("/", rt.support(func(h *support.Handlers) http.HandlerFunc {
					return h.CreateMessage
				}))
				r.Get("/", rt.support(func(h *support.Handlers) http.HandlerFunc {
					return h.GetMessages
				}))
				r.Patch("/{messageId}", rt.support(func(h *support.Handlers) http.HandlerFunc {
					return h.UpdateMessageReadStatus
				}))

				// [пока впадлу ковыряться с соединениями]
				r.Get("/ws", rt.supportWebsocket(func(h *support.Handlers) http.HandlerFunc {
					return h.MessagesWebSocket
				}))
			})
		})
	})

	// HRM module
	r.Route("/hrm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM))

		// employees
		r.Route("/employees", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_EMPLOYEES))

			r.Get("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
				return h.GetEmployees
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_EMPLOYEES_CREATE)).
				Post("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
					return h.CreateEmployee
				}))
			r.Route("/{employeeId}", func(r chi.Router) {
				r.Get("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
					return h.GetEmployee
				}))

				// обновление
				r.Group(func(r chi.Router) {
					r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_EMPLOYEES_UPDATE))

					r.Post("/deactivate", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.DeactivateEmployee
					}))
					r.Post("/activate", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.ActivateEmployee
					}))
					r.Patch("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.UpdateEmployee
					}))
					r.Post("/link-account", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.LinkAccountEmployee
					}))
					r.Post("/unlink-account", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.UnlinkAccountEmployee
					}))
					r.Post("/link-position", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.LinkPosition
					}))
					r.Post("/unlink-position", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.UnlinkPosition
					}))
				})
			})
		})

		// positions
		r.Route("/positions", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_POSITIONS))

			r.Get("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
				return h.GetPositions
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_POSITIONS_CREATE)).
				Post("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
					return h.CreatePosition
				}))
			r.Route("/{positionId}", func(r chi.Router) {
				r.Get("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
					return h.GetPositionByID
				}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_POSITIONS_UPDATE)).
					Patch("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.UpdatePosition
					}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_POSITIONS_DELETE)).
					Delete("/", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
						return h.DeletePosition
					}))
			})
		})

		// analysis
		r.Route("/analysis", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_HRM_ANALYSIS))

			r.Get("/summary", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
				return h.GetSummary
			}))
			r.Get("/grouped", rt.hrm(func(h *hrm.Handlers) http.HandlerFunc {
				return h.AnalyseGrouped
			}))
		})
	})

	// FM module
	r.Route("/fm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM))

		// transactions
		r.Route("/transactions", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_TRANSACTIONS))

			r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetTransactions
			}))

			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_TRANSACTIONS_CREATE)).
				Post("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
					return h.CreateTransaction
				}))

			// NO update or delete action
			// for specific transaction
			r.Route("/{transactionId}", func(r chi.Router) {
				r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetTransaction
				}))

				// create reverse transaction
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_TRANSACTIONS_REVERSE)).
					Post("/reverse", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.CreateReverseTransaction
					}))
			})

			// transactions categories
			r.Route("/categories", func(r chi.Router) {
				r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES))

				r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetCategories
				}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE)).
					Post("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.CreateCategory
					}))
				r.Route("/{categoryId}", func(r chi.Router) {
					r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.GetCategory
					}))
					r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE)).
						Patch("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
							return h.UpdateCategory
						}))
					r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE)).
						Delete("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
							return h.DeleteCategory
						}))
				})
			})
		})

		// analysis
		r.Route("/analysis", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_ANALYSIS))

			r.Get("/summary", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetAnalysisSummary
			}))
			r.Get("/grouped", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetGroupedTransactions
			}))
		})

		// counterparties
		r.Route("/counterparties", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_COUNTERPARTIES))

			r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetCounterparties
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_COUNTERPARTIES_CREATE)).
				Post("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
					return h.CreateCounterparty
				}))
			r.Route("/{counterpartyId}", func(r chi.Router) {
				r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetCounterparty
				}))

				// [update counterparty] no hard delete!
				r.Group(func(r chi.Router) {
					r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_COUNTERPARTIES_UPDATE))

					r.Patch("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.UpdateCounterparty
					}))
					r.Post("/deactivate", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.DeactivateCounterparty
					}))
					r.Post("/activate", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.ActivateCounterparty
					}))
				})
			})
		})

		// credits
		r.Route("/credits", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_CREDITS))

			r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
				return h.GetCredits
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_CREDITS_CREATE)).
				Post("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
					return h.CreateCredit
				}))
			r.Route("/{creditId}", func(r chi.Router) {
				r.Get("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
					return h.GetCredit
				}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_CREDITS_TRANSACTIONS)).
					Get("/transactions", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.GetCreditTransactions
					}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_CREDITS_PAY)).
					Post("/pay", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.PayCredit
					}))

				// [update credit] no hard delete!
				r.Group(func(r chi.Router) {
					r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_FM_COUNTERPARTIES_UPDATE))

					r.Patch("/", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.UpdateCredit
					}))
					r.Post("/deactivate", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.DeactivateCredit
					}))
					r.Post("/activate", rt.fm(func(h *fm.Handlers) http.HandlerFunc {
						return h.ActivateCredit
					}))
				})
			})
		})
	})

	// CRM module
	r.Route("/crm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM))

		// sources
		r.Route("/sources", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM_SOURCES))

			r.Get("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
				return h.GetClientSources
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM_SOURCES_CREATE)).
				Post("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
					return h.CreateClientSource
				}))
			r.Route("/{sourceId}", func(r chi.Router) {
				r.Get("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
					return h.GetClientSource
				}))

				// [update source] no hard delete!
				r.Group(func(r chi.Router) {
					r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM_SOURCES_UPDATE))

					r.Patch("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
						return h.UpdateClientSource
					}))
					r.Post("/deactivate", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
						return h.DeactivateClientSource
					}))
					r.Post("/activate", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
						return h.ActivateClientSource
					}))
				})
			})
		})

		// clients
		r.Route("/clients", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM_CLIENTS))

			r.Get("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
				return h.GetClients
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM_CLIENTS_CREATE)).
				Post("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
					return h.CreateClient
				}))
			r.Route("/{clientId}", func(r chi.Router) {
				r.Get("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
					return h.GetClient
				}))

				// [update client] no hard delete!
				r.Group(func(r chi.Router) {
					r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM_CLIENTS_UPDATE))

					r.Patch("/", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
						return h.UpdateClient
					}))
					r.Post("/deactivate", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
						return h.DeactivateClient
					}))
					r.Post("/activate", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
						return h.ActivateClient
					}))
				})
			})
		})

		// analysis
		r.Route("/analysis", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_CRM_ANALYSIS))

			r.Get("/summary", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
				return h.GetClientsSummary
			}))
			r.Get("/grouped", rt.crm(func(h *crm.Handlers) http.HandlerFunc {
				return h.GetGroupedClients
			}))
		})
	})

	// WM module
	r.Route("/wm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM))

		// catalog
		r.Route("/catalog", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_CATALOG))

			// categories
			r.Route("/categories", func(r chi.Router) {
				r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_CATALOG_CATEGORIES))

				r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
					return h.GetCatalogCategories
				}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_CATALOG_CATEGORIES_CREATE)).
					Post("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
						return h.CreateCatalogCategory
					}))
				r.Route("/{categoryId}", func(r chi.Router) {
					r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
						return h.GetCatalogCategory
					}))

					// [update category] no hard delete!
					r.Group(func(r chi.Router) {
						r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE))

						r.Patch("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
							return h.UpdateCatalogCategory
						}))
						r.Post("/deactivate", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
							return h.DeactivateCatalogCategory
						}))
						r.Post("/activate", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
							return h.ActivateCatalogCategory
						}))
					})
				})
			})

			// units
			r.Route("/units", func(r chi.Router) {
				r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_CATALOG_UNITS))

				r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
					return h.GetCatalogUnits
				}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_CATALOG_UNITS_CREATE)).
					Post("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
						return h.CreateCatalogUnit
					}))
				r.Route("/{unitId}", func(r chi.Router) {
					r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
						return h.GetCatalogUnit
					}))

					// [update unit] no hard delete!
					r.Group(func(r chi.Router) {
						r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_CATALOG_UNITS_UPDATE))

						r.Patch("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
							return h.UpdateCatalogUnit
						}))
						r.Post("/deactivate", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
							return h.DeactivateCatalogUnit
						}))
						r.Post("/activate", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
							return h.ActivateCatalogUnit
						}))
					})
				})
			})
		})

		// stocks
		r.Route("/stocks", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_STOCKS))

			// batches
			r.Route("/batches", func(r chi.Router) {
				r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_STOCKS_BATCHES))

				r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
					return h.GetStockBatches
				}))
				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_STOCKS_BATCHES_CREATE)).
					Post("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
						return h.CreateStockBatch
					}))
				r.Route("/{batchId}", func(r chi.Router) {
					r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
						return h.GetStockBatch
					}))
				})
			})

			// positions
			r.Route("/positions", func(r chi.Router) {
				r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_WM_STOCKS_POSITIONS))

				r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
					return h.GetStockPositions
				}))
				r.Route("/{positionId}", func(r chi.Router) {
					r.Get("/", rt.wm(func(h *wm.Handlers) http.HandlerFunc {
						return h.GetStockPosition
					}))
				})
			})
		})
	})

	// DM module
	r.Route("/dm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_DM))

		// types
		r.Route("/types", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_TYPES))

			r.Get("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
				return h.GetDealTypes
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_TYPES_CREATE)).
				Post("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
					return h.CreateDealType
				}))

			r.Route("/{typeId}", func(r chi.Router) {
				r.Get("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
					return h.GetDealType
				}))

				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_TYPES_UPDATE)).
					Patch("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
						return h.UpdateDealType
					}))

				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_TYPES_DELETE)).
					Delete("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
						return h.DeleteDealType
					}))
			})
		})

		// statuses
		r.Route("/statuses", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_STATUSES))

			r.Get("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
				return h.GetDealStatuses
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_STATUSES_CREATE)).
				Post("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
					return h.CreateDealStatus
				}))

			// reorder collection
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_STATUSES_UPDATE)).
				Put("/reorder", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
					return h.ReorderDealStatuses
				}))

			r.Route("/{statusId}", func(r chi.Router) {
				r.Get("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
					return h.GetDealStatus
				}))

				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_STATUSES_UPDATE)).
					Patch("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
						return h.UpdateDealStatus
					}))

				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_STATUSES_DELETE)).
					Delete("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
						return h.DeleteDealStatus
					}))
			})
		})

		// deals
		r.Route("/deals", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_DEALS))

			r.Get("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
				return h.GetDeals
			}))
			r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_DEALS_CREATE)).
				Post("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
					return h.CreateDeal
				}))

			r.Route("/{dealId}", func(r chi.Router) {
				r.Get("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
					return h.GetDeal
				}))

				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_DEALS_UPDATE)).
					Patch("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
						return h.UpdateDeal
					}))

				r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_DEALS_DELETE)).
					Delete("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
						return h.DeleteDeal
					}))

				r.Route("/transactions", func(r chi.Router) {
					r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_DEALS_TRANSACTIONS))

					r.Get("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
						return h.GetDealTransactions
					}))

					r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE)).
						Post("/", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
							return h.CreateDealTransaction
						}))

					r.With(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_DEALS_TRANSACTIONS_SUMMARY)).
						Get("/summary", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
							return h.GetDealTransactionsSummary
						}))
				})
			})
		})

		// analysis
		r.Route("/analysis", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(permDeps, config.PERMISSION_DM_ANALYSIS))

			r.Get("/summary", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
				return h.GetAnalysisSummary
			}))
			r.Get("/grouped", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
				return h.GetAnalysisGrouped
			}))
			r.Get("/financial-summary", rt.dm(func(h *dm.Handlers) http.HandlerFunc {
				return h.GetAnalysisFinancialSummary
			}))
		})
	})
}

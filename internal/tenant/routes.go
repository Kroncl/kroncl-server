package tenant

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/core"
	"kroncl-server/internal/permissioner"
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
	compoaniesService *companies.Service,
) *Routes {
	return &Routes{
		storageService:    storageService,
		permissionService: permissionService,
		accountsService:   accountsService,
		companiesService:  compoaniesService,
	}
}

func (rt *Routes) Register(r chi.Router) {
	// Accounts - employees actions
	r.Route("/accounts", func(r chi.Router) {
		r.With(permissioner.RequirePermission(rt.permissionService, "accounts.delete")).Delete("/{accountId}", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
			return h.RemoveEmployeeAccount
		}))
	})

	// HRM module
	r.Route("/hrm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(rt.permissionService, "hrm"))

		// Employees
		r.Route("/employees", func(r chi.Router) {
			r.Use(permissioner.RequirePermission(rt.permissionService, "hrm.employees"))

			r.Get("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
				return h.GetEmployees
			}))
			r.With(permissioner.RequirePermission(rt.permissionService, "hrm.employees.create")).Post("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
				return h.CreateEmployee
			}))
			r.Route("/{employeeId}", func(r chi.Router) {
				r.Get("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
					return h.GetEmployee
				}))
				r.With(permissioner.RequirePermission(rt.permissionService, "hrm.employees.update")).Patch("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
					return h.UpdateEmployee
				}))
				r.With(permissioner.RequirePermission(rt.permissionService, "hrm.employees.delete")).Delete("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
					return h.DeleteEmployee
				}))
			})
		})
	})
}

// withHRMHandlers создает middleware, которое внедряет HRM хэндлеры
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

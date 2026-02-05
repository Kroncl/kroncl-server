package tenant

import (
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
}

func NewRoutes(storageService *storage.Service, permissionService *permissioner.Service) *Routes {
	return &Routes{
		storageService:    storageService,
		permissionService: permissionService,
	}
}

func (rt *Routes) Register(r chi.Router) {
	// HRM module
	r.Route("/hrm", func(r chi.Router) {
		r.Use(permissioner.RequirePermission(rt.permissionService, "hrm.view"))

		// Employees
		r.Route("/employees", func(r chi.Router) {
			r.Get("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
				return h.GetEmployees
			}))

			r.Route("/{employeeId}", func(r chi.Router) {
				r.Get("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
					return h.GetEmployee
				}))
				r.Patch("/", rt.withHRMHandlers(func(h *hrm.Handlers) http.HandlerFunc {
					return h.UpdateEmployee
				}))
			})
		})
	})

	// Можно добавить другие модули аналогично:
	// rt.registerAccountingRoutes(r)
	// rt.registerCRMHandlers(r)
}

// withHRMHandlers создает middleware, которое внедряет HRM хэндлеры
func (rt *Routes) withHRMHandlers(factory func(*hrm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := core.GetCompanyIDFromContext(r.Context())
		if !ok {
			core.SendError(w, http.StatusInternalServerError, "Company ID not found")
			return
		}

		tenantPool, err := rt.storageService.GetTenantPool(r.Context(), companyID)
		if err != nil {
			core.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer tenantPool.Close()

		// Создаем хэндлеры
		repo := hrm.NewRepository(tenantPool)
		handlers := hrm.NewHandlers(repo)

		// Вызываем целевой обработчик через фабрику
		handler := factory(handlers)
		handler(w, r)
	}
}

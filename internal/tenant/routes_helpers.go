package tenant

import (
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/dm"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/logs"
	"kroncl-server/internal/tenant/wm"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ебашим мидлвар на создание хэндлеров модулей
// мидлвар модуля -> этот метод -> фабрика модуля -> готовые хэндлеры
// ---------------->[достаём пул]->[передаём модулям]----------------
func withModuleMiddleware[H any](
	rt *Routes,
	factory func(*pgxpool.Pool, *logs.Service, *Routes) H,
	handlerFunc func(H) http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantPool, ok := rt.storageService.GetTenantPoolFromRequest(r)
		if !ok {
			core.SendError(w, http.StatusInternalServerError, "Error getting a storage connection.")
			return
		}

		logsService := logs.NewService(tenantPool)
		handler := factory(tenantPool, logsService, rt)

		handlerFunc(handler)(w, r)
	}
}

func (rt *Routes) hrm(h func(*hrm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withModuleMiddleware(rt, createHRMHandlers, h)
}

func (rt *Routes) fm(h func(*fm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withModuleMiddleware(rt, createFMHandlers, h)
}

func (rt *Routes) crm(h func(*crm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withModuleMiddleware(rt, createCRMHandlers, h)
}

func (rt *Routes) wm(h func(*wm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withModuleMiddleware(rt, createWMHandlers, h)
}

func (rt *Routes) logs(h func(*logs.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withModuleMiddleware(rt, createLogsHandlers, h)
}

func (rt *Routes) dm(h func(*dm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withModuleMiddleware(rt, createDMHandlers, h)
}

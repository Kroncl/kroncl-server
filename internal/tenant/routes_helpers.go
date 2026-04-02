package tenant

import (
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/dm"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/logs"
	"kroncl-server/internal/tenant/support"
	"kroncl-server/internal/tenant/wm"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func withPublicPoolMiddleware[H any](
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

		handler := factory(rt.publicPool, logsService, rt)
		handlerFunc(handler)(w, r)
	}
}

// ебашим мидлвар на создание хэндлеров модулей - глобальный пул не нужен
// мидлвар модуля -> этот метод -> фабрика модуля -> готовые хэндлеры
// ---------------->[достаём пул]->[передаём модулям]----------------
func withTenantPoolMiddleware[H any](
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
	return withPublicPoolMiddleware(rt, createHRMHandlers, h)
}

func (rt *Routes) fm(h func(*fm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withPublicPoolMiddleware(rt, createFMHandlers, h)
}

func (rt *Routes) crm(h func(*crm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withPublicPoolMiddleware(rt, createCRMHandlers, h)
}

func (rt *Routes) wm(h func(*wm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withPublicPoolMiddleware(rt, createWMHandlers, h)
}

func (rt *Routes) logs(h func(*logs.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withPublicPoolMiddleware(rt, createLogsHandlers, h)
}

func (rt *Routes) dm(h func(*dm.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withPublicPoolMiddleware(rt, createDMHandlers, h)
}

func (rt *Routes) support(h func(*support.Handlers) http.HandlerFunc) http.HandlerFunc {
	return withPublicPoolMiddleware(rt, createSupportHandlers, h)
}

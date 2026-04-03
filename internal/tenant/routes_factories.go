package tenant

import (
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/dm"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/logs"
	"kroncl-server/internal/tenant/support"
	"kroncl-server/internal/tenant/wm"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Support tickets factory
func createSupportHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *support.Handlers {
	supportService := support.NewService(pool, rt.accountsService, rt.companiesService)
	return support.NewHandlers(supportService, logsService)
}

// HRM factory
func createHRMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *hrm.Handlers {
	hrmRepo := hrm.NewRepository(pool, rt.accountsService, rt.companiesService)
	return hrm.NewHandlers(hrmRepo, logsService)
}

// FM factory
func createFMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *fm.Handlers {
	hrmRepo := hrm.NewRepository(pool, rt.accountsService, rt.companiesService)
	fmRepo := fm.NewRepository(pool, hrmRepo)
	return fm.NewHandlers(fmRepo, logsService)
}

// CRM factory
func createCRMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *crm.Handlers {
	crmRepo := crm.NewRepository(pool)
	return crm.NewHandlers(crmRepo, logsService)
}

// WM factory
func createWMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *wm.Handlers {
	wmRepo := wm.NewRepository(pool)
	return wm.NewHandlers(wmRepo, logsService)
}

// Logs factory
func createLogsHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *logs.Handlers {
	return logs.NewHandlers(logsService)
}

// DM factory
func createDMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *dm.Handlers {
	hrmRepo := hrm.NewRepository(pool, rt.accountsService, rt.companiesService)
	fmRepo := fm.NewRepository(pool, hrmRepo)
	crmRepo := crm.NewRepository(pool)
	wmRepo := wm.NewRepository(pool)
	dmRepo := dm.NewRepository(pool, fmRepo, hrmRepo, crmRepo, wmRepo)
	return dm.NewHandlers(dmRepo, logsService)
}

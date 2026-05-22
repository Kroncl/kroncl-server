package tenant

import (
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/dm"
	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/excelizer"
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
	excelizerService := excelizer.NewService(rt.storageService.Media)
	docsService := docs.NewService(pool)

	hrmRepo := hrm.NewRepository(
		pool,
		rt.accountsService,
		rt.companiesService,
		rt.storageService.Media,
		excelizerService,
		docsService,
	)
	return hrm.NewHandlers(hrmRepo, logsService)
}

// FM factory
func createFMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *fm.Handlers {
	excelizerService := excelizer.NewService(rt.storageService.Media)
	docsService := docs.NewService(pool)

	hrmRepo := hrm.NewRepository(
		pool,
		rt.accountsService,
		rt.companiesService,
		rt.storageService.Media,
		excelizerService,
		docsService,
	)
	fmRepo := fm.NewRepository(pool, hrmRepo, rt.storageService.Media, excelizerService, docsService)
	return fm.NewHandlers(fmRepo, logsService)
}

// CRM factory
func createCRMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *crm.Handlers {
	excelizerService := excelizer.NewService(rt.storageService.Media)
	docsService := docs.NewService(pool)

	crmRepo := crm.NewRepository(pool, rt.storageService.Media, excelizerService, docsService)
	return crm.NewHandlers(crmRepo, logsService)
}

// WM factory
func createWMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *wm.Handlers {
	excelizerService := excelizer.NewService(rt.storageService.Media)
	docsService := docs.NewService(pool)
	wmRepo := wm.NewRepository(pool, rt.storageService.Media, excelizerService, docsService)
	return wm.NewHandlers(wmRepo, logsService)
}

// Logs factory
func createLogsHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *logs.Handlers {
	return logs.NewHandlers(logsService)
}

// Docs
func createDocsHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *docs.Handlers {
	docsService := docs.NewService(pool)
	return docs.NewHandlers(docsService, logsService)
}

// DM factory
func createDMHandlers(pool *pgxpool.Pool, logsService *logs.Service, rt *Routes) *dm.Handlers {
	excelizerService := excelizer.NewService(rt.storageService.Media)
	docsService := docs.NewService(pool)

	hrmRepo := hrm.NewRepository(
		pool,
		rt.accountsService,
		rt.companiesService,
		rt.storageService.Media,
		excelizerService,
		docsService,
	)
	fmRepo := fm.NewRepository(pool, hrmRepo, rt.storageService.Media, excelizerService, docsService)
	crmRepo := crm.NewRepository(pool, rt.storageService.Media, excelizerService, docsService)
	wmRepo := wm.NewRepository(pool, rt.storageService.Media, excelizerService, docsService)
	dmRepo := dm.NewRepository(pool, fmRepo, hrmRepo, crmRepo, wmRepo)
	return dm.NewHandlers(dmRepo, logsService)
}

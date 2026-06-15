package di

import (
	"context"
	"fmt"
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/admin"
	adminaccounts "kroncl-server/internal/admin/accounts"
	adminauth "kroncl-server/internal/admin/auth"
	adminclientele "kroncl-server/internal/admin/clientele"
	admincompanies "kroncl-server/internal/admin/companies"
	admindb "kroncl-server/internal/admin/db"
	adminmedia "kroncl-server/internal/admin/media"
	adminpartners "kroncl-server/internal/admin/partners"
	adminpricing "kroncl-server/internal/admin/pricing"
	adminserver "kroncl-server/internal/admin/server"
	adminsupport "kroncl-server/internal/admin/support"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/billing"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	corestatus "kroncl-server/internal/core/status"
	coreworkers "kroncl-server/internal/core/workers"
	"kroncl-server/internal/mailer"
	"kroncl-server/internal/media"
	"kroncl-server/internal/migrator"
	"kroncl-server/internal/permissioner"
	"kroncl-server/internal/pricing"
	"kroncl-server/internal/public"
	"kroncl-server/internal/tenant"
	"kroncl-server/internal/tenant/pdfgen"
	"kroncl-server/internal/tenant/storage"
	storagedb "kroncl-server/internal/tenant/storage/db"
	storagemedia "kroncl-server/internal/tenant/storage/media"
	"kroncl-server/utils"
	"log"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	// system-core
	Config            *config.Config
	DB                *pgxpool.Pool
	JWTService        *auth.JWTService
	PermissionService *permissioner.Service
	Migrator          *migrator.Migrator
	Mailer            *mailer.Service
	MediaRepo         *media.Repository
	MediaService      *media.Service
	MediaHandlers     *media.Handlers
	Pdfgen            *pdfgen.Service // pdfgen (gotenberg)

	// business-core
	AccountsService   *accounts.Service
	CompaniesService  *companies.Service
	PricingService    *pricing.Service
	PublicService     *public.Service
	BillingService    *billing.Service
	AccountsHandlers  *accounts.Handlers
	CompaniesHandlers *companies.Handlers
	PricingHandlers   *pricing.Handlers
	PublicHandlers    *public.Handlers
	BillingHandlers   *billing.Handlers

	// tenant storage ctrl [db + media]
	StorageService       *storage.Service
	StorageHandlers      *storage.Handlers
	StorageDbService     *storagedb.Service
	StorageDbHandlers    *storagedb.Handlers
	StorageMediaService  *storagemedia.Service
	StorageMediaHandlers *storagemedia.Handlers

	// мидлварь зависимости
	PermissionDeps *permissioner.PermissionDeps

	// Tenant модули
	TenantRoutes *tenant.Routes

	// admin
	AdminAuthService       *adminauth.Service
	AdminAuthHandlers      *adminauth.Handlers
	AdminDbService         *admindb.Service
	AdminDbHandlers        *admindb.Handlers
	AdminAccountsService   *adminaccounts.Service
	AdminAccountsHandlers  *adminaccounts.Handlers
	AdminClienteleService  *adminclientele.Service
	AdminClienteleHandlers *adminclientele.Handlers
	AdminCompaniesService  *admincompanies.Service
	AdminCompaniesHandlers *admincompanies.Handlers
	AdminSupportService    *adminsupport.Service
	AdminSupportHandlers   *adminsupport.Handlers
	AdminPartnersService   *adminpartners.Service
	AdminPartnersHandlers  *adminpartners.Handlers
	AdminServerService     *adminserver.Service
	AdminServerHandlers    *adminserver.Handlers
	AdminPricingService    *adminpricing.Service
	AdminPricingHandlers   *adminpricing.Handlers
	AdminMediaService      *adminmedia.Service
	AdminMediaHandlers     *adminmedia.Handlers
	AdminRoutes            chi.Router

	// workers
	CoreWorkersService         *coreworkers.Service
	CoreDbMetricsWorker        *coreworkers.Worker
	CoreClienteleMetricsWorker *coreworkers.Worker
	CoreServerMetricsWorker    *coreworkers.Worker
	CoreMediaMetricsWorker     *coreworkers.Worker

	// core-status
	CoreStatusService  *corestatus.Service
	CoreStatusHandlers *corestatus.Handlers
}

func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	if err := c.initDB(ctx); err != nil {
		return nil, err
	}
	if err := c.initMigrator(); err != nil {
		return nil, err
	}
	if err := c.initServices(ctx); err != nil {
		return nil, err
	}
	if err := c.initTenantRoutes(); err != nil {
		return nil, err
	}
	if err := c.initAdminRoutes(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Container) initDB(ctx context.Context) error {
	dsn, err := utils.BuildDSN(c.Config.Database)
	if err != nil {
		return err
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	c.DB = pool
	log.Println("✅ Подключение к БД установлено")
	return nil
}

func (c *Container) initMigrator() error {
	// Получаем абсолютный путь к миграциям
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	migrationsPath := filepath.Join(cwd, "migrations")

	// Создаем конфиг мигратора
	migratorCfg := migrator.Config{
		MigrationsPath: migrationsPath,
		SchemaType:     migrator.SchemaTypeTenant,
	}

	// Создаем мигратор
	mig, err := migrator.NewMigrator(c.DB, migratorCfg)
	if err != nil {
		return err
	}

	c.Migrator = mig
	log.Printf("📁 Мигратор инициализирован: %s", migrationsPath)
	return nil
}

func (c *Container) initServices(ctx context.Context) error {
	// JWT
	c.JWTService = auth.NewJWTService(
		c.Config.JWT.SecretKey,
		c.Config.JWT.AccessDuration,
		c.Config.JWT.RefreshDuration,
		c.Config.JWT.ResetPasswordSecretKey,
		c.Config.JWT.ResetPasswordDuration,
	)

	// admin-auth [используется в APP->accounts]
	c.AdminAuthService = adminauth.NewService(c.DB)

	// -----------
	// TENANT STORAGE CTRL
	// -----------

	dbStorageRepo := storagedb.NewRepository(c.DB)
	c.StorageDbService = storagedb.NewService(dbStorageRepo, c.Migrator, c.DB)
	storageMediaService, err := storagemedia.NewService(c.Config.MinIO)
	if err != nil {
		return fmt.Errorf("failed to init storage media service: %w", err)
	}
	c.StorageMediaService = storageMediaService
	c.StorageMediaHandlers = storagemedia.NewHandlers(c.StorageMediaService)

	// -->abstract storage service for tenant
	c.StorageService = storage.NewService(
		c.StorageDbService,
		c.StorageMediaService,
	)
	c.StorageHandlers = storage.NewHandlers(c.StorageService)

	// -----------
	// pdfgen [gotenberg]
	// -----------
	pdfGenService, err := pdfgen.NewService(pdfgen.Config{
		Endpoint: c.Config.Gotenberg.Endpoint,
	})
	if err != nil {
		return fmt.Errorf("failed to init pdfgen service: %w", err)
	}
	c.Pdfgen = pdfGenService

	// ------------
	// APP
	// ------------

	// Pricing+billing Service
	c.PricingService = pricing.NewService(c.DB)
	c.BillingService = billing.NewService(c.DB)

	// Mailer Service
	c.Mailer = mailer.NewService(&c.Config.MailSender)

	// Companies Service (зависит от Storage)
	c.CompaniesService = companies.NewService(c.DB, c.StorageService, c.PricingService, c.Mailer)

	// Accounts Service (зависит от JWT и Companies)
	c.AccountsService = accounts.NewService(c.DB, c.JWTService, c.CompaniesService, c.Mailer, c.AdminAuthService)

	// -------> accounts -> jwtService
	c.JWTService.SetApiKeyValidator(c.AccountsService)

	// Permission Service
	c.PermissionService = permissioner.NewService(c.CompaniesService)

	// Public Service
	c.PublicService = public.NewService(c.DB, c.Mailer)

	// media
	mediaRepo := media.NewRepository(c.DB)
	mediaService, err := media.NewService(media.Config{
		Endpoint:   c.Config.MinIO.Endpoint,
		AccessKey:  c.Config.MinIO.RootUser,
		SecretKey:  c.Config.MinIO.RootPassword,
		UseSSL:     c.Config.MinIO.UseSSL,
		Bucket:     c.Config.MinIO.PublicBucket,
		PublicHost: c.Config.MinIO.ExternalHost,
	}, mediaRepo)

	if err != nil {
		return fmt.Errorf("failed to init media service: %w", err)
	}

	mediaHandlers := media.NewHandlers(mediaService)

	// Permission Service
	c.PermissionService = permissioner.NewService(c.CompaniesService)
	c.PermissionDeps = &permissioner.PermissionDeps{
		PermService:      c.PermissionService,
		StorageDbService: c.StorageDbService,
	}

	// Сохраняем в контейнер
	c.MediaRepo = mediaRepo
	c.MediaService = mediaService
	c.MediaHandlers = mediaHandlers

	// Handlers
	c.AccountsHandlers = accounts.NewHandlers(c.AccountsService)
	c.CompaniesHandlers = companies.NewHandlers(c.CompaniesService)
	c.StorageDbHandlers = storagedb.NewHandlers(c.StorageDbService)
	c.PricingHandlers = pricing.NewHandlers(c.PricingService)
	c.PublicHandlers = public.NewHandlers(c.PublicService)
	c.BillingHandlers = billing.NewHandlers(c.BillingService)

	// ---------
	// WORKERS
	// ---------
	c.CoreWorkersService = coreworkers.NewService(c.DB, c.PricingService, c.CompaniesService, c.AccountsService, c.StorageMediaService)
	c.CoreDbMetricsWorker = coreworkers.NewDbWorker(c.CoreWorkersService, config.WORKER_METRICS_DB_PERIOD_CRON)
	c.CoreClienteleMetricsWorker = coreworkers.NewClienteleWorker(c.CoreWorkersService, config.WORKER_METRICS_CLIENTELE_PERIOD_CRON)
	c.CoreServerMetricsWorker = coreworkers.NewServerWorker(c.CoreWorkersService, config.WORKER_METRICS_SERVER_PERIOD_CRON)
	c.CoreMediaMetricsWorker = coreworkers.NewMediaWorker(c.CoreWorkersService, config.WORKER_METRICS_MEDIA_PERIOD_CRON)

	// ----------
	// ADMIN
	// ----------
	c.AdminAuthHandlers = adminauth.NewHandlers(c.AdminAuthService)
	c.AdminDbService = admindb.NewService(c.DB, c.CoreWorkersService, c.Migrator)
	c.AdminDbHandlers = admindb.NewHandlers(c.AdminDbService)
	c.AdminAccountsService = adminaccounts.NewService(c.DB, c.AccountsService, c.AdminAuthService)
	c.AdminAccountsHandlers = adminaccounts.NewHandlers(c.AdminAccountsService)
	c.AdminClienteleService = adminclientele.NewService(c.DB, c.CoreWorkersService)
	c.AdminClienteleHandlers = adminclientele.NewHandlers(c.AdminClienteleService)
	c.AdminCompaniesService = admincompanies.NewService(c.DB, c.CompaniesService, c.StorageService)
	c.AdminCompaniesHandlers = admincompanies.NewHandlers(c.AdminCompaniesService)
	c.AdminSupportService = adminsupport.NewService(c.DB, c.CompaniesService, c.AccountsService, c.Mailer)
	c.AdminSupportHandlers = adminsupport.NewHandlers(c.AdminSupportService)
	c.AdminPartnersService = adminpartners.NewService(c.DB, c.PublicService)
	c.AdminPartnersHandlers = adminpartners.NewHandlers(c.AdminPartnersService)
	c.AdminServerService = adminserver.NewService(c.DB, c.CoreWorkersService)
	c.AdminServerHandlers = adminserver.NewHandlers(c.AdminServerService)
	c.AdminPricingService = adminpricing.NewService(c.DB, c.PricingService)
	c.AdminPricingHandlers = adminpricing.NewHandlers(c.AdminPricingService)
	c.AdminMediaService = adminmedia.NewService(c.DB, c.CoreWorkersService, c.StorageMediaService)
	c.AdminMediaHandlers = adminmedia.NewHandlers(c.AdminMediaService)

	// -----------
	// CORE-STATUS
	// -----------

	c.CoreStatusService = corestatus.NewService(c.DB, c.CoreWorkersService)
	c.CoreStatusHandlers = corestatus.NewHandlers(c.CoreStatusService)

	return nil
}

func (c *Container) initTenantRoutes() error {
	c.TenantRoutes = tenant.NewRoutes(
		c.DB,
		c.StorageService,
		c.AccountsService,
		c.CompaniesService,
		c.Pdfgen,
		c.Mailer,
	)
	return nil
}

func (c *Container) initAdminRoutes() error {
	c.AdminRoutes = admin.NewRoutes(admin.Deps{
		JWTService:             c.JWTService,
		AdminAuthService:       c.AdminAuthService,
		AdminDbHandlers:        c.AdminDbHandlers,
		AdminAccountsHandlers:  c.AdminAccountsHandlers,
		AdminAuthHandlers:      c.AdminAuthHandlers,
		AdminClienteleHandlers: c.AdminClienteleHandlers,
		AdminCompaniesHandlers: c.AdminCompaniesHandlers,
		AdminSupportHandlers:   c.AdminSupportHandlers,
		AdminPartnersHandlers:  c.AdminPartnersHandlers,
		AdminServerHandlers:    c.AdminServerHandlers,
		AdminPricingHandlers:   c.AdminPricingHandlers,
		AdminMediaHandlers:     c.AdminMediaHandlers,
	})
	return nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.StorageService != nil {
		c.StorageDbService.CloseAll()
	}
}

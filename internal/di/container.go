package di

import (
	"context"
	"fmt"
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/admin"
	adminaccounts "kroncl-server/internal/admin/accounts"
	adminauth "kroncl-server/internal/admin/auth"
	admindb "kroncl-server/internal/admin/db"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	coreworkers "kroncl-server/internal/core/workers"
	"kroncl-server/internal/mailer"
	"kroncl-server/internal/media"
	"kroncl-server/internal/migrator"
	"kroncl-server/internal/permissioner"
	"kroncl-server/internal/pricing"
	"kroncl-server/internal/public"
	"kroncl-server/internal/tenant"
	"kroncl-server/internal/tenant/storage"
	"kroncl-server/utils"
	"log"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Container собирает все зависимости
type Container struct {
	Config *config.Config
	DB     *pgxpool.Pool

	// Сервисы
	JWTService        *auth.JWTService
	AccountsService   *accounts.Service
	CompaniesService  *companies.Service
	PricingService    *pricing.Service
	PermissionService *permissioner.Service
	StorageService    *storage.Service
	Migrator          *migrator.Migrator
	Mailer            *mailer.Service
	PublicService     *public.Service

	// Media
	MediaRepo     *media.Repository
	MediaService  *media.Service
	MediaHandlers *media.Handlers

	// Хэндлеры
	AccountsHandlers  *accounts.Handlers
	CompaniesHandlers *companies.Handlers
	StorageHandlers   *storage.Handlers
	PricingHandlers   *pricing.Handlers
	PublicHandlers    *public.Handlers

	// мидлварь зависимости
	PermissionDeps *permissioner.PermissionDeps

	// Tenant модули
	TenantRoutes *tenant.Routes

	// admin
	AdminAuthService      *adminauth.Service
	AdminAuthHandlers     *adminauth.Handlers
	AdminDbService        *admindb.Service
	AdminDbHandlers       *admindb.Handlers
	AdminAccountsService  *adminaccounts.Service
	AdminAccountsHandlers *adminaccounts.Handlers
	AdminRoutes           chi.Router

	// workers
	CoreWorkersService *coreworkers.Service
	CoreWorkers        *coreworkers.Worker
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

	// workers
	c.CoreWorkersService = coreworkers.NewService(c.DB)
	c.CoreWorkers = coreworkers.NewWorker(c.CoreWorkersService, "@every 60s")

	// admin-auth [используется в APP->accounts]
	c.AdminAuthService = adminauth.NewService(c.DB)

	// ------------
	// APP
	// ------------

	// Storage
	storageRepo := storage.NewRepository(c.DB)
	c.StorageService = storage.NewService(storageRepo, c.Migrator, c.DB)

	// Pricing Service
	c.PricingService = pricing.NewService(c.DB)

	// Mailer Service
	c.Mailer = mailer.NewService(&c.Config.MailSender)

	// Companies Service (зависит от Storage)
	c.CompaniesService = companies.NewService(c.DB, c.StorageService, c.PricingService, c.Mailer)

	// Accounts Service (зависит от JWT и Companies)
	c.AccountsService = accounts.NewService(c.DB, c.JWTService, c.CompaniesService, c.Mailer, c.AdminAuthService)

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
		PermService:    c.PermissionService,
		StorageService: c.StorageService,
	}

	// Сохраняем в контейнер
	c.MediaRepo = mediaRepo
	c.MediaService = mediaService
	c.MediaHandlers = mediaHandlers

	// Handlers
	c.AccountsHandlers = accounts.NewHandlers(c.AccountsService)
	c.CompaniesHandlers = companies.NewHandlers(c.CompaniesService)
	c.StorageHandlers = storage.NewHandlers(c.StorageService)
	c.PricingHandlers = pricing.NewHandlers(c.PricingService)
	c.PublicHandlers = public.NewHandlers(c.PublicService)

	// ----------
	// ADMIN
	// ----------
	c.AdminAuthHandlers = adminauth.NewHandlers(c.AdminAuthService)
	c.AdminDbService = admindb.NewService(c.DB, c.CoreWorkersService, c.Migrator)
	c.AdminDbHandlers = admindb.NewHandlers(c.AdminDbService)
	c.AdminAccountsService = adminaccounts.NewService(c.DB, c.AccountsService, c.AdminAuthService)
	c.AdminAccountsHandlers = adminaccounts.NewHandlers(c.AdminAccountsService)

	return nil
}

func (c *Container) initTenantRoutes() error {
	c.TenantRoutes = tenant.NewRoutes(
		c.DB,
		c.StorageService,
		c.AccountsService,
		c.CompaniesService,
	)
	return nil
}

func (c *Container) initAdminRoutes() error {
	c.AdminRoutes = admin.NewRoutes(admin.Deps{
		JWTService:            c.JWTService,
		AdminAuthService:      c.AdminAuthService,
		AdminDbHandlers:       c.AdminDbHandlers,
		AdminAccountsService:  c.AdminAccountsService,
		AdminAccountsHandlers: c.AdminAccountsHandlers,
		AdminAuthHandlers:     c.AdminAuthHandlers,
	})
	return nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.StorageService != nil {
		c.StorageService.CloseAll()
	}
}

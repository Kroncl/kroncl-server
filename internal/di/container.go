// internal/di/container.go
package di

import (
	"context"
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/config"
	"kroncl-server/internal/migrator"
	"kroncl-server/internal/permissioner"
	"kroncl-server/internal/tenant"
	"kroncl-server/internal/tenant/storage"
	"kroncl-server/utils"
	"log"
	"os"
	"path/filepath"

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
	PermissionService *permissioner.Service
	StorageService    *storage.Service
	Migrator          *migrator.Migrator

	// Хэндлеры
	AccountsHandlers  *accounts.Handlers
	CompaniesHandlers *companies.Handlers
	StorageHandlers   *storage.Handlers

	// Tenant модули
	TenantRoutes *tenant.Routes
}

func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	// Инициализация в правильном порядке
	if err := c.initDB(ctx); err != nil {
		return nil, err
	}
	if err := c.initMigrator(); err != nil {
		return nil, err
	}
	if err := c.initServices(); err != nil {
		return nil, err
	}
	if err := c.initTenantRoutes(); err != nil {
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

func (c *Container) initServices() error {
	// JWT
	c.JWTService = auth.NewJWTService(
		c.Config.JWT.SecretKey,
		c.Config.JWT.AccessDuration,
		c.Config.JWT.RefreshDuration,
	)

	// Storage
	storageRepo := storage.NewRepository(c.DB)
	c.StorageService = storage.NewService(storageRepo, c.Migrator, c.DB)

	// Services
	c.AccountsService = accounts.NewService(c.DB, c.JWTService)
	c.CompaniesService = companies.NewService(c.DB, c.StorageService)
	c.PermissionService = permissioner.NewService(c.DB)

	// Handlers
	c.AccountsHandlers = accounts.NewHandlers(c.AccountsService)
	c.CompaniesHandlers = companies.NewHandlers(c.CompaniesService)
	c.StorageHandlers = storage.NewHandlers(c.StorageService)

	return nil
}

func (c *Container) initTenantRoutes() error {
	c.TenantRoutes = tenant.NewRoutes(c.StorageService, c.PermissionService)
	return nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}

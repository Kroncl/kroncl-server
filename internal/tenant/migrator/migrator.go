package migrator

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Migrator struct {
	pool           *pgxpool.Pool
	migrationsPath string
}

func NewMigrator(pool *pgxpool.Pool, migrationsPath string) (*Migrator, error) {
	// Получаем абсолютный путь
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Проверяем существование
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	// Логируем для отладки
	log.Printf("📁 Migrator initialized with path: %s", absPath)

	// Проверяем файлы
	files, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir: %w", err)
	}

	log.Printf("📁 Found %d files in migrations directory:", len(files))
	for _, f := range files {
		log.Printf("  - %s (dir: %v)", f.Name(), f.IsDir())
	}

	return &Migrator{
		pool:           pool,
		migrationsPath: absPath,
	}, nil
}

func (m *Migrator) CreateSchema(ctx context.Context, schemaName string) error {
	// Безопасное экранирование имени схемы
	escapedSchemaName := pgx.Identifier{schemaName}.Sanitize()
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", escapedSchemaName)

	_, err := m.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	log.Printf("Schema created: %s", schemaName)
	return nil
}

func (m *Migrator) buildDSN() (string, error) {
	// Получаем конфигурацию из пула
	config := m.pool.Config()

	var user, password, host, dbname string
	var port uint16

	if config.ConnConfig != nil {
		user = config.ConnConfig.User
		password = config.ConnConfig.Password
		host = config.ConnConfig.Host
		port = config.ConnConfig.Port
		dbname = config.ConnConfig.Database
	}

	// Фолбэк на переменные окружения
	if user == "" {
		user = os.Getenv("DB_USER")
	}
	if password == "" {
		password = os.Getenv("DB_PASSWORD")
	}
	if host == "" {
		host = os.Getenv("DB_HOST")
	}
	if port == 0 {
		portStr := os.Getenv("DB_PORT")
		if portStr != "" {
			if p, err := strconv.ParseUint(portStr, 10, 16); err == nil {
				port = uint16(p)
			} else {
				port = 5432 // default
			}
		} else {
			port = 5432 // default
		}
	}
	if dbname == "" {
		dbname = os.Getenv("DB_NAME")
	}

	// Экранируем специальные символы
	escapedPassword := url.QueryEscape(password)
	escapedUser := url.QueryEscape(user)

	// Формируем базовый DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		escapedUser, escapedPassword, host, port, dbname)

	return dsn, nil
}

func (m *Migrator) getDSNWithSchema(schemaName string) (string, error) {
	baseDSN, err := m.buildDSN()
	if err != nil {
		return "", err
	}

	// Добавляем параметр для указания схемы
	// Используем x-migrations-table чтобы каждая схема имела свою таблицу миграций
	dsnWithSchema := fmt.Sprintf("%s&x-migrations-table=%s.schema_migrations",
		baseDSN, pgx.Identifier{schemaName}.Sanitize())

	log.Printf("📡 DSN for schema %s: %s", schemaName, dsnWithSchema)
	return dsnWithSchema, nil
}

func (m *Migrator) getSourceURL() string {
	// Преобразуем путь для Windows
	absPath := m.migrationsPath

	// Для Windows: file:///C:/Users/...
	// Убедимся, что есть три слеша после file:
	path := strings.ReplaceAll(absPath, "\\", "/")

	// Убираем возможный дублирующийся слеш в начале
	if strings.HasPrefix(path, "//") {
		path = path[1:]
	}

	// Добавляем file:///
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	sourceURL := "file://" + path
	log.Printf("🔧 Generated source URL: %s", sourceURL)
	return sourceURL
}

func (m *Migrator) Up(ctx context.Context, schemaName string) error {
	log.Printf("🚀 Applying migrations to schema: %s", schemaName)
	log.Printf("📁 Using migrations from: %s", m.migrationsPath)

	dsn, err := m.getDSNWithSchema(schemaName)
	if err != nil {
		return fmt.Errorf("failed to build DSN for schema %s: %w", schemaName, err)
	}

	sourceURL := m.getSourceURL()
	log.Printf("📡 Source URL: %s", sourceURL)

	// Создаем мигратор
	migrateInstance, err := migrate.New(sourceURL, dsn)
	if err != nil {
		// Детальная ошибка
		return fmt.Errorf("failed to create migrate instance:\n"+
			"  Source: %s\n"+
			"  DSN: %s\n"+
			"  Error: %w", sourceURL, dsn, err)
	}
	defer migrateInstance.Close()

	// Применяем миграции
	log.Printf("⚡ Applying migrations...")
	if err := migrateInstance.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations for schema %s: %w", schemaName, err)
	}

	if err == migrate.ErrNoChange {
		log.Printf("✅ All migrations already applied for schema: %s", schemaName)
	} else {
		log.Printf("✅ Migrations successfully applied for schema: %s", schemaName)
	}

	return nil
}

// Down откатывает миграции для схемы
func (m *Migrator) Down(ctx context.Context, schemaName string, steps int) error {
	dsn, err := m.getDSNWithSchema(schemaName)
	if err != nil {
		return fmt.Errorf("failed to build DSN: %w", err)
	}

	sourceURL := m.getSourceURL()
	migrateInstance, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer migrateInstance.Close()

	if steps > 0 {
		log.Printf("🔙 Rolling back %d steps for schema: %s", steps, schemaName)
		if err := migrateInstance.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to step down migrations: %w", err)
		}
	} else {
		log.Printf("🔙 Rolling back all migrations for schema: %s", schemaName)
		if err := migrateInstance.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to apply down migrations: %w", err)
		}
	}

	log.Printf("✅ Migrations rolled back for schema: %s", schemaName)
	return nil
}

// CheckSchemaExists проверяет существование схемы
func (m *Migrator) CheckSchemaExists(ctx context.Context, schemaName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`

	var exists bool
	err := m.pool.QueryRow(ctx, query, schemaName).Scan(&exists)
	return exists, err
}

// GetVersion получает текущую версию миграций схемы
func (m *Migrator) GetVersion(ctx context.Context, schemaName string) (uint, bool, error) {
	dsn, err := m.getDSNWithSchema(schemaName)
	if err != nil {
		return 0, false, fmt.Errorf("failed to build DSN: %w", err)
	}

	sourceURL := m.getSourceURL()
	migrateInstance, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer migrateInstance.Close()

	version, dirty, err := migrateInstance.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

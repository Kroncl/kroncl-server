package migrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"kroncl-server/utils"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (m *Migrator) GetPool() *pgxpool.Pool {
	return m.pool
}

func NewMigrator(pool *pgxpool.Pool, cfg Config) (*Migrator, error) {
	absPath, err := filepath.Abs(cfg.MigrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Проверяем существование базовой папки
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	// Проверяем существование папки для типа схемы
	schemaPath := filepath.Join(absPath, string(cfg.SchemaType))
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Создаем директорию, если ее нет
		if err := os.MkdirAll(schemaPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create schema directory: %w", err)
		}
		log.Printf("📁 Created migrations directory: %s", schemaPath)
	}

	// Получаем конфиг БД из пула или окружения
	connConfig := pool.Config().ConnConfig
	dbConfig := utils.DBConfig{
		Host:     connConfig.Host,
		Port:     int(connConfig.Port),
		Username: connConfig.User,
		Password: connConfig.Password,
		Name:     connConfig.Database,
		SSLMode:  "disable",
	}

	log.Printf("📁 Migrator initialized:")
	log.Printf("   - Base path: %s", absPath)
	log.Printf("   - Schema type: %s", cfg.SchemaType)
	log.Printf("   - Database: %s@%s:%d/%s",
		dbConfig.Username, dbConfig.Host, dbConfig.Port, dbConfig.Name)

	return &Migrator{
		pool:     pool,
		basePath: absPath,
		config:   dbConfig,
	}, nil
}

// GetMigrationsPath возвращает путь к миграциям для типа схемы
func (m *Migrator) GetMigrationsPath() string {
	// Если нужно получить путь к конкретной папке
	// (например, для CLI утилиты)
	return filepath.Join(m.basePath, "public") // или "tenant"
}

// CreateSchema создает новую схему (для тенантов)
func (m *Migrator) CreateSchema(ctx context.Context, schemaName string) error {
	escapedSchemaName := pgx.Identifier{schemaName}.Sanitize()
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", escapedSchemaName)

	_, err := m.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	log.Printf("✅ Schema created: %s", schemaName)
	return nil
}

// buildDSN строит DSN для подключения с использованием utils
func (m *Migrator) buildDSN(schemaName string) (string, error) {
	// Используем готовую функцию из utils
	dsn, err := utils.BuildDSN(m.config)
	if err != nil {
		return "", fmt.Errorf("failed to build base DSN: %w", err)
	}

	// Добавляем параметр search_path
	if schemaName != "" && schemaName != "public" {
		// Экранируем имя схемы для URL
		encodedSchemaName := strings.ReplaceAll(schemaName, "\"", "")
		dsn += fmt.Sprintf("&search_path=%s", encodedSchemaName)
	}

	return dsn, nil
}

// Run применяет миграции к схеме
func (m *Migrator) Run(ctx context.Context, schemaName string, command string, steps int) error {
	// Определяем тип схемы по имени
	var schemaType SchemaType
	if schemaName == "public" || schemaName == "" {
		schemaType = SchemaTypePublic
		schemaName = "public"
	} else {
		schemaType = SchemaTypeTenant
	}

	// Путь к миграциям для этого типа схемы
	migrationsPath := filepath.Join(m.basePath, string(schemaType))

	// Проверяем существование папки с миграциями
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsPath)
	}

	// ТОЧНО ТАК ЖЕ КАК В РАБОЧЕЙ УТИЛИТЕ:
	// Исправление для Windows - нужно правильно сформировать file:// URL
	sourcePath := strings.ReplaceAll(migrationsPath, "\\", "/")

	// На Windows нужно добавить ведущий слеш после file://
	// file:///C:/Users/... вместо file://C:\Users\...
	var sourceURL string
	if strings.Contains(sourcePath, ":/") {
		// Windows путь с диском (C:/Users/...)
		sourceURL = fmt.Sprintf("file:///%s", sourcePath)
	} else {
		// Unix путь
		sourceURL = fmt.Sprintf("file://%s", sourcePath)
	}

	log.Printf("📁 Загружаю миграции из: %s", sourceURL)

	// Строим DSN с search_path
	dsn, err := m.buildDSN(schemaName)
	if err != nil {
		return fmt.Errorf("failed to build DSN: %w", err)
	}

	// Для тенантов используем отдельную таблицу миграций
	if schemaType == SchemaTypeTenant {
		dsn += fmt.Sprintf("&x-migrations-table=%s.schema_migrations",
			pgx.Identifier{schemaName}.Sanitize())
	}

	log.Printf("🚀 Running command '%s' for schema '%s'", command, schemaName)
	log.Printf("🔗 DSN: %s", maskPassword(dsn))

	// Создаем инстанс мигратора
	migrateInstance, err := migrate.New(sourceURL, dsn)
	if err != nil {
		// Дополнительная отладочная информация
		log.Printf("🔄 Пробую альтернативный формат пути...")

		// Попробуем относительный путь
		migrateInstance, err = migrate.New("file://migrations/"+string(schemaType), dsn)
		if err != nil {
			return fmt.Errorf("❌ Ошибка создания мигратора: %v\n"+
				"💡 Проверь:\n"+
				"   1. Что папка '%s' существует\n"+
				"   2. Что в ней есть файлы *.up.sql и *.down.sql\n"+
				"   3. Что PostgreSQL запущен\n"+
				"   4. Что DSN корректный: %s", err, migrationsPath, maskPassword(dsn))
		}
	}
	defer migrateInstance.Close()

	// Выполняем команду
	switch command {
	case "up":
		log.Println("⚡ Applying migrations...")
		if err := migrateInstance.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
		if err == migrate.ErrNoChange {
			log.Println("✅ All migrations already applied")
		} else {
			log.Println("✅ Migrations successfully applied")
		}

	case "down":
		if steps > 0 {
			log.Printf("🔙 Rolling back %d steps...", steps)
			if err := migrateInstance.Steps(-steps); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to step down: %w", err)
			}
		} else {
			log.Println("🔙 Rolling back all migrations...")
			if err := migrateInstance.Down(); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to rollback: %w", err)
			}
		}
		log.Println("✅ Migrations rolled back")

	case "drop":
		log.Println("💣 Dropping all tables...")
		if err := migrateInstance.Drop(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to drop: %w", err)
		}
		log.Println("✅ All tables dropped")

	case "version":
		version, dirty, err := migrateInstance.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return fmt.Errorf("failed to get version: %w", err)
		}
		status := "✅ clean"
		if dirty {
			status = "❌ DIRTY"
		}
		log.Printf("📊 Version: %d, status: %s", version, status)

	case "force":
		if steps <= 0 {
			return fmt.Errorf("version must be > 0 for force command")
		}
		log.Printf("🔧 Forcing version %d...", steps)
		if err := migrateInstance.Force(steps); err != nil {
			return fmt.Errorf("failed to force version: %w", err)
		}
		log.Printf("✅ Version forced to %d", steps)

	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}

// Up - удобный метод для применения миграций (используется в бизнес-логике)
func (m *Migrator) Up(ctx context.Context, schemaName string) error {
	return m.Run(ctx, schemaName, "up", 0)
}

// Down - удобный метод для отката миграций
func (m *Migrator) Down(ctx context.Context, schemaName string, steps int) error {
	return m.Run(ctx, schemaName, "down", steps)
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

// DropSchema удаляет схему (для тенантов)
func (m *Migrator) DropSchema(ctx context.Context, schemaName string, cascade bool) error {
	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s", pgx.Identifier{schemaName}.Sanitize())
	if cascade {
		query += " CASCADE"
	}

	_, err := m.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop schema %s: %w", schemaName, err)
	}

	log.Printf("✅ Schema dropped: %s", schemaName)
	return nil
}

// maskPassword скрывает пароль в DSN для логов
func maskPassword(dsn string) string {
	parts := strings.Split(dsn, ":")
	if len(parts) >= 3 {
		// Маскируем пароль: postgres://user:*****@host...
		parts[2] = "*****"
		return strings.Join(parts, ":")
	}
	return dsn
}

// ListTenantSchemas возвращает список всех схем тенантов
func (m *Migrator) ListTenantSchemas(ctx context.Context) ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'company_%' 
		OR schema_name LIKE 'tenant_%'
		ORDER BY schema_name
	`

	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	return schemas, rows.Err()
}

// MigrateAllTenants применяет миграции ко всем тенантам
func (m *Migrator) MigrateAllTenants(ctx context.Context, command string, steps int) error {
	schemas, err := m.ListTenantSchemas(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant schemas: %w", err)
	}

	if len(schemas) == 0 {
		log.Println("ℹ️  No tenant schemas found")
		return nil
	}

	log.Printf("🔧 Found %d tenant schemas", len(schemas))

	successCount := 0
	failCount := 0

	for _, schema := range schemas {
		log.Printf("🚀 Processing tenant: %s", schema)

		err := m.Run(ctx, schema, command, steps)
		if err != nil {
			log.Printf("❌ Failed to migrate %s: %v", schema, err)
			failCount++

			// Можно продолжить с другими схемами или остановиться
			// Если хотите остановиться при первой ошибке, раскомментируйте:
			// return fmt.Errorf("failed to migrate schema %s: %w", schema, err)
		} else {
			log.Printf("✅ Successfully migrated %s", schema)
			successCount++
		}
	}

	log.Printf("📊 Migration summary:")
	log.Printf("   ✅ Success: %d", successCount)
	log.Printf("   ❌ Failed: %d", failCount)
	log.Printf("   📋 Total: %d", len(schemas))

	if failCount > 0 {
		return fmt.Errorf("%d tenant migrations failed", failCount)
	}

	return nil
}

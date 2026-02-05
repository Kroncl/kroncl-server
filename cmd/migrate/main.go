package main

import (
	"context"
	"flag"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/migrator"
	"kroncl-server/utils"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Парсинг флагов
	command := flag.String("cmd", "up", "Команда: up, down, version, force, steps, create, drop")
	// version := flag.Int("version", 0, "Версия для force")
	steps := flag.Int("steps", 0, "Количество шагов")
	name := flag.String("name", "", "Имя для новой миграции")
	schemaType := flag.String("type", "public", "Тип схемы: public или tenant")
	schemaName := flag.String("schema", "public", "Имя схемы (для tenant)")

	flag.Parse()

	// Команда create
	if *command == "create" {
		if *name == "" && len(flag.Args()) > 0 {
			*name = flag.Args()[0]
		}
		if *name == "" {
			log.Fatal("❌ Для команды create необходимо указать имя миграции")
		}
		createMigration(*name, *schemaType)
		return
	}

	// Загрузка окружения
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env не найден, использую переменные окружения")
	}

	// Получаем конфиг БД
	config := config.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(config)
	if err != nil {
		log.Fatalf("❌ Ошибка сборки DSN: %v", err)
	}

	// Подключаемся к БД
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer pool.Close()

	log.Printf("🔗 Подключено к: %s@%s:%d/%s",
		config.Username, config.Host, config.Port, config.Name)

	// Определяем путь к миграциям
	cwd, _ := os.Getwd()
	migrationsPath := filepath.Join(cwd, "migrations")

	// Создаем мигратор
	migratorInstance, err := migrator.NewMigrator(pool, migrator.Config{
		MigrationsPath: migrationsPath,
		SchemaType:     migrator.SchemaType(*schemaType),
	})
	if err != nil {
		log.Fatalf("❌ Ошибка создания мигратора: %v", err)
	}

	// Специальная команда для всех тенантов
	if *schemaType == "all-tenants" {
		if err := migratorInstance.MigrateAllTenants(context.Background(), *command, *steps); err != nil {
			log.Fatalf("❌ Ошибка миграции всех тенантов: %v", err)
		}
		return
	}

	// Выполняем команду
	if err := migratorInstance.Run(context.Background(), *schemaName, *command, *steps); err != nil {
		log.Fatalf("❌ Ошибка выполнения команды '%s': %v", *command, err)
	}
}

func createMigration(name, schemaType string) {
	// Определяем папку для миграции
	migrationsDir := filepath.Join("migrations", schemaType)

	// Создаем директорию если её нет
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationsDir, 0755); err != nil {
			log.Fatalf("❌ Не удалось создать директорию: %v", err)
		}
		log.Printf("📁 Создана директория: %s", migrationsDir)
	}

	// Генерируем timestamp
	timestamp := time.Now().Unix()

	// Создаем файлы
	upFile := filepath.Join(migrationsDir, fmt.Sprintf("%d_%s.up.sql", timestamp, name))
	downFile := filepath.Join(migrationsDir, fmt.Sprintf("%d_%s.down.sql", timestamp, name))

	// Базовые шаблоны
	upTemplate := fmt.Sprintf(`-- Up Migration: %s
-- Type: %s
-- Created: %s

`, name, schemaType, time.Now().Format("2006-01-02 15:04:05"))

	downTemplate := fmt.Sprintf(`-- Down Migration: %s
-- Type: %s
-- Created: %s

`, name, schemaType, time.Now().Format("2006-01-02 15:04:05"))

	// Записываем файлы
	if err := os.WriteFile(upFile, []byte(upTemplate), 0644); err != nil {
		log.Fatalf("❌ Ошибка создания up файла: %v", err)
	}
	if err := os.WriteFile(downFile, []byte(downTemplate), 0644); err != nil {
		log.Fatalf("❌ Ошибка создания down файла: %v", err)
	}

	log.Printf("✅ Созданы файлы миграции:")
	log.Printf("   📁 Папка: %s", migrationsDir)
	log.Printf("   📄 UP:   %s", filepath.Base(upFile))
	log.Printf("   📄 DOWN: %s", filepath.Base(downFile))
}

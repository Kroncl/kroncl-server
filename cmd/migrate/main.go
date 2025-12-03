package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"matrix-authorization-server/utils"
)

func main() {
	// Парсинг флагов
	command := flag.String("cmd", "up", "Команда: up, down, version, force, steps, create")
	version := flag.Int("version", 0, "Версия для force (например: 2)")
	steps := flag.Int("steps", 0, "Количество шагов (+ вверх, - вниз)")
	name := flag.String("name", "", "Имя для новой миграции (для create)")

	flag.Parse()

	// Для команды create не нужна БД
	if *command == "create" {
		if *name == "" {
			log.Fatal("❌ Для команды create необходимо указать имя: -name <имя>")
		}
		createMigration(*name)
		return
	}

	// Загрузка окружения
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env не найден, использую переменные окружения")
	}

	// Подключение к БД
	config := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(config)
	if err != nil {
		log.Fatalf("❌ Ошибка сборки DSN: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer pool.Close()

	// Проверка подключения
	if err := testConnection(pool); err != nil {
		log.Fatalf("❌ Ошибка подключения: %v", err)
	}

	// Создаем мигратор
	migrator, err := createMigrator(pool)
	if err != nil {
		log.Fatalf("❌ Ошибка создания мигратора: %v", err)
	}
	defer migrator.Close()

	// Выполняем команду
	ctx := context.Background()
	if err := runCommand(ctx, migrator, *command, *version, *steps); err != nil {
		log.Fatalf("❌ Ошибка выполнения команды '%s': %v", *command, err)
	}
}

// createMigrator создает экземпляр migrate.Migrate для pgxpool
func createMigrator(pool *pgxpool.Pool) (*migrate.Migrate, error) {
	// Конвертируем pgxpool в *sql.DB через stdlib
	db := stdlib.OpenDBFromPool(pool)

	// Создаем драйвер базы данных
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("создание драйвера: %w", err)
	}

	// Создаем мигратор
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // путь к миграциям
		"postgres", driver,
	)
	if err != nil {
		return nil, fmt.Errorf("создание мигратора: %w", err)
	}

	return m, nil
}

// runCommand выполняет указанную команду
func runCommand(ctx context.Context, m *migrate.Migrate, cmd string, version, steps int) error {
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		if err := m.Up(); err == migrate.ErrNoChange {
			log.Println("✅ Все миграции уже применены")
			return nil
		}
		log.Println("✅ Миграции успешно применены")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		if err := m.Down(); err == migrate.ErrNoChange {
			log.Println("✅ Нет миграций для отката")
			return nil
		}
		log.Println("✅ Миграции успешно откачены")

	case "version":
		v, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return err
		}
		status := "clean"
		if dirty {
			status = "DIRTY (требует force)"
		}
		log.Printf("📊 Версия: %d, статус: %s", v, status)

	case "force":
		if version <= 0 {
			return fmt.Errorf("версия должна быть > 0")
		}
		if err := m.Force(version); err != nil {
			return err
		}
		log.Printf("✅ Версия установлена: %d", version)

	case "steps":
		if steps == 0 {
			return fmt.Errorf("укажите количество шагов (например: -steps 1 или -steps -1)")
		}
		if err := m.Steps(steps); err != nil {
			return err
		}
		direction := "вверх"
		if steps < 0 {
			direction = "вниз"
		}
		log.Printf("✅ Применено %d шагов (%s)", steps, direction)

	default:
		printHelp()
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}

	return nil
}

// testConnection проверяет подключение к БД
func testConnection(pool *pgxpool.Pool) error {
	var version string
	err := pool.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		return err
	}
	log.Printf("✅ PostgreSQL: %s", version)
	return nil
}

// createMigration создает файлы миграции
func createMigration(name string) {
	// Проверяем/создаем директорию
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		if err := os.MkdirAll("migrations", 0755); err != nil {
			log.Fatalf("❌ Не удалось создать директорию: %v", err)
		}
	}

	// Генерируем timestamp (голанг-миграт использует timestamp вместо последовательных номеров)
	timestamp := time.Now().Unix()

	// Создаем файлы
	upFile := fmt.Sprintf("migrations/%d_%s.up.sql", timestamp, name)
	downFile := fmt.Sprintf("migrations/%d_%s.down.sql", timestamp, name)

	// Шаблоны
	upTemplate := `-- Up Migration
-- Добавьте SQL для применения миграции

`
	downTemplate := `-- Down Migration  
-- Добавьте SQL для отката миграции

`

	// Записываем файлы
	if err := os.WriteFile(upFile, []byte(upTemplate), 0644); err != nil {
		log.Fatalf("❌ Ошибка создания up файла: %v", err)
	}
	if err := os.WriteFile(downFile, []byte(downTemplate), 0644); err != nil {
		log.Fatalf("❌ Ошибка создания down файла: %v", err)
	}

	log.Printf("✅ Созданы файлы миграции:")
	log.Printf("   UP:   %s", upFile)
	log.Printf("   DOWN: %s", downFile)
	log.Println("\n⚠️  Заполните SQL команды в созданных файлах!")
}

// printHelp выводит справку
func printHelp() {
	fmt.Println(`
🚀 Утилита миграций для PostgreSQL (golang-migrate)

Использование:
  migrate -cmd up                    # Применить все миграции
  migrate -cmd down                  # Откатить все миграции  
  migrate -cmd version               # Показать текущую версию
  migrate -cmd force -version 2      # Установить конкретную версию
  migrate -cmd steps -steps 1        # Применить одну миграцию вперед
  migrate -cmd steps -steps -1       # Откатить одну миграцию назад
  migrate -cmd create -name add_users # Создать новую миграцию

Примеры:
  go run cmd/migrate/main.go -cmd up
  go run cmd/migrate/main.go -cmd steps -steps 2
  go run cmd/migrate/main.go -cmd create -name create_users_table

Формат файлов:
  migrations/
    ├── 1678901234_create_users.up.sql    # Применение
    └── 1678901234_create_users.down.sql  # Откат
`)
}

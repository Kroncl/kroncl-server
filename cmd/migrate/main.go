package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"

	"matrix-authorization-server/utils"
)

func main() {
	// Парсинг флагов
	command := flag.String("cmd", "up", "Команда: up, down, version, force, steps, create, drop")
	version := flag.Int("version", 0, "Версия для force")
	steps := flag.Int("steps", 0, "Количество шагов")
	name := flag.String("name", "", "Имя для новой миграции")

	flag.Parse()

	// Команда create не требует подключения к БД
	if *command == "create" {
		if *name == "" {
			// Пробуем получить имя из аргументов
			if len(flag.Args()) > 0 {
				*name = flag.Args()[0]
			}
			if *name == "" {
				log.Fatal("❌ Для команды create необходимо указать имя: -name <имя> или как аргумент")
			}
		}
		createMigration(*name)
		return
	}

	// Загрузка окружения
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env не найден, использую переменные окружения")
	}

	// Получаем конфиг БД
	config := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(config)
	if err != nil {
		log.Fatalf("❌ Ошибка сборки DSN: %v", err)
	}

	// Проверяем подключение к БД (логика в runCommand)
	log.Printf("🔗 Подключаюсь к базе: %s@%s:%s/%s",
		config.Username, config.Host, config.Port, config.Name)

	// Получаем абсолютный путь к миграциям
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Ошибка получения текущей директории: %v", err)
	}
	migrationsPath := filepath.Join(cwd, "migrations")

	// Создаём директорию если её нет
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationsPath, 0755); err != nil {
			log.Fatalf("❌ Не удалось создать директорию: %v", err)
		}
		log.Printf("📁 Создана директория для миграций: %s", migrationsPath)
	}

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

	// Создаём мигратор
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		// Дополнительная отладочная информация
		log.Printf("🔄 Пробую альтернативный формат пути...")

		// Попробуем относительный путь
		m, err = migrate.New("file://migrations", dsn)
		if err != nil {
			log.Fatalf("❌ Ошибка создания мигратора: %v\n"+
				"💡 Проверь:\n"+
				"   1. Что папка 'migrations' существует\n"+
				"   2. Что в ней есть файлы *.up.sql и *.down.sql\n"+
				"   3. Что PostgreSQL запущен\n"+
				"   4. Что DSN корректный: %s", err, maskPassword(dsn))
		}
	}
	defer m.Close()

	// Выполняем команду
	if err := runCommand(m, *command, *version, *steps); err != nil {
		log.Fatalf("❌ Ошибка выполнения команды '%s': %v", *command, err)
	}
}

func runCommand(m *migrate.Migrate, cmd string, version, steps int) error {
	switch cmd {
	case "up":
		log.Println("🚀 Применяю миграции...")
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("✅ Все миграции уже применены")
				return nil
			}
			// Проверяем, есть ли вообще миграции
			log.Printf("⚠️  Возможные проблемы:\n" +
				"   - Нет файлов миграций в папке 'migrations'\n" +
				"   - Файлы имеют неверный формат (должны быть: 1234567890_name.up.sql)\n" +
				"   - Ошибка подключения к БД")
			return fmt.Errorf("применение миграций: %w", err)
		}
		log.Println("✅ Миграции успешно применены")

	case "down":
		log.Println("🔙 Откатываю миграции...")
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("✅ Нет миграций для отката")
				return nil
			}
			return fmt.Errorf("откат миграций: %w", err)
		}
		log.Println("✅ Миграции успешно откачены")

	case "drop":
		log.Println("💣 Удаляю все таблицы...")
		if err := m.Drop(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("✅ Нет таблиц для удаления")
				return nil
			}
			return fmt.Errorf("удаление таблиц: %w", err)
		}
		log.Println("✅ Все таблицы удалены")

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Println("📊 Версия: 0 (нет применённых миграций)")
				log.Println("💡 Для создания первой миграции: task migrate:create -- init_accounts")
				return nil
			}
			return fmt.Errorf("получение версии: %w", err)
		}
		status := "✅ clean"
		if dirty {
			status = "❌ DIRTY (требует force)"
		}
		log.Printf("📊 Версия: %d, статус: %s", v, status)

	case "force":
		if version <= 0 {
			return fmt.Errorf("версия должна быть > 0")
		}
		log.Printf("🔧 Устанавливаю версию %d...", version)
		if err := m.Force(version); err != nil {
			return fmt.Errorf("установка версии: %w", err)
		}
		log.Printf("✅ Версия установлена: %d", version)

	case "steps":
		if steps == 0 {
			return fmt.Errorf("укажите количество шагов (например: -steps 1 или -steps -1)")
		}
		direction := "вперёд"
		if steps < 0 {
			direction = "назад"
		}
		log.Printf("📈 Применяю %d шагов (%s)...", steps, direction)
		if err := m.Steps(steps); err != nil {
			return fmt.Errorf("применение шагов: %w", err)
		}
		log.Printf("✅ Применено %d шагов", steps)

	default:
		printHelp()
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}

	return nil
}

func createMigration(name string) {
	// Проверяем/создаем директорию
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationsDir, 0755); err != nil {
			log.Fatalf("❌ Не удалось создать директорию: %v", err)
		}
		log.Printf("📁 Создана директория: %s", migrationsDir)
	}

	// Генерируем timestamp
	timestamp := time.Now().Unix()

	// Создаем файлы
	upFile := fmt.Sprintf("%s/%d_%s.up.sql", migrationsDir, timestamp, name)
	downFile := fmt.Sprintf("%s/%d_%s.down.sql", migrationsDir, timestamp, name)

	// Базовые шаблоны с примером SQL
	upTemplate := fmt.Sprintf(`-- Up Migration: %s
-- Created: %s

-- Пример создания таблицы:
-- CREATE TABLE IF NOT EXISTS accounts (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     email VARCHAR(255) UNIQUE NOT NULL,
--     name VARCHAR(100) NOT NULL,
--     password_hash VARCHAR(255) NOT NULL,
--     auth_type VARCHAR(50) DEFAULT 'password',
--     created_at TIMESTAMPTZ DEFAULT NOW(),
--     updated_at TIMESTAMPTZ DEFAULT NOW()
-- );
-- 
-- CREATE INDEX idx_accounts_email ON accounts(email);

`, name, time.Now().Format("2006-01-02 15:04:05"))

	downTemplate := fmt.Sprintf(`-- Down Migration: %s  
-- Created: %s

-- Пример отката:
-- DROP TABLE IF EXISTS accounts;

`, name, time.Now().Format("2006-01-02 15:04:05"))

	// Записываем файлы
	if err := os.WriteFile(upFile, []byte(upTemplate), 0644); err != nil {
		log.Fatalf("❌ Ошибка создания up файла: %v", err)
	}
	if err := os.WriteFile(downFile, []byte(downTemplate), 0644); err != nil {
		log.Fatalf("❌ Ошибка создания down файла: %v", err)
	}

	log.Printf("✅ Созданы файлы миграции:")
	log.Printf("   📄 UP:   %s", upFile)
	log.Printf("   📄 DOWN: %s", downFile)
	log.Println("\n💡 Заполните SQL команды в созданных файлах!")
	log.Println("   Затем выполните: task up")
}

func printHelp() {
	fmt.Println(`
🚀 Утилита миграций для PostgreSQL

Использование:
  go run cmd/migrate/main.go -cmd up
  go run cmd/migrate/main.go -cmd down
  go run cmd/migrate/main.go -cmd version
  go run cmd/migrate/main.go -cmd force -version 1
  go run cmd/migrate/main.go -cmd steps -steps 2
  go run cmd/migrate/main.go -cmd create -name create_table

Команды:
  up      - Применить все миграции
  down    - Откатить все миграции
  drop    - Удалить все таблицы
  version - Показать текущую версию
  force   - Установить конкретную версию
  steps   - Применить N шагов
  create  - Создать новую миграцию

Примеры через Taskfile:
  task migrate:create -- init_accounts
  task migrate:up
  task migrate:version
  task migrate:steps -- 1

Структура файлов:
  migrations/
    ├── 1701634408_init_accounts.up.sql    # Применение
    └── 1701634408_init_accounts.down.sql  # Откат
`)
}

// maskPassword скрывает пароль в DSN для логов
func maskPassword(dsn string) string {
	// Простая маскировка пароля в строке подключения
	if strings.Contains(dsn, "password=") {
		parts := strings.Split(dsn, "password=")
		if len(parts) > 1 {
			subparts := strings.Split(parts[1], " ")
			if len(subparts) > 0 {
				dsn = parts[0] + "password=*****" + strings.Join(subparts[1:], " ")
			}
		}
	}
	return dsn
}

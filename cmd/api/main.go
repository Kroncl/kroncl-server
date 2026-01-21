package main

import (
	"context"
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/core"
	"kroncl-server/internal/migrator"
	"kroncl-server/internal/permissioner"
	"kroncl-server/internal/tenant/storage"
	"kroncl-server/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Внимание: не удалось загрузить .env файл")
	}

	dbConfig := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(dbConfig)
	if err != nil {
		log.Fatal("Ошибка формирования DSN:", err)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer pool.Close()

	log.Println("✅ Подключение к БД установлено")

	jwtConfig := auth.LoadJWTConfig()
	jwtService := auth.NewJWTService(
		jwtConfig.SecretKey,
		jwtConfig.AccessDuration,
		jwtConfig.RefreshDuration,
	)

	// Инициализация сервисов
	accountsService := accounts.NewService(pool, jwtService)
	accountsHandlers := accounts.NewHandlers(accountsService)

	// Получаем абсолютный путь к папке migrations
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Ошибка получения текущей директории:", err)
	}

	migrationsPath := filepath.Join(cwd, "migrations")
	log.Printf("📁 Путь к миграциям: %s", migrationsPath)

	migratorService, err := migrator.NewMigrator(pool, migrator.Config{
		MigrationsPath: migrationsPath,
		SchemaType:     migrator.SchemaTypeTenant,
	})
	if err != nil {
		log.Fatal("Migrator init error:", err)
	}

	storageRepository := storage.NewRepository(pool)
	storageService := storage.NewService(storageRepository, migratorService)
	storageHandlers := storage.NewHandlers(storageService)
	companiesService := companies.NewService(pool, storageService)
	companiesHandlers := companies.NewHandlers(companiesService)
	permissionService := permissioner.NewService(pool)

	// Создаем роутер
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(core.BaseResponse)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Total-Count", "X-Access-Token", "X-Refresh-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", core.HealthCheck)

		// Auth routes (public)
		r.Route("/account", func(r chi.Router) {
			r.Post("/reg", accountsHandlers.Register)
			r.Get("/check-email-unique", accountsHandlers.CheckEmailUnique)
			r.Post("/auth", accountsHandlers.Login)
			r.Post("/refresh", accountsHandlers.Refresh)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(jwtService.RequireAuth)
				r.Get("/", accountsHandlers.GetProfile)
				r.Post("/confirm", accountsHandlers.ConfirmEmail)
				r.Post("/confirm/resend", accountsHandlers.ResendConfirmationCode)
			})
		})

		// Protected routes (require auth)
		r.Group(func(r chi.Router) {
			r.Use(jwtService.RequireAuth)

			r.Route("/companies", func(r chi.Router) {
				// Company creation
				r.Post("/", companiesHandlers.Create)
				r.Get("/my", companiesHandlers.GetUserCompanies)
				r.Get("/check-slug-unique", companiesHandlers.CheckSlugUnique)

				// Specific company routes
				r.Route("/{id}", func(r chi.Router) {
					// Company context + access check
					r.Use(companies.CompanyMembership(pool))

					r.Get("/", companiesHandlers.GetUserCompanyById)
					r.With(permissioner.RequirePermission(permissionService, "company.update")).Patch("/", companiesHandlers.Update)

					// company storage
					r.Route("/storage", func(r chi.Router) {
						r.Get("/", storageHandlers.Get)
					})

					// TM module
					r.Route("/tm", func(r chi.Router) {
						r.Use(permissioner.RequirePermission(permissionService, "tm.view"))
						// TM handlers will be here
					})

					// HRM module
					r.Route("/hrm", func(r chi.Router) {
						r.Use(permissioner.RequirePermission(permissionService, "hrm.view"))
						// HRM handlers will be here
					})

					// CRM module
					r.Route("/crm", func(r chi.Router) {
						r.Use(permissioner.RequirePermission(permissionService, "crm.view"))
						// CRM handlers will be here
					})
				})
			})
		})
	})

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	addr := host + ":" + port
	log.Printf("🚀 Сервер запущен на http://%s", addr)
	log.Printf("📡 Доступ по:")
	log.Printf("   - localhost: http://localhost:%s", port)
	log.Printf("   - 127.0.0.1: http://127.0.0.1:%s", port)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

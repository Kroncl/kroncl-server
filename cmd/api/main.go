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
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// chanels
	serverErrors := make(chan error, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runServer(ctx, serverErrors); err != nil {
			log.Printf("Ошибка запуска сервера: %v", err)
			serverErrors <- err
		}
	}()

	// Ожидаем сигналов или ошибок
	select {
	case err := <-serverErrors:
		log.Printf("Ошибка в работе сервера: %v", err)
		cancel()
	case sig := <-signals:
		log.Printf("Получен сигнал: %v. Начинаем graceful shutdown...", sig)
		cancel()

		// Даем серверу время на завершение обработки запросов
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Ждем завершения всех горутин
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-shutdownCtx.Done():
			log.Println("Таймаут graceful shutdown, принудительное завершение")
		case <-done:
			log.Println("Все горутины завершили работу")
		}
	}

	log.Println("Сервер остановлен")
}

func runServer(ctx context.Context, serverErrors chan<- error) error {
	err := godotenv.Load()
	if err != nil {
		log.Println("Внимание: не удалось загрузить .env файл")
	}

	dbConfig := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(dbConfig)
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Проверка соединения с БД
	if err := pool.Ping(ctx); err != nil {
		return err
	}
	log.Println("Подключение к БД установлено")

	// close pool
	go func() {
		<-ctx.Done()
		log.Println("Закрытие пула соединений с БД...")
		pool.Close()
	}()

	jwtConfig := auth.LoadJWTConfig()
	jwtService := auth.NewJWTService(
		jwtConfig.SecretKey,
		jwtConfig.AccessDuration,
		jwtConfig.RefreshDuration,
	)

	// system services
	accountsService := accounts.NewService(pool, jwtService)
	accountsHandlers := accounts.NewHandlers(accountsService)

	// migrations
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	migrationsPath := filepath.Join(cwd, "migrations")
	log.Printf("📁 Путь к миграциям: %s", migrationsPath)

	migratorService, err := migrator.NewMigrator(pool, migrator.Config{
		MigrationsPath: migrationsPath,
		SchemaType:     migrator.SchemaTypeTenant,
	})
	if err != nil {
		return err
	}

	// company system services
	storageRepository := storage.NewRepository(pool)
	storageService := storage.NewService(storageRepository, migratorService)
	storageHandlers := storage.NewHandlers(storageService)
	companiesService := companies.NewService(pool, storageService)
	companiesHandlers := companies.NewHandlers(companiesService)
	permissionService := permissioner.NewService(pool)

	// init router
	r := chi.NewRouter()

	// global middleware
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
				r.Patch("/", accountsHandlers.Update)
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

					// по хорошему уже на этом моменте мы
					// должны иметь pool для схемы тенанта
					// и передать этот пул в нужные репозитории модулей

					r.Get("/", companiesHandlers.GetUserCompanyById)
					r.With(permissioner.RequirePermission(permissionService, "company.update")).Patch("/", companiesHandlers.Update)

					// company storage
					r.Route("/storage", func(r chi.Router) {
						r.Get("/", storageHandlers.Get)
						r.With(permissioner.RequirePermission(permissionService, "storage.sources")).Get("/sources", storageHandlers.GetSources)
					})

					// HRM module
					r.Route("/hrm", func(r chi.Router) {
						r.Use(permissioner.RequirePermission(permissionService, "hrm.view"))
					})
				})
			})
		})
	})

	// http + timeoutes
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	addr := host + ":" + port

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// run it
	go func() {
		log.Printf("🚀 Сервер запущен на http://%s", addr)
		log.Printf("📡 Доступ по:")
		log.Printf("   - localhost: http://localhost:%s", port)
		log.Printf("   - 127.0.0.1: http://127.0.0.1:%s", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ Ошибка сервера: %v", err)
			serverErrors <- err
		}
	}()

	// stop signal
	<-ctx.Done()
	log.Println("Получен сигнал завершения, останавливаем сервер...")

	// Graceful shutdown сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при graceful shutdown сервера: %v", err)
		return err
	}

	log.Println("HTTP сервер корректно остановлен")
	return nil
}

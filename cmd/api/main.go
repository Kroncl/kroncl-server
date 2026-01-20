package main

import (
	"context"
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"kroncl-server/utils"
	"log"
	"net/http"
	"os"

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

	// jwt
	jwtService := auth.NewJWTService(
		jwtConfig.SecretKey,
		jwtConfig.AccessDuration,
		jwtConfig.RefreshDuration,
	)

	// Инициализируем сервисы
	accountsService := accounts.NewService(pool, jwtService)
	accountsHandlers := accounts.NewHandlers(accountsService)

	// Создаем роутер
	r := chi.NewRouter()

	// Middleware
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

		// обслуживание
		r.Route("/account", func(r chi.Router) {
			r.Post("/reg", accountsHandlers.Register)                        // публичный
			r.Post("/check-email-unique", accountsHandlers.CheckEmailUniqie) // публичный
			r.Post("/auth", accountsHandlers.Login)                          // публичный
			r.Post("/refresh", accountsHandlers.Refresh)                     // публичный

			// Защищенный маршрут для подтверждения email
			r.Group(func(r chi.Router) {
				r.Use(jwtService.RequireAuth)
				r.Get("/", accountsHandlers.GetProfile)
				r.Post("/confirm", accountsHandlers.ConfirmEmail)
				r.Post("/confirm/resend", accountsHandlers.ResendConfirmationCode)
			})
		})

		// мясо
		r.Group(func(r chi.Router) {
			r.Use(jwtService.RequireAuth)

			r.Route("/tm", func(r chi.Router) {
				// управление транзакциями
			})
			r.Route("/hrm", func(r chi.Router) {
				// управление персоналом
			})
			r.Route("/crm", func(r chi.Router) {
				// управление клиентами
			})
		})
	})

	// Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0" // Слушаем все интерфейсы по умолчанию
	}

	addr := host + ":" + port
	log.Printf("🚀 Сервер запущен на http://%s", addr)

	// Логируем все доступные адреса
	log.Printf("📡 Доступ по:")
	log.Printf("   - localhost: http://localhost:%s", port)
	log.Printf("   - 127.0.0.1: http://127.0.0.1:%s", port)
	log.Printf("   - LAN IP:    http://YOUR_LOCAL_IP:%s", port)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

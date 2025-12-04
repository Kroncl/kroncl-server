package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"matrix-authorization-server/internal/accounts"
	"matrix-authorization-server/internal/auth"
	"matrix-authorization-server/internal/core"
	"matrix-authorization-server/utils"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		log.Println("Внимание: не удалось загрузить .env файл")
	}

	// Получаем конфиг БД
	dbConfig := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(dbConfig)
	if err != nil {
		log.Fatal("Ошибка формирования DSN:", err)
	}

	// Создаем пул соединений
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer pool.Close()

	log.Println("✅ Подключение к БД установлено")

	// Загружаем конфиг JWT
	jwtConfig := auth.LoadJWTConfig()

	// Инициализируем JWT сервис
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
	r.Use(core.BaseResponse) // наш кастомный middleware ДОЛЖЕН БЫТЬ ПОСЛЕ chi middleware

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
		// Публичные маршруты (не требуют аутентификации)
		r.Get("/health", core.HealthCheck)

		// Маршруты аккаунтов - публичные
		r.Route("/account", func(r chi.Router) {
			r.Post("/reg", accountsHandlers.Register) // публичный
			r.Post("/auth", accountsHandlers.Login)   // публичный

			// Защищенный маршрут для подтверждения email
			r.Group(func(r chi.Router) {
				r.Use(jwtService.RequireAuth) // защищаем только confirm
				r.Post("/confirm", accountsHandlers.ConfirmEmail)
			})
		})

		// Приватные маршруты (требуют аутентификации)
		r.Route("/user", func(r chi.Router) {
			r.Use(jwtService.RequireAuth)

			r.Get("/profile", accountsHandlers.GetProfile)
			// другие защищенные маршруты...
		})
	})

	// Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Сервер запущен на http://localhost:%s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

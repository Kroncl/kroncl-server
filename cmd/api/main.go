package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"matrix-authorization-server/internal/accounts"
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
	config := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(config)
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

	// Инициализируем сервисы
	accountsService := accounts.NewService(pool)
	accountsHandlers := accounts.NewHandlers(accountsService)

	// Создаем роутер
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(core.BaseResponse) // Наш кастомный middleware

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API
	r.Route("/api", func(r chi.Router) {

		// Маршруты
		r.Get("/health", core.HealthCheck)

		// Маршруты аккаунтов
		r.Route("/account", func(r chi.Router) {
			r.Post("/reg", accountsHandlers.Register)
			r.Post("/auth", accountsHandlers.Login)
			r.Post("/confirm", accountsHandlers.ConfirmEmail)
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

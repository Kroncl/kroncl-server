package main

import (
	"log"
	"net/http"
	"os"

	"matrix-authorization-server/internal/handlers"
	mymiddleware "matrix-authorization-server/internal/middleware" // наш middleware

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware" // chi middleware
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	// Базовые middleware от chi
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Наш кастомный middleware для стандартного ответа
	r.Use(mymiddleware.BaseResponse)

	// Маршруты
	r.Get("/health", handlers.HealthCheck)

	// Группа accounts
	r.Route("/accounts", func(r chi.Router) {
		r.Post("/register", handlers.Register)
		r.Get("/login", handlers.Login)
		r.Post("/confirm", handlers.ConfirmEmail)
	})

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Сервер запущен на :%s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

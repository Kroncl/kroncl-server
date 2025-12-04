package main

import (
	"log"
	"net/http"
	"os"

	"matrix-authorization-server/internal/accounts"
	"matrix-authorization-server/internal/core"
	customMiddleware "matrix-authorization-server/internal/core" // наш middleware

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
	r.Use(customMiddleware.BaseResponse)

	// Маршруты
	r.Get("/health", core.HealthCheck)

	r.Route("/account", func(r chi.Router) {
		r.Get("/auth", accounts.Login)
		r.Get("/reg", accounts.Register)
		r.Post("/confirm", accounts.ConfirmEmail)
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

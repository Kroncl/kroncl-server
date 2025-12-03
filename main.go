package main

import (
	"context"
	"fmt"
	"log"
	"matrix-authorization-server/internal/accounts"
	"matrix-authorization-server/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// грузим окружение
	err := godotenv.Load()
	if err != nil {
		log.Println("Внимание: не удалось загрузить .env файл, используем переменные окружения")
	}

	// мапим константы из конфига в dsn
	config := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(config)
	if err != nil {
		log.Fatal("Ошибка формирования DSN:", err)
	}

	// инит пула к базе
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer pool.Close()

	// Тестовые данные
	email := "med.serafim@gmail.com"
	name := "Serafim"
	password := "1234*lkjL56" // Пример простого пароля - валидация должна его отвергнуть

	// Пытаемся создать аккаунт
	result, err := accounts.Create(pool, email, name, password)
	if err != nil {
		log.Printf("Ошибка создания аккаунта: %v\n", err)
		return
	}

	if result {
		fmt.Println("Аккаунт успешно создан!")
	} else {
		fmt.Println("Аккаунт не создан.")
	}
}

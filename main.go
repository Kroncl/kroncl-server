package main

import (
	"context"
	"fmt"
	"log"
	"matrix-authorization-server/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// грузим окружение
	err := godotenv.Load()
	if err != nil {
		log.Fatal("пиздец, не можем загрузить .env")
	}

	// мапим константы из конфига в dsn
	config := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(config)
	if err != nil {
		log.Fatal(err)
	}

	// инит пула к базе
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	testConnection(pool)
	defer pool.Close()
}

func testConnection(pool *pgxpool.Pool) {
	var version string
	err := pool.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PostgreSQL version:", version)
}

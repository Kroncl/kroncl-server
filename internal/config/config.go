package config

import (
	"fmt"
	"kroncl-server/utils"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database utils.DBConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type JWTConfig struct {
	SecretKey       string
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func Load() (*Config, error) {
	// Загружаем .env файл
	if err := loadEnvFile(); err != nil {
		log.Printf("⚠️  Warning: %v", err)
	}

	dbConfig := LoadDBConfigFromEnv()

	// Логируем конфиг (без пароля)
	log.Printf("📋 Конфигурация загружена:")
	log.Printf("   - Server: %s:%s", getEnv("HOST", "0.0.0.0"), getEnv("PORT", "8080"))
	log.Printf("   - Database: %s@%s:%d/%s",
		dbConfig.Username,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name)

	return &Config{
		Server: ServerConfig{
			Host:         getEnv("HOST", "0.0.0.0"),
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: dbConfig,
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET", "development-secret-key-change-in-production"),
			AccessDuration:  60 * 24 * time.Minute, // dev
			RefreshDuration: 7 * 24 * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link", "X-Total-Count", "X-Access-Token", "X-Refresh-Token"},
			AllowCredentials: true,
			MaxAge:           300,
		},
	}, nil
}

func loadEnvFile() error {
	// Пробуем несколько возможных мест
	paths := []string{
		".env",
		".env.local",
		"../.env",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				return fmt.Errorf("ошибка загрузки %s: %w", path, err)
			}
			log.Printf("✅ Загружен .env файл: %s", path)
			return nil
		}
	}

	return fmt.Errorf(".env файл не найден")
}

func LoadDBConfigFromEnv() utils.DBConfig {
	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		port = 5432
	}

	return utils.DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     port,
		Name:     getEnv("DB_NAME", "kroncl"),
		Username: getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

package config

import (
	"fmt"
	"kroncl-server/utils"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var WebSocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Config struct {
	Server     ServerConfig
	Database   utils.DBConfig
	JWT        JWTConfig
	CORS       CORSConfig
	MinIO      MinIOConfig
	MailSender MailSenderConfig
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

type MinIOConfig struct {
	RootUser     string
	RootPassword string
	Endpoint     string
	UseSSL       bool
	PublicBucket string
	ExternalHost string
}

type MailSenderConfig struct {
	ApiUrl       string
	ApiKey       string
	NotifyDomain string
}

func Load() (*Config, error) {
	if err := loadEnvFile(); err != nil {
		log.Printf("⚠️  Warning: %v", err)
	}

	dbConfig := LoadDBConfigFromEnv()

	minioConfig := MinIOConfig{
		RootUser:     getEnv("MINIO_ROOT_USER", "minioadmin"),
		RootPassword: getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
		Endpoint:     getEnv("MINIO_ENDPOINT", "minio:9000"),
		UseSSL:       getEnvAsBool("MINIO_USE_SSL", false),
		PublicBucket: getEnv("MINIO_PUBLIC_BUCKET", "public"),
		ExternalHost: getEnv("MINIO_EXTERNAL_HOST", "localhost:9000"),
	}

	mailSenderConfig := MailSenderConfig{
		ApiUrl:       getEnv("UNISENDER_GO_API_URL", "https://goapi.unisender.ru/ru/transactional/api/v1"),
		ApiKey:       getEnv("UNISENDER_GO_API_KEY", ""),
		NotifyDomain: getEnv("UNISENDER_GO_NOTIFY_DOMAIN", "notify@kroncl.com"),
	}
	maskedApiKey := utils.MaskApiKey(mailSenderConfig.ApiKey)

	log.Printf("📋 Конфигурация загружена:")
	log.Printf("   - Server: %s:%s", getEnv("HOST", "0.0.0.0"), getEnv("PORT", "8080"))
	log.Printf("   - Mail Sender: %s:%s", maskedApiKey, mailSenderConfig.NotifyDomain)
	log.Printf("   - Database: %s@%s:%d/%s",
		dbConfig.Username,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name)
	log.Printf("   - MinIO: %s (bucket: %s)", minioConfig.Endpoint, minioConfig.PublicBucket)

	allowedOrigins := getCORSOrigins()

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
			SecretKey:       getEnv("JWT_SECRET_KEY", "development-secret-key-change-in-production"),
			AccessDuration:  parseDuration(getEnv("JWT_ACCESS_DURATION", "24h")),
			RefreshDuration: parseDuration(getEnv("JWT_REFRESH_DURATION", "168h")),
		},
		CORS: CORSConfig{
			AllowedOrigins:   allowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link", "X-Total-Count"},
			AllowCredentials: true,
			MaxAge:           300,
		},
		MinIO:      minioConfig,
		MailSender: mailSenderConfig,
	}, nil
}

func getCORSOrigins() []string {
	env := getEnv("ENV", "development")
	if env == "production" {
		return []string{
			"https://kroncl.com",
			"https://www.kroncl.com",
			"https://kroncl-client.vercel.app",
		}
	}
	return []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://app.localhost:3000",
	}
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("⚠️ Failed to parse duration '%s', using 24h", s)
		return 24 * time.Hour
	}
	return d
}

func loadEnvFile() error {
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
		Username: getEnv("DB_USERNAME", "postgres"),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

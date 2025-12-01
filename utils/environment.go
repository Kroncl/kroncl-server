package utils

import (
	"os"
	"strconv"
)

// LoadDBConfigFromEnv загружает конфиг из переменных окружения
func LoadDBConfigFromEnv() DBConfig {
	// Используем хелпер для получения порта
	port := getEnvAsInt("AUTH_DB_PORT", 5432)

	return DBConfig{
		Host:     getEnv("AUTH_DB_HOST", "localhost"),
		Port:     port,
		Name:     getEnv("AUTH_DB_NAME", ""),
		Username: getEnv("AUTH_DB_USERNAME", ""),
		Password: getEnv("AUTH_DB_PASSWORD", ""),
		SSLMode:  getEnv("AUTH_DB_SSLMODE", "disable"),
	}
}

// getEnv получает переменную окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает переменную окружения как int
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool получает переменную окружения как bool
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

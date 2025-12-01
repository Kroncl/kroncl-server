package utils

import (
	"fmt"
	"net/url"
)

type DBConfig struct {
	Host     string
	Port     int
	Name     string
	Username string
	Password string
	SSLMode  string // "disable", "require", "verify-ca", "verify-full"
}

// BuildDSN создает строку подключения для PostgreSQL
func BuildDSN(config DBConfig) (string, error) {
	// Устанавливаем значения по умолчанию
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}

	// Проверяем обязательные поля
	if config.Username == "" {
		return "", fmt.Errorf("username is required")
	}
	if config.Name == "" {
		return "", fmt.Errorf("database name is required")
	}

	// Вариант 1: URL формат (рекомендуемый)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(config.Username),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		url.QueryEscape(config.Name),
		config.SSLMode,
	)

	return dsn, nil
}

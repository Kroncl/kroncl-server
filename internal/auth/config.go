package auth

import (
	"os"
	"time"
)

type JWTConfig struct {
	SecretKey       string        // Секретный ключ для подписи
	AccessDuration  time.Duration // Время жизни access токена
	RefreshDuration time.Duration // Время жизни refresh токена
}

func LoadJWTConfig() *JWTConfig {
	secretKey := os.Getenv("JWT_SECRET_KEY")

	accessDurationStr := os.Getenv("JWT_ACCESS_DURATION")
	if accessDurationStr == "" {
		accessDurationStr = "15m"
	}

	refreshDurationStr := os.Getenv("JWT_REFRESH_DURATION")
	if refreshDurationStr == "" {
		refreshDurationStr = "168h"
	}

	accessDuration, _ := time.ParseDuration(accessDurationStr)
	refreshDuration, _ := time.ParseDuration(refreshDurationStr)

	return &JWTConfig{
		SecretKey:       secretKey,
		AccessDuration:  accessDuration,
		RefreshDuration: refreshDuration,
	}
}

package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService - сервис для работы с JWT токенами
type JWTService struct {
	secretKey       []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// AccessClaims - claims для access токена
type AccessClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTService создает новый экземпляр JWT сервиса
func NewJWTService(secretKey string, accessDuration, refreshDuration time.Duration) *JWTService {
	return &JWTService{
		secretKey:       []byte(secretKey),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// GenerateAccessToken создает access токен (только user_id)
func (s *JWTService) GenerateAccessToken(userID string) (string, error) {
	claims := &AccessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "matrix-authorization-server",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken создает refresh токен
func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "matrix-authorization-server",
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateAccessToken проверяет access токен
func (s *JWTService) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken валидирует refresh токен
func (s *JWTService) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserIDFromToken извлекает user_id из токена без полной валидации
func (s *JWTService) GetUserIDFromToken(tokenString string) (string, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	// Сначала пробуем как AccessClaims
	token, _, err := parser.ParseUnverified(tokenString, &AccessClaims{})
	if err == nil {
		if claims, ok := token.Claims.(*AccessClaims); ok && claims.UserID != "" {
			return claims.UserID, nil
		}
	}

	// Пробуем как стандартные RegisteredClaims (для refresh токена)
	token, _, err = parser.ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && claims.Subject != "" {
		return claims.Subject, nil
	}

	return "", errors.New("user_id not found in token claims")
}

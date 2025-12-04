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

// Claims - кастомные claims для нашего приложения
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
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

// GenerateAccessToken создает access токен
func (s *JWTService) GenerateAccessToken(userID, email, name string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
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

// ValidateToken проверяет токен и возвращает claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshTokens обновляет пару токенов
func (s *JWTService) RefreshTokens(refreshToken string) (accessToken, newRefreshToken string, err error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	// Генерируем новую пару токенов
	accessToken, err = s.GenerateAccessToken(claims.UserID, claims.Email, claims.Name)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = s.GenerateRefreshToken(claims.UserID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// GetUserIDFromToken извлекает user_id из токена без полной валидации
// (использовать осторожно, только для не критичных операций)
func (s *JWTService) GetUserIDFromToken(tokenString string) (string, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims.UserID, nil
	}

	return "", errors.New("invalid token claims")
}

package accounts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Service - бизнес-логика для работы с аккаунтами
type Service struct {
	pool *pgxpool.Pool
}

// NewService создает новый экземпляр сервиса
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Create создает новый аккаунт
func (s *Service) Create(email, name, password string) (string, error) {
	// Проверяем валидность email
	if !s.validateEmailFormat(email) {
		return "", fmt.Errorf("invalid email format")
	}

	// Проверяем уникальность email
	isUnique, err := s.checkEmailUnique(email)
	if err != nil {
		return "", fmt.Errorf("email uniqueness check failed: %w", err)
	}
	if !isUnique {
		return "", fmt.Errorf("email already exists")
	}

	// Проверяем валидность имени
	if err := s.validateName(name); err != nil {
		return "", fmt.Errorf("name validation failed: %w", err)
	}

	// Проверяем пароль
	if !s.validatePassword(password) {
		return "", fmt.Errorf("password too weak")
	}

	// Хэшируем пароль
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return "", fmt.Errorf("password hashing failed: %w", err)
	}

	// Создаем аккаунт в БД
	return s.createAccountInDB(email, name, hashedPassword)
}

// GetByEmail возвращает аккаунт по email
func (s *Service) GetByEmail(email string) (*Account, error) {
	ctx := context.Background()

	query := `
		SELECT id, email, name, auth_type, status, created_at, updated_at
		FROM accounts 
		WHERE email = $1
	`

	var account Account
	err := s.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&account.Id,
		&account.Email,
		&account.Name,
		&account.AuthType,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("account not found or database error: %w", err)
	}

	return &account, nil
}

// Приватные вспомогательные методы
func (s *Service) validateEmailFormat(email string) bool {
	// Простая проверка формата
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (s *Service) checkEmailUnique(email string) (bool, error) {
	ctx := context.Background()
	var count int
	query := `SELECT COUNT(*) FROM accounts WHERE email = $1`

	err := s.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database query failed: %w", err)
	}

	return count == 0, nil
}

func (s *Service) validateName(name string) error {
	if len(name) < 2 {
		return fmt.Errorf("name too short")
	}
	if len(name) > 100 {
		return fmt.Errorf("name too long")
	}
	return nil
}

func (s *Service) validatePassword(password string) bool {
	// Минимальные требования
	return len(password) >= 8
}

func (s *Service) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *Service) createAccountInDB(email, name, hashedPassword string) (string, error) {
	ctx := context.Background()

	// Генерируем UUID
	uuid, err := s.generateUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	currentTime := time.Now().Format(time.RFC3339)

	query := `
		INSERT INTO accounts (id, email, name, password_hash, auth_type, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id string
	err = s.pool.QueryRow(
		ctx,
		query,
		uuid,
		strings.ToLower(email),
		name,
		hashedPassword,
		"password", // auth_type
		currentTime,
		currentTime,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}

	return id, nil
}

func (s *Service) generateUUID() (string, error) {
	// Используем pgx для генерации UUID
	ctx := context.Background()
	var uuid string
	err := s.pool.QueryRow(ctx, "SELECT gen_random_uuid()").Scan(&uuid)
	return uuid, err
}

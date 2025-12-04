package accounts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"matrix-authorization-server/internal/auth"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Service - бизнес-логика для работы с аккаунтами
type Service struct {
	pool       *pgxpool.Pool
	jwtService *auth.JWTService
}

// NewService создает новый экземпляр сервиса
func NewService(pool *pgxpool.Pool, jwtService *auth.JWTService) *Service {
	return &Service{
		pool:       pool,
		jwtService: jwtService,
	}
}

// Create создает новый аккаунт и возвращает токены
func (s *Service) Create(email, name, password string) (*Account, string, string, error) {
	// Валидация email
	if !s.validateEmailFormat(email) {
		return nil, "", "", fmt.Errorf("invalid email format")
	}

	// Проверка уникальности email
	isUnique, err := s.checkEmailUnique(email)
	if err != nil {
		return nil, "", "", fmt.Errorf("email uniqueness check failed: %w", err)
	}
	if !isUnique {
		return nil, "", "", fmt.Errorf("email already exists")
	}

	// Валидация имени
	if err := s.validateName(name); err != nil {
		return nil, "", "", fmt.Errorf("name validation failed: %w", err)
	}

	// Валидация пароля
	if !s.validatePassword(password) {
		return nil, "", "", fmt.Errorf("password too weak")
	}

	// Хэширование пароля
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("password hashing failed: %w", err)
	}

	// Создание аккаунта в БД
	account, err := s.createAccountInDB(email, name, hashedPassword)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create account: %w", err)
	}

	// Генерация JWT токенов
	accessToken, err := s.jwtService.GenerateAccessToken(account.ID, account.Email, account.Name)
	if err != nil {
		return account, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(account.ID)
	if err != nil {
		return account, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Генерация и отправка кода подтверждения
	code, err := GenerateConfirmationCode(s.pool, account.ID, "email_confirmation", 6, 15)
	if err != nil {
		// Логируем ошибку, но не прерываем регистрацию
		fmt.Printf("⚠️ Failed to generate confirmation code: %v\n", err)
	} else {
		// Отправляем email (заглушка)
		go func() {
			if err := SendConfirmationEmail(account.Email, code); err != nil {
				fmt.Printf("⚠️ Failed to send confirmation email: %v\n", err)
			} else {
				fmt.Printf("✅ Confirmation email sent to %s\n", account.Email)
			}
		}()
	}

	return account, accessToken, refreshToken, nil
}

// Authenticate проверяет логин/пароль и возвращает токены
func (s *Service) Authenticate(email, password string) (*Account, string, string, error) {
	// Находим аккаунт
	account, err := s.GetByEmail(email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	// Получаем хэш пароля из БД
	hashedPassword, err := s.getPasswordHash(account.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("authentication failed")
	}

	// Проверяем пароль
	if !s.verifyPassword(hashedPassword, password) {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	// Проверяем статус аккаунта
	if account.Status != "confirmed" {
		return nil, "", "", fmt.Errorf("account not confirmed")
	}

	// Генерируем токены
	accessToken, err := s.jwtService.GenerateAccessToken(account.ID, account.Email, account.Name)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(account.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return account, accessToken, refreshToken, nil
}

// ConfirmEmail подтверждает email по коду
func (s *Service) ConfirmEmail(userID, code string) error {
	// Проверяем код
	valid, err := VerifyConfirmationCode(s.pool, userID, code, "email_confirmation")
	if err != nil {
		return fmt.Errorf("confirmation code verification failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid or expired confirmation code")
	}

	// Обновляем статус аккаунта
	return s.markAccountAsConfirmed(userID)
}

// Вспомогательные методы
func (s *Service) createAccountInDB(email, name, hashedPassword string) (*Account, error) {
	ctx := context.Background()

	// Генерируем UUID
	uuid, err := s.generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	currentTime := time.Now()

	query := `
		INSERT INTO accounts (id, email, name, password_hash, auth_type, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, email, name, auth_type, status, created_at, updated_at
	`

	var account Account
	err = s.pool.QueryRow(
		ctx,
		query,
		uuid,
		strings.ToLower(email),
		name,
		hashedPassword,
		"password",
		"waiting", // статус по умолчанию
		currentTime,
		currentTime,
	).Scan(
		&account.ID,
		&account.Email,
		&account.Name,
		&account.AuthType,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &account, nil
}

func (s *Service) GetByEmail(email string) (*Account, error) {
	ctx := context.Background()

	query := `
		SELECT id, email, name, auth_type, status, created_at, updated_at
		FROM accounts 
		WHERE email = $1
	`

	var account Account
	err := s.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&account.ID,
		&account.Email,
		&account.Name,
		&account.AuthType,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	return &account, nil
}

func (s *Service) getPasswordHash(userID string) (string, error) {
	ctx := context.Background()

	var passwordHash string
	err := s.pool.QueryRow(ctx,
		`SELECT password_hash FROM accounts WHERE id = $1`,
		userID,
	).Scan(&passwordHash)

	return passwordHash, err
}

func (s *Service) markAccountAsConfirmed(userID string) error {
	ctx := context.Background()

	query := `
		UPDATE accounts 
		SET status = 'confirmed', updated_at = NOW()
		WHERE id = $1
	`

	_, err := s.pool.Exec(ctx, query, userID)
	return err
}

func (s *Service) validateEmailFormat(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (s *Service) checkEmailUnique(email string) (bool, error) {
	ctx := context.Background()
	var count int
	query := `SELECT COUNT(*) FROM accounts WHERE email = $1`

	err := s.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(&count)
	if err != nil {
		return false, err
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
	return len(password) >= 8
}

func (s *Service) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *Service) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *Service) generateUUID() (string, error) {
	ctx := context.Background()
	var uuid string
	err := s.pool.QueryRow(ctx, "SELECT gen_random_uuid()").Scan(&uuid)
	return uuid, err
}

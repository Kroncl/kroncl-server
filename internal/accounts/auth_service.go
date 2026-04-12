package accounts

import (
	"context"
	"fmt"
	"kroncl-server/internal/mailer"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Create создает новый аккаунт и возвращает токены
func (s *Service) Create(ctx context.Context, email, name, password string) (*Account, string, string, error) {
	// Валидация email
	if !s.validateEmailFormat(email) {
		return nil, "", "", fmt.Errorf("invalid email format")
	}

	// Проверка уникальности email
	isUnique, err := s.checkEmailUnique(ctx, email)
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
	account, err := s.createAccountInDB(ctx, email, name, hashedPassword)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create account: %w", err)
	}

	// Генерация JWT токенов
	accessToken, err := s.jwtService.GenerateAccessToken(account.ID)
	if err != nil {
		return account, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(account.ID)
	if err != nil {
		return account, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	_, err = s.GenerateAndSendCode(ctx, account)

	if err != nil {
		return account, "", "", fmt.Errorf("Error sending the code to the mail")
	}

	return account, accessToken, refreshToken, nil
}

// Authenticate проверяет логин/пароль и возвращает токены
func (s *Service) Authenticate(ctx context.Context, email, password string) (*Account, string, string, error) {
	// Находим аккаунт
	account, err := s.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	// Получаем хэш пароля из БД
	hashedPassword, err := s.getPasswordHash(ctx, account.ID)
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
	accessToken, err := s.jwtService.GenerateAccessToken(account.ID)
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
func (s *Service) ConfirmEmail(ctx context.Context, userID, code string) error {
	account, err := s.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем код
	valid, err := s.VerifyConfirmationCode(ctx, userID, code, "email_confirmation")
	if err != nil {
		return fmt.Errorf("confirmation code verification failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid or expired confirmation code")
	}

	// Отправляем письмо асинхронно
	go func() {
		bgCtx := context.Background()

		data := &mailer.RegistrationSuccessData{
			UserEmail: account.Email,
			UserName:  account.Name,
		}

		s.mailer.SendRegistrationSuccess(bgCtx, data)
	}()

	// Обновляем статус аккаунта
	return s.markAccountAsConfirmed(ctx, userID)
}

// ResendConfirmationCode повторно отправляет код подтверждения
func (s *Service) ResendConfirmationCode(ctx context.Context, userID string) error {
	account, err := s.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if account.Status != "waiting" {
		return fmt.Errorf("account cannot be verified: current status is %s", account.Status)
	}

	// Генерируем новый код
	code, err := s.GenerateConfirmationCode(ctx, account.ID, "email_confirmation", 6, 15)
	if err != nil {
		return fmt.Errorf("failed to generate confirmation code: %w", err)
	}

	// Получаем время истечения
	activeCode, err := s.GetActiveCode(ctx, account.ID, "email_confirmation")
	if err != nil {
		return fmt.Errorf("failed to get active code: %w", err)
	}

	// Отправляем письмо асинхронно
	go func() {
		bgCtx := context.Background()

		data := &mailer.ConfirmationCodeData{
			UserEmail: account.Email,
			UserName:  account.Name,
			Code:      code,
			ExpiresAt: activeCode.ExpiresAt,
		}

		s.mailer.SendConfirmationCodeResend(bgCtx, data)
	}()

	return nil
}

// Вспомогательные методы
func (s *Service) createAccountInDB(ctx context.Context, email, name, hashedPassword string) (*Account, error) {

	// Генерируем UUID
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	currentTime := time.Now()

	query := `
		INSERT INTO accounts (
			id, email, name, password_hash, auth_type, status, 
			created_at, updated_at
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING 
			id, email, name, auth_type, status, 
			created_at, updated_at, 
			COALESCE(avatar_url, '') as avatar_url
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
		&account.AvatarURL,
	)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &account, nil
}

// RefreshTokens обновляет пару токенов по refresh токену
func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	// Валидируем refresh токен и получаем user_id
	userID, err := s.jwtService.GetUserIDFromToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Получаем информацию о пользователе
	account, err := s.GetByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	// Проверяем, что аккаунт подтвержден
	if account.Status != "confirmed" {
		return "", "", fmt.Errorf("account not confirmed")
	}

	// Генерируем новую пару токенов
	accessToken, err = s.jwtService.GenerateAccessToken(account.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err = s.jwtService.GenerateRefreshToken(account.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (s *Service) getPasswordHash(ctx context.Context, userID string) (string, error) {
	var passwordHash string
	err := s.pool.QueryRow(ctx,
		`SELECT password_hash FROM accounts WHERE id = $1`,
		userID,
	).Scan(&passwordHash)

	return passwordHash, err
}

func (s *Service) markAccountAsConfirmed(ctx context.Context, userID string) error {
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

func (s *Service) checkEmailUnique(ctx context.Context, email string) (bool, error) {
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

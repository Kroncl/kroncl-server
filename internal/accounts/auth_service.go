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

const EMAIL_CONFIRMATION_TYPE = "email_confirmation"

func (s *Service) Create(ctx context.Context, email, name, password string) (*Account, string, string, error) {
	if !s.validateEmailFormat(email) {
		return nil, "", "", fmt.Errorf("invalid email format")
	}

	isUnique, err := s.checkEmailUnique(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("email uniqueness check failed: %w", err)
	}
	if !isUnique {
		return nil, "", "", fmt.Errorf("email already exists")
	}

	if err := s.validateName(name); err != nil {
		return nil, "", "", fmt.Errorf("name validation failed: %w", err)
	}

	if !s.validatePassword(password) {
		return nil, "", "", fmt.Errorf("password too weak")
	}

	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("password hashing failed: %w", err)
	}

	account, err := s.createAccountInDB(ctx, email, name, hashedPassword)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create account: %w", err)
	}

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

func (s *Service) Authenticate(ctx context.Context, email, password string) (*Account, string, string, error) {
	account, err := s.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	hashedPassword, err := s.getPasswordHash(ctx, account.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("authentication failed")
	}

	if !s.verifyPassword(hashedPassword, password) {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	if account.Status != ACCOUNT_STATUS_CONFIRMED {
		return nil, "", "", fmt.Errorf("account not confirmed")
	}

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

func (s *Service) ConfirmEmail(ctx context.Context, userID, code string) error {
	account, err := s.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	valid, err := s.VerifyConfirmationCode(ctx, userID, code, EMAIL_CONFIRMATION_TYPE)
	if err != nil {
		return fmt.Errorf("confirmation code verification failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid or expired confirmation code")
	}

	go func() {
		bgCtx := context.Background()

		data := &mailer.RegistrationSuccessData{
			UserEmail: account.Email,
			UserName:  account.Name,
		}

		s.mailer.SendRegistrationSuccess(bgCtx, data)
	}()

	return s.markAccountAsConfirmed(ctx, userID)
}

func (s *Service) ResendConfirmationCode(ctx context.Context, userID string) error {
	account, err := s.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if account.Status != ACCOUNT_STATUS_WAITING {
		return fmt.Errorf("account cannot be verified: current status is %s", account.Status)
	}

	code, err := s.GenerateConfirmationCode(ctx, account.ID, EMAIL_CONFIRMATION_TYPE, 6, 15)
	if err != nil {
		return fmt.Errorf("failed to generate confirmation code: %w", err)
	}

	activeCode, err := s.GetActiveCode(ctx, account.ID, EMAIL_CONFIRMATION_TYPE)
	if err != nil {
		return fmt.Errorf("failed to get active code: %w", err)
	}

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

func (s *Service) createAccountInDB(ctx context.Context, email, name, hashedPassword string) (*Account, error) {
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
		ACCOUNT_STATUS_WAITING,
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

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	userID, err := s.jwtService.GetUserIDFromToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	account, err := s.GetByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	if account.Status != ACCOUNT_STATUS_CONFIRMED {
		return "", "", fmt.Errorf("account not confirmed")
	}

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

func (s *Service) ResetPassword(ctx context.Context, accountID, newPassword string) error {
	if !s.validatePassword(newPassword) {
		return fmt.Errorf("password too weak")
	}

	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	query := `
		UPDATE accounts 
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.pool.Exec(ctx, query, hashedPassword, accountID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

// --------
// UTILS
// --------

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
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := s.pool.Exec(ctx, query, ACCOUNT_STATUS_CONFIRMED, userID)
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

package accounts

import (
	"context"
	"fmt"
	"kroncl-server/internal/mailer"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) GenerateAndSendCode(ctx context.Context, account *Account) (bool, error) {
	code, err := s.GenerateConfirmationCode(ctx, account.ID, "email_confirmation", 6, 15)
	if err != nil {
		return false, fmt.Errorf("failed to generate confirmation code: %w", err)
	}

	activeCode, err := s.GetActiveCode(ctx, account.ID, "email_confirmation")
	if err != nil {
		return false, fmt.Errorf("failed to get active code: %w", err)
	}

	go func() {
		bgCtx := context.Background()

		data := &mailer.ConfirmationCodeData{
			UserEmail: account.Email,
			UserName:  account.Name,
			Code:      code,
			ExpiresAt: activeCode.ExpiresAt,
		}

		s.mailer.SendConfirmationCode(bgCtx, data)
	}()

	return true, nil
}

func (s *Service) GenerateConfirmationCode(ctx context.Context, accountID, codeType string, length, expiryMinutes int) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	markQuery := `
		UPDATE confirmation_codes 
		SET used = TRUE
		WHERE account_id = $1 
		  AND type = $2
		  AND used = FALSE
	`
	_, err = tx.Exec(ctx, markQuery, accountID, codeType)
	if err != nil {
		return "", fmt.Errorf("failed to mark old codes as used: %w", err)
	}

	code := generateRandomCode(length)
	codeHash, err := s.hashCode(code)
	if err != nil {
		return "", fmt.Errorf("failed to hash code: %w", err)
	}

	insertQuery := `
		INSERT INTO confirmation_codes (account_id, code_hash, type, expires_at)
		VALUES ($1, $2, $3, NOW() + $4 * INTERVAL '1 minute')
		RETURNING code_hash
	`

	var resultHash string
	err = tx.QueryRow(ctx, insertQuery, accountID, codeHash, codeType, expiryMinutes).Scan(&resultHash)
	if err != nil {
		return "", fmt.Errorf("failed to create confirmation code: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return code, nil
}

func (s *Service) VerifyConfirmationCode(ctx context.Context, accountID, code, codeType string) (bool, error) {
	query := `
		SELECT code_hash
		FROM confirmation_codes 
		WHERE account_id = $1 
		  AND type = $2
		  AND used = FALSE
		  AND expires_at > NOW()
	`

	rows, err := s.pool.Query(ctx, query, accountID, codeType)
	if err != nil {
		return false, fmt.Errorf("failed to query codes: %w", err)
	}
	defer rows.Close()

	var matchingHash string
	var found bool

	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			continue
		}
		if s.verifyCode(hash, code) {
			matchingHash = hash
			found = true
			break
		}
	}

	if !found {
		return false, nil
	}

	updateQuery := `
		UPDATE confirmation_codes 
		SET used = TRUE
		WHERE account_id = $1 
		  AND code_hash = $2
		  AND type = $3
		  AND used = FALSE
	`
	_, err = s.pool.Exec(ctx, updateQuery, accountID, matchingHash, codeType)
	if err != nil {
		return false, fmt.Errorf("failed to mark code as used: %w", err)
	}

	return true, nil
}

func (s *Service) GetActiveCode(ctx context.Context, accountID, codeType string) (*ConfirmationCode, error) {
	query := `
		SELECT id, account_id, code_hash, type, expires_at, used, created_at
		FROM confirmation_codes 
		WHERE account_id = $1 
		  AND type = $2
		  AND used = FALSE
		  AND expires_at > NOW()
		LIMIT 1
	`

	var code ConfirmationCode
	err := s.pool.QueryRow(ctx, query, accountID, codeType).Scan(
		&code.ID,
		&code.AccountID,
		&code.CodeHash,
		&code.Type,
		&code.ExpiresAt,
		&code.Used,
		&code.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("no active code found: %w", err)
	}

	return &code, nil
}

func (s *Service) CleanupExpiredCodes(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM confirmation_codes 
		WHERE expires_at < NOW() - INTERVAL '1 hour'
		OR (used = TRUE AND created_at < NOW() - INTERVAL '24 hours')
	`

	result, err := s.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired codes: %w", err)
	}

	return result.RowsAffected(), nil
}

func generateRandomCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)

	rand.Seed(time.Now().UnixNano())
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}

	return string(code)
}

func (s *Service) hashCode(code string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *Service) verifyCode(hash, code string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(code))
	return err == nil
}

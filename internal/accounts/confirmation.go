package accounts

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ConfirmationCode struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Code      string    `json:"code"`
	Type      string    `json:"type"` // email_confirmation, password_reset, etc.
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// GenerateConfirmationCode создает код подтверждения для пользователя
func GenerateConfirmationCode(pool *pgxpool.Pool, accountID, codeType string, length, expiryMinutes int) (string, error) {
	ctx := context.Background()

	// Генерируем случайный код
	code := generateRandomCode(length)

	// Удаляем старые коды того же типа
	query := `
		DELETE FROM confirmation_codes 
		WHERE account_id = $1 AND type = $2
	`
	_, err := pool.Exec(ctx, query, accountID, codeType)
	if err != nil {
		return "", fmt.Errorf("failed to cleanup old codes: %w", err)
	}

	// Вставляем новый код
	insertQuery := `
		INSERT INTO confirmation_codes (account_id, code, type, expires_at)
		VALUES ($1, $2, $3, NOW() + $4 * INTERVAL '1 minute')
		RETURNING code
	`

	var resultCode string
	err = pool.QueryRow(ctx, insertQuery, accountID, code, codeType, expiryMinutes).Scan(&resultCode)
	if err != nil {
		return "", fmt.Errorf("failed to create confirmation code: %w", err)
	}

	return resultCode, nil
}

// VerifyConfirmationCode проверяет код подтверждения
func VerifyConfirmationCode(pool *pgxpool.Pool, accountID, code, codeType string) (bool, error) {
	ctx := context.Background()

	query := `
		UPDATE confirmation_codes 
		SET used = TRUE
		WHERE account_id = $1 
		  AND code = $2
		  AND type = $3
		  AND used = FALSE
		  AND expires_at > NOW()
		RETURNING id
	`

	var id string
	err := pool.QueryRow(ctx, query, accountID, code, codeType).Scan(&id)

	if err != nil {
		// Код не найден или не валиден
		return false, nil
	}

	return true, nil
}

// GetActiveCode возвращает активный код для пользователя
func GetActiveCode(pool *pgxpool.Pool, accountID, codeType string) (*ConfirmationCode, error) {
	ctx := context.Background()

	query := `
		SELECT id, account_id, code, type, expires_at, used, created_at
		FROM confirmation_codes 
		WHERE account_id = $1 
		  AND type = $2
		  AND used = FALSE
		  AND expires_at > NOW()
		LIMIT 1
	`

	var code ConfirmationCode
	err := pool.QueryRow(ctx, query, accountID, codeType).Scan(
		&code.ID,
		&code.AccountID,
		&code.Code,
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

// CleanupExpiredCodes удаляет устаревшие коды
func CleanupExpiredCodes(pool *pgxpool.Pool) (int64, error) {
	ctx := context.Background()

	query := `
		DELETE FROM confirmation_codes 
		WHERE expires_at < NOW() - INTERVAL '1 hour'
		OR (used = TRUE AND created_at < NOW() - INTERVAL '24 hours')
	`

	result, err := pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired codes: %w", err)
	}

	return result.RowsAffected(), nil
}

// generateRandomCode генерирует случайный цифровой код
func generateRandomCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)

	rand.Seed(time.Now().UnixNano())
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}

	return string(code)
}

// SendConfirmationEmail отправляет код на email (заглушка)
func SendConfirmationEmail(email, code string) error {
	// Здесь должна быть реализация отправки email
	// Например, через SMTP, SendGrid, Mailgun и т.д.

	fmt.Printf("📧 Отправляем код %s на email %s\n", code, email)
	// В реальности: отправляем email с кодом

	return nil
}

package accounts

import (
	"context"
	"fmt"
	"math/rand"
	"time"
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

// Генерация, очистка старых + отправка нового кода подтверждения
func (s *Service) GenerateAndSendCode(ctx context.Context, account *Account) (bool, error) {
	// Генерация и отправка кода подтверждения
	code, err := s.GenerateConfirmationCode(ctx, account.ID, "email_confirmation", 6, 15)
	if err != nil {
		// Логируем ошибку, но не прерываем регистрацию
		fmt.Printf("⚠️ Failed to generate confirmation code: %v\n", err)
	} else {
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return // Контекст отменен
			default:
				if err := SendConfirmationEmail(account.Email, code); err != nil {
					// Логируем, но не паникуем
					fmt.Printf("⚠️ Failed to send confirmation email: %v\n", err)
				}
			}
		}(context.WithoutCancel(ctx))
	}
	return true, nil
}

// GenerateConfirmationCode создает код подтверждения для пользователя
func (s *Service) GenerateConfirmationCode(ctx context.Context, accountID, codeType string, length, expiryMinutes int) (string, error) {
	// Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // В случае ошибки откатываем

	// Сначала помечаем все старые коды как использованные (а не удаляем!)
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

	// Генерируем случайный код
	code := generateRandomCode(length)

	// Вставляем новый код
	insertQuery := `
		INSERT INTO confirmation_codes (account_id, code, type, expires_at)
		VALUES ($1, $2, $3, NOW() + $4 * INTERVAL '1 minute')
		RETURNING code
	`

	var resultCode string
	err = tx.QueryRow(ctx, insertQuery, accountID, code, codeType, expiryMinutes).Scan(&resultCode)
	if err != nil {
		return "", fmt.Errorf("failed to create confirmation code: %w", err)
	}

	// Коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return resultCode, nil
}

// VerifyConfirmationCode проверяет код подтверждения
func (s *Service) VerifyConfirmationCode(ctx context.Context, accountID, code, codeType string) (bool, error) {
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
	err := s.pool.QueryRow(ctx, query, accountID, code, codeType).Scan(&id)

	if err != nil {
		// Код не найден или не валиден
		return false, nil
	}

	return true, nil
}

// GetActiveCode возвращает активный код для пользователя
func (s *Service) GetActiveCode(ctx context.Context, accountID, codeType string) (*ConfirmationCode, error) {
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
	err := s.pool.QueryRow(ctx, query, accountID, codeType).Scan(
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
	fmt.Printf("📧 Отправляем код %s на email %s\n", code, email)

	return nil
}

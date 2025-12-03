package accounts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Create создает новый аккаунт
func Create(pool *pgxpool.Pool, email string, name string, password string) (string, error) {
	// Проверяем валидность email
	ok, err := validateEmail(email)
	if err != nil {
		return "", fmt.Errorf("email validation failed: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("invalid email format")
	}

	// Проверяем уникальность email в БД
	isUnique, err := checkEmailUniqueDB(pool, email)
	if err != nil {
		return "", fmt.Errorf("email uniqueness check failed: %w", err)
	}
	if !isUnique {
		return "", fmt.Errorf("email already exists")
	}

	// Проверяем валидность имени
	if err := validateName(name); err != nil {
		return "", fmt.Errorf("name validation failed: %w", err)
	}

	// Проверяем пароль
	ok, err = validatePassword(password)
	if err != nil {
		return "", fmt.Errorf("password validation failed: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("password too weak")
	}

	// Хэшируем пароль
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return "", fmt.Errorf("password hashing failed: %w", err)
	}

	// Создаем аккаунт в БД
	accountID, err := createAccountInDB(pool, email, name, hashedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to create account: %w", err)
	}

	return accountID, nil
}

// Создание аккаунта в БД
// в случае успеха возвращает id аккаунта
func createAccountInDB(pool *pgxpool.Pool, email string, name string, hashedPassword string) (string, error) {
	ctx := context.Background()

	// Генерируем UUID
	uuid, err := generateUUID()
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
	err = pool.QueryRow(
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

// GetByEmail возвращает аккаунт по email
func GetByEmail(pool *pgxpool.Pool, email string) (*Account, error) {
	ctx := context.Background()

	query := `
		SELECT id, email, name, auth_type, created_at, updated_at
		FROM accounts 
		WHERE email = $1
	`

	var account Account
	err := pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&account.Id,
		&account.Email,
		&account.Name,
		&account.AuthType,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("account not found or database error: %w", err)
	}

	return &account, nil
}

// GetByID возвращает аккаунт по ID
func GetByID(pool *pgxpool.Pool, id string) (*Account, error) {
	ctx := context.Background()

	query := `
		SELECT id, email, name, auth_type, created_at, updated_at
		FROM accounts 
		WHERE id = $1
	`

	var account Account
	err := pool.QueryRow(ctx, query, id).Scan(
		&account.Id,
		&account.Email,
		&account.Name,
		&account.AuthType,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("account not found or database error: %w", err)
	}

	return &account, nil
}

// UpdatePassword обновляет пароль аккаунта
func UpdatePassword(pool *pgxpool.Pool, accountID string, newPassword string) error {
	// Проверяем пароль
	ok, err := validatePassword(newPassword)
	if err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("password too weak")
	}

	// Хэшируем пароль
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	ctx := context.Background()
	currentTime := time.Now().Format(time.RFC3339)

	query := `
		UPDATE accounts 
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = pool.Exec(ctx, query, hashedPassword, currentTime, accountID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// проверяет соответствие пароля хэшу
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// проверка уникальности email в базе данных
func checkEmailUniqueDB(pool *pgxpool.Pool, email string) (bool, error) {
	ctx := context.Background()

	var count int
	query := `SELECT COUNT(*) FROM accounts WHERE email = $1`

	err := pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database query failed: %w", err)
	}

	return count == 0, nil
}

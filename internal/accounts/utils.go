package accounts

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword хэширует пароль с использованием bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// generateUUID генерирует UUID v4
func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	// Версия 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Вариант 8
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return hex.EncodeToString(uuid), nil
}

// validateName проверяет валидность имени
func validateName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}

	if len(name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}

	// Проверяем, что имя содержит только буквы, пробелы и дефисы
	for _, char := range name {
		if !unicode.IsLetter(char) && char != ' ' && char != '-' && char != '\'' {
			return fmt.Errorf("name contains invalid characters")
		}
	}

	return nil
}

func validateEmail(email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if len(email) < 4 || len(email) >= 254 {
		return false, fmt.Errorf("bad email size")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	if !emailRegex.MatchString(email) {
		return false, fmt.Errorf("bad email format")
	}

	return true, nil
}

func validatePassword(password string) (bool, error) {
	if len(password) < 8 || len(password) > 255 {
		return false, fmt.Errorf("bad password size")
	}

	var hasUpper, hasLower, hasDigit bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return false, fmt.Errorf("bad password complexity")
	}

	return true, nil
}

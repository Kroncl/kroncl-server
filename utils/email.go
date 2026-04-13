package utils

import "strings"

// Вспомогательная функция для валидации email
func IsValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at < 1 || at > len(email)-4 {
		return false
	}

	dot := strings.LastIndex(email[at:], ".")
	if dot < 2 || dot > len(email[at:])-3 {
		return false
	}

	return true
}

// Функция для маскировки ключа
func MaskApiKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	prefix := key[:4]
	suffix := key[len(key)-4:]
	return prefix + "********" + suffix
}

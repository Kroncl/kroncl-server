package companies

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func (s *Service) ValidateCompanyName(name string) error {
	name = strings.TrimSpace(name)

	// Проверка на пустоту
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}

	// Проверка длины в символах (Unicode)
	length := utf8.RuneCountInString(name)
	if length < 2 {
		return fmt.Errorf("company name must be at least 2 characters")
	}
	if length > 100 {
		return fmt.Errorf("company name must be no more than 100 characters")
	}

	// Преобразуем в runes для корректной обработки Unicode
	runes := []rune(name)

	// Проверка каждого символа
	for i, r := range runes {
		// Разрешаем буквы любых языков
		if unicode.IsLetter(r) {
			continue
		}

		// Разрешаем цифры
		if unicode.IsDigit(r) {
			// Нельзя чтобы имя состояло только из цифр
			if length == 1 {
				return fmt.Errorf("company name cannot be only a number")
			}
			continue
		}

		// Разрешаем пробелы (но не в начале/конце)
		if unicode.IsSpace(r) {
			if i == 0 || i == len(runes)-1 {
				return fmt.Errorf("company name cannot start or end with a space")
			}
			// Проверка на множественные пробелы
			if i > 0 && unicode.IsSpace(runes[i-1]) {
				return fmt.Errorf("company name cannot contain multiple consecutive spaces")
			}
			continue
		}

		// Разрешаем некоторые знаки препинания
		allowedPunctuation := []rune{'-', '_', '.', ',', '\'', '&', '(', ')'}
		allowed := false
		for _, p := range allowedPunctuation {
			if r == p {
				allowed = true
				break
			}
		}

		if allowed {
			// Нельзя начинать или заканчивать пунктуацией
			if i == 0 || i == len(runes)-1 {
				return fmt.Errorf("company name cannot start or end with punctuation")
			}
			continue
		}

		// Все остальное - ошибка
		return fmt.Errorf("company name contains invalid character: '%c'", r)
	}

	// Дополнительная проверка: имя не должно состоять только из цифр
	allDigits := true
	for _, r := range runes {
		if !unicode.IsDigit(r) {
			allDigits = false
			break
		}
	}
	if allDigits {
		return fmt.Errorf("company name cannot consist only of numbers")
	}

	return nil
}

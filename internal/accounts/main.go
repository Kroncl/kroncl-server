package accounts

import (
	"fmt"
	"regexp"
	"strings"
)

type Account struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AuthType  string `json:"auth_type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func checkEmailUnique(email string) (bool, error) {
	_, err := validateEmail(email)

	if err != nil {
		return false, err
	}

	return true, nil
}

func validateEmail(email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if len(email) < 4 || len(email) >= 254 {
		return false, fmt.Errorf("bad email size")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	if emailRegex.MatchString(email) != true {
		return false, fmt.Errorf("bad email format")
	}

	return true, nil
}

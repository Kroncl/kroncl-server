package mailer

import "fmt"

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("unisender api error (code %d): %s", e.Code, e.Message)
}

func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

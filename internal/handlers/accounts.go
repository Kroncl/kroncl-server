package handlers

import (
	"encoding/json"
	"net/http"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Просто отправляем ошибку, middleware обернет
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request",
		})
		return
	}

	// Здесь логика регистрации
	// ...

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registration successful",
		"user_id": "generated-id",
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	// Логика входа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": "jwt-token-here",
		"user": map[string]interface{}{
			"id":    "user-id",
			"email": "user@example.com",
			"name":  "John Doe",
		},
	})
}

func ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	// Подтверждение email
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Email confirmed successfully",
	})
}

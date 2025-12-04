package core

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthCheck проверка работоспособности
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Просто пишем в ResponseWriter, middleware обернет
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ErrorResponse хелпер для ошибок
func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

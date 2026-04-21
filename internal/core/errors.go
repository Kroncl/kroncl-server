package core

import (
	"encoding/json"
	"net/http"
)

// SendError отправляет ошибку в формате JSON (только message)
func SendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}

// SendErrorWithData отправляет ошибку с дополнительными данными
func SendErrorWithData(w http.ResponseWriter, statusCode int, message string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"message": message,
	}

	if len(data) > 0 {
		response["data"] = data
	}

	json.NewEncoder(w).Encode(response)
}

// SendValidationError отправляет ошибку валидации
func SendValidationError(w http.ResponseWriter, message string) {
	SendError(w, http.StatusBadRequest, message)
}

// SendUnauthorized отправляет ошибку аутентификации
func SendUnauthorized(w http.ResponseWriter, message string) {
	SendError(w, http.StatusUnauthorized, message)
}

func SendForbidden(w http.ResponseWriter, message string) {
	SendError(w, http.StatusForbidden, message)
}

// SendNotFound отправляет ошибку "не найдено"
func SendNotFound(w http.ResponseWriter, message string) {
	SendError(w, http.StatusNotFound, message)
}

// SendInternalError отправляет внутреннюю ошибку
func SendInternalError(w http.ResponseWriter, message string) {
	SendError(w, http.StatusInternalServerError, message)
}

// SendSuccess отправляет успешный ответ
func SendSuccess(w http.ResponseWriter, data interface{}, message ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := make(map[string]interface{})

	// Добавляем данные
	if data != nil {
		response["data"] = data
	}

	// Добавляем сообщение если есть
	if len(message) > 0 && message[0] != "" {
		response["message"] = message[0]
	}

	json.NewEncoder(w).Encode(response)
}

// SendCreated отправляет ответ на создание
func SendCreated(w http.ResponseWriter, data interface{}, message ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := make(map[string]interface{})

	if data != nil {
		response["data"] = data
	}

	if len(message) > 0 && message[0] != "" {
		response["message"] = message[0]
	}

	json.NewEncoder(w).Encode(response)
}

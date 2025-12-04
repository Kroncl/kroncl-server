package accounts

import (
	"encoding/json"
	"net/http"
)

// Handlers содержит HTTP хендлеры для аккаунтов
type Handlers struct {
	service *Service
}

// NewHandlers создает новый экземпляр хендлеров
func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

// Register обрабатывает запрос на регистрацию
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем запрос
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Вызываем бизнес-логику
	userID, err := h.service.Create(req.Email, req.Name, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	resp := RegisterResponse{
		Message: "Registration successful",
		UserID:  userID,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login обрабатывает запрос на вход
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: реализовать логику входа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": "jwt-token-here",
	})
}

// ConfirmEmail подтверждает email
func (h *Handlers) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	// TODO: реализовать подтверждение email
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email confirmed successfully",
	})
}

package accounts

import (
	"encoding/json"
	"kroncl-server/internal/core"
	"net/http"
)

// Register обрабатывает запрос на регистрацию
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Создаем аккаунт и получаем токены
	account, accessToken, refreshToken, err := h.service.Create(
		r.Context(),
		req.Email,
		req.Name,
		req.Password,
	)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	// Формируем данные для ответа
	data := map[string]interface{}{
		"user_id":       account.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"email_sent":    true,
	}

	// Отправляем ответ
	core.SendCreated(w, data, "Registration successful. Please check your email to confirm your account.")
}

// Login обрабатывает запрос на вход
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Аутентификация
	account, accessToken, refreshToken, err := h.service.Authenticate(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		core.SendUnauthorized(w, err.Error())
		return
	}

	// Формируем данные
	data := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          account,
	}

	core.SendSuccess(w, data, "Login successful")
}

// Refresh обновляет токены по refresh токену
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	if req.RefreshToken == "" {
		core.SendValidationError(w, "Refresh token is required")
		return
	}

	// Обновляем токены
	accessToken, refreshToken, err := h.service.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		core.SendUnauthorized(w, err.Error())
		return
	}

	// Формируем ответ
	data := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	core.SendSuccess(w, data, "Tokens refreshed successfully")
}

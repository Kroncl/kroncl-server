package accounts

import (
	"encoding/json"
	"net/http"

	"matrix-authorization-server/internal/auth"
	"matrix-authorization-server/internal/core"
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

// ConfirmEmail подтверждает email
func (h *Handlers) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Проверяем, что user_id в запросе совпадает с токеном
	if req.UserID != claims.UserID {
		core.SendUnauthorized(w, "User ID mismatch")
		return
	}

	// Подтверждаем email
	err := h.service.ConfirmEmail(req.UserID, req.Code)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	// Ответ с пустыми данными
	core.SendSuccess(w, map[string]interface{}{}, "Email confirmed successfully")
}

// Повторная отправка кода подтверждения
func (h *Handlers) ResendConfirmationCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Повторяем отправку кода
	err := h.service.ResendConfirmationCode(claims.UserID)
	if err != nil {
		// Просто отправляем ошибку как есть, middleware обработает
		core.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Ответ с пустыми данными
	core.SendSuccess(w, map[string]interface{}{}, "Confirmation code has been resent to your email")
}

// GetProfile получает профиль пользователя
func (h *Handlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Получаем аккаунт из БД
	account, err := h.service.GetByID(claims.UserID)
	if err != nil {
		core.SendNotFound(w, "User not found")
		return
	}

	// Отправляем профиль
	core.SendSuccess(w, account, "Profile retrieved successfully")
}

// проверка уникальности почты
func (h *Handlers) CheckEmailUniqie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	ok, err := h.service.checkEmailUnique(req.Email)

	if err != nil {
		core.SendInternalError(w, err.Error())
		return
	}

	if !ok {
		core.SendValidationError(w, "The mail is not unique")
		return
	}

	core.SendSuccess(w, map[string]interface{}{}, "The mail is unique")
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
	accessToken, refreshToken, err := h.service.RefreshTokens(req.RefreshToken)
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

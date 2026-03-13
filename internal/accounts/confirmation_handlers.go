package accounts

import (
	"encoding/json"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"
)

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
	err := h.service.ConfirmEmail(r.Context(), req.UserID, req.Code)
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
	err := h.service.ResendConfirmationCode(r.Context(), claims.UserID)
	if err != nil {
		// Просто отправляем ошибку как есть, middleware обработает
		core.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Ответ с пустыми данными
	core.SendSuccess(w, map[string]interface{}{}, "Confirmation code has been resent to your email")
}

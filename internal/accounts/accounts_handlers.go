package accounts

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"kroncl-server/utils"
	"log"
	"net/http"
	"strings"
)

// обновление данных пользователя (avatar/name)
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Incorrect account data.")
		return
	}

	// Получаем аккаунт из БД
	account, err := h.service.UpdateById(r.Context(), claims.UserID, &req)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("User update error: %s", err.Error()))
		return
	}

	// Отправляем профиль
	core.SendSuccess(w, account, "Profile updated successfully")
}

// GetPublicAccounts возвращает список аккаунтов с пагинацией и поиском
func (h *Handlers) GetPublicAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Извлекаем параметры поиска
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	// Получаем параметры пагинации
	paginationParams := core.GetDefaultPaginationParams(r)

	var accounts []AccountPublic
	var pagination core.Pagination

	accounts, pagination, err := h.service.GetPublicAccounts(
		r.Context(),
		search,
		paginationParams,
	)

	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Error receiving accounts: %s", err.Error()))
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"accounts":   accounts,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Accounts retrieved successfully")
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

	log.Printf("user id: %s", claims.UserID)

	// Получаем аккаунт из БД
	account, err := h.service.GetByID(r.Context(), claims.UserID)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("User not found: %s", err.Error()))
		return
	}

	// Отправляем профиль
	core.SendSuccess(w, account, "Profile retrieved successfully")
}

func (h *Handlers) CheckEmailUnique(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем email из query параметров
	email := r.URL.Query().Get("email")
	if email == "" {
		core.SendValidationError(w, "email parameter is required")
		return
	}

	// Валидация email
	if !utils.IsValidEmail(email) {
		core.SendValidationError(w, "Invalid email format")
		return
	}

	// Проверяем уникальность
	unique, err := h.service.checkEmailUnique(r.Context(), email)
	if err != nil {
		core.SendInternalError(w, err.Error())
		return
	}

	if !unique {
		core.SendValidationError(w, "The mail is not unique")
		return
	}

	core.SendSuccess(w, map[string]interface{}{}, "The mail is unique")
}

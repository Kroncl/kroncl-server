package companies

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// получение одного участника компании
func (h *Handlers) GetCompanyMember(w http.ResponseWriter, r *http.Request) {
	// получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	// получаем ID участника из URL
	memberID := chi.URLParam(r, "accountId")
	if memberID == "" {
		core.SendValidationError(w, "Member ID required.")
		return
	}

	// получаем информацию об участнике
	member, err := h.service.GetCompanyMember(r.Context(), companyID, memberID)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Company member not found: %v", err))
		return
	}

	// отправляем ответ
	core.SendSuccess(w, member, "Company member retrieved successfully.")
}

// получение участников компании с фильтрами (расширенная версия)
func (h *Handlers) GetCompanyMembers(w http.ResponseWriter, r *http.Request) {
	// получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	// парсим параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// создаем запрос с фильтрами
	req := &GetCompanyMembersRequest{
		Page:      pagination.Page,
		Limit:     pagination.Limit,
		Search:    r.URL.Query().Get("search"),
		Role:      r.URL.Query().Get("role"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	// валидация роли
	if req.Role != "" && req.Role != "all" {
		validRoles := map[string]bool{
			"owner":  true,
			"admin":  true,
			"member": true,
			"guest":  true,
		}
		if !validRoles[req.Role] {
			core.SendValidationError(w, "Invalid role. Allowed values: owner, admin, member, guest, all")
			return
		}
	}

	// получаем участников компании с фильтрами
	response, err := h.service.GetCompanyMembers(
		r.Context(),
		companyID,
		req,
	)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get company members: %v", err))
		return
	}

	// отправляем ответ
	core.SendSuccess(w, response, "Company members retrieved successfully.")
}

// получение организации
func (h *Handlers) GetUserCompanyById(w http.ResponseWriter, r *http.Request) {
	// получаем пользователя из контекста
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required.")
		return
	}

	// получаем company_id из URL параметра (не из query!)
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	data, err := h.service.GetUserCompanyById(
		r.Context(),
		account.UserID,
		companyID,
	)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Company not found: %v", err))
		return
	}

	// Отправляем ответ (используйте SendSuccess для GET запроса)
	core.SendSuccess(w, data, "Company retrieved successfully.")
}

// обновление организации
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required.")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Incorrect company data.")
		return
	}

	updatedCompany, err := h.service.UpdateById(r.Context(), companyID, &req)
	if err != nil {
		core.SendValidationError(w, fmt.Sprintf("Company update error: %v", err))
		return
	}

	// Отправляем ответ
	core.SendCreated(w, updatedCompany, "Company updated successful.")
}

// получение организаций пользователя
func (h *Handlers) GetUserCompanies(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Парсим параметры запроса
	query := r.URL.Query()

	// Страница (по умолчанию 1)
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	// Лимит на страницу (по умолчанию 20, максимум 100)
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Роль (owner, guest, admin, member, all)
	role := query.Get("role")
	if role == "" {
		role = "all"
	}

	// Поиск по названию
	search := query.Get("search")

	// Формируем запрос
	req := &GetUserCompaniesRequest{
		Page:   page,
		Limit:  limit,
		Role:   role,
		Search: search,
	}

	// Получаем компании через сервис
	response, err := h.service.GetUserCompanies(r.Context(), account.UserID, req)
	if err != nil {
		// Проверяем тип ошибки для соответствующего HTTP статуса
		if err.Error() == "invalid role filter. Allowed values: all, owner, admin, member, guest" {
			core.SendValidationError(w, err.Error())
		} else {
			core.SendInternalError(w, err.Error())
		}
		return
	}

	// Отправляем успешный ответ
	core.SendSuccess(w, response, "User companies retrieved successfully")
}

// создание организации
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Получаем пользователя из контекста
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Создаем аккаунт и получаем токены
	data, err := h.service.Create(
		r.Context(),
		account.UserID,
		req.Slug,
		req.Name,
		req.Description,
		req.AvatarUrl,
		req.IsPublic,
		req.PlanCode,
	)
	if err != nil {
		core.SendValidationError(w, err.Error())
		return
	}

	// Отправляем ответ
	core.SendCreated(w, data, "Company created successful.")
}

// проверка уникальности slug компании
func (h *Handlers) CheckSlugUnique(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		core.SendValidationError(w, "slug parameter is required")
		return
	}

	ok, err := h.service.checkSlugUnique(r.Context(), slug)
	if err != nil {
		core.SendInternalError(w, err.Error())
		return
	}

	if !ok {
		core.SendValidationError(w, "The slug is not unique")
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"slug":   slug,
		"unique": true,
	}, "The slug is unique")
}

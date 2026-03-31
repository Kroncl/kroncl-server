package companies

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetCompanyPlan возвращает текущий план компании
func (h *Handlers) GetCompanyPricingPlan(w http.ResponseWriter, r *http.Request) {
	// Получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Проверяем принадлежность к компании
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Проверяем, что пользователь состоит в компании
	member, err := h.service.GetUserCompanyById(r.Context(), account.UserID, companyID)
	if err != nil || member == nil {
		core.SendNotFound(w, "Company not found or access denied")
		return
	}

	// Получаем план компании
	plan, err := h.service.GetCompanyPlan(r.Context(), companyID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get company plan: %v", err))
		return
	}

	core.SendSuccess(w, plan, "Company plan retrieved successfully")
}

// GetCompanyTransactions возвращает историю транзакций компании
func (h *Handlers) GetCompanyPricingTransactions(w http.ResponseWriter, r *http.Request) {
	// Получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Проверяем принадлежность к компании
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	member, err := h.service.GetUserCompanyById(r.Context(), account.UserID, companyID)
	if err != nil || member == nil {
		core.SendNotFound(w, "Company not found or access denied")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Получаем транзакции
	transactions, total, err := h.service.GetCompanyTransactions(
		r.Context(),
		companyID,
		pagination.Page,
		pagination.Limit,
	)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get company transactions: %v", err))
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Company transactions retrieved successfully")
}

// CreateCompanyTransaction создает новую транзакцию для смены плана
func (h *Handlers) MigratePricingPlan(w http.ResponseWriter, r *http.Request) {
	// Получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Получаем пользователя из контекста
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Проверяем, что пользователь состоит в компании
	member, err := h.service.GetUserCompanyById(r.Context(), account.UserID, companyID)
	if err != nil || member == nil {
		core.SendNotFound(w, "Company not found or access denied")
		return
	}

	// Парсим тело запроса
	var req MigratePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Валидация
	if req.PlanCode == "" {
		core.SendValidationError(w, "plan_code is required")
		return
	}
	if req.Period != "month" && req.Period != "year" {
		core.SendValidationError(w, "period must be 'month' or 'year'")
		return
	}

	// Создаем транзакцию
	tx, err := h.service.CreateNewTransaction(
		r.Context(),
		companyID,
		account.UserID,
		&req,
	)
	if err != nil {
		core.SendValidationError(w, fmt.Sprintf("Failed to create transaction: %v", err))
		return
	}

	core.SendCreated(w, tx, "Transaction created successfully")
}

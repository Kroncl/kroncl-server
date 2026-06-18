package companies

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/pricing"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RevokePricingTransaction отменяет транзакцию
func (h *Handlers) RevokePricingTransaction(w http.ResponseWriter, r *http.Request) {
	// Получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Получаем ID транзакции из URL
	transactionID := chi.URLParam(r, "transactionId")
	if transactionID == "" {
		core.SendValidationError(w, "Transaction ID required")
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

	// Отменяем транзакцию
	err = h.service.RevokeTransaction(r.Context(), companyID, transactionID)
	if err != nil {
		core.SendValidationError(w, fmt.Sprintf("Failed to revoke transaction: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Transaction revoked successfully")
}

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

// MigratePricingPlan создает транзакцию и инициирует платеж
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

	// Формируем successURL
	var successURL string
	if req.SuccessURL != "" {
		successURL = req.SuccessURL
	} else {
		clientDomain := config.GetClientDomain()
		successURL = fmt.Sprintf("%s/platform/%s/pricing/success", clientDomain, companyID)
	}

	// Создаем транзакцию и инициируем платеж
	result, err := h.service.InitPayment(
		r.Context(),
		companyID,
		account.UserID,
		&req,
		successURL,
	)
	if err != nil {
		core.SendValidationError(w, fmt.Sprintf("Failed to init payment: %v", err))
		return
	}

	// Возвращаем транзакцию и URL для оплаты
	response := struct {
		Transaction    *pricing.PricingTransaction `json:"transaction"`
		PaymentPageURL string                      `json:"payment_page_url"`
		PaymentID      string                      `json:"payment_id"`
	}{
		Transaction:    result.Transaction,
		PaymentPageURL: result.PaymentPageURL,
		PaymentID:      result.PaymentID,
	}

	core.SendCreated(w, response, "Transaction created successfully")
}

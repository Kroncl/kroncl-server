package fm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/core"
	"net/http"
	"strings"
	"time"
)

type Handlers struct {
	repository *Repository
}

func NewHandlers(repository *Repository) *Handlers {
	return &Handlers{repository: repository}
}

// --------
// TRANSACTIONS
// --------

func (h *Handlers) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Парсим тело запроса
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация обязательных полей
	if req.BaseAmount <= 0 {
		core.SendError(w, http.StatusBadRequest, "Amount must be greater than 0.")
		return
	}
	if req.EmployeeID == "" {
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}
	if req.Direction == "" {
		core.SendError(w, http.StatusBadRequest, "Transaction direction is required.")
		return
	}
	if req.Currency == "" {
		core.SendError(w, http.StatusBadRequest, "Currency is required.")
		return
	}

	transaction, err := h.repository.CreateTransaction(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "invalid employee_id"):
			core.SendNotFound(w, "Employee not found.")
		case strings.Contains(errorMsg, "invalid transaction direction"):
			core.SendValidationError(w, "Invalid transaction direction. Use 'income' or 'expense'.")
		case strings.Contains(errorMsg, "invalid currency"):
			core.SendValidationError(w, "Invalid currency. Supported: RUB, USD, EUR, KZT.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to create transaction: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, transaction, "Transaction created successfully.")
}

func (h *Handlers) GetTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID транзакции из URL
	transactionID := r.PathValue("transactionId")
	if transactionID == "" {
		core.SendError(w, http.StatusBadRequest, "Transaction ID is required.")
		return
	}

	transaction, err := h.repository.GetTransactionByID(r.Context(), transactionID)
	if err != nil {
		core.SendNotFound(w, "Transaction not found.")
		return
	}

	core.SendSuccess(w, transaction, "Transaction retrieved successfully.")
}

func (h *Handlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Парсим query параметры в структуру фильтров
	var filters GetTransactionsRequest

	// Date filters
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			filters.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			filters.EndDate = &t
		}
	}

	// Direction filter
	if dir := r.URL.Query().Get("direction"); dir != "" {
		d := TransactionDirection(dir)
		if d != TransactionDirectionIncome && d != TransactionDirectionExpense {
			core.SendValidationError(w, "Invalid direction. Use 'income' or 'expense'.")
			return
		}
		filters.Direction = &d
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		s := TransactionStatus(status)
		switch s {
		case TransactionStatusPending, TransactionStatusCompleted, TransactionStatusFailed, TransactionStatusCancelled:
			filters.Status = &s
		default:
			core.SendValidationError(w, "Invalid status. Use: pending, completed, failed, cancelled.")
			return
		}
	}

	// Category filter
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filters.CategoryID = &categoryID
	}

	// Employee filter
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		filters.EmployeeID = &employeeID
	}

	// Search filter
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	transactions, total, err := h.repository.GetTransactions(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
		filters,
	)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get transactions: %s", err.Error()))
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"pagination": core.NewPagination(
			int(total),
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Transactions retrieved successfully.")
}

func (h *Handlers) CreateReverseTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID транзакции из URL
	transactionID := r.PathValue("transactionId")
	if transactionID == "" {
		core.SendError(w, http.StatusBadRequest, "Transaction ID is required.")
		return
	}

	transaction, err := h.repository.CreateReverseTransaction(r.Context(), transactionID)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("Failed create reverse transaction: %s", err.Error()))
		return
	}

	core.SendSuccess(w, transaction, "Reverse transaction created successfully.")
}

// --------
// CATEGORIES
// базовый круд без хуйни
// --------

func (h *Handlers) GetCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	category, err := h.repository.GetCategoryByID(r.Context(), categoryID)
	if err != nil {
		core.SendNotFound(w, "Category not found.")
		return
	}

	core.SendSuccess(w, category, "Category retrieved successfully.")
}

func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Фильтры
	search := r.URL.Query().Get("search")

	var direction *TransactionCategoryDirection
	if dir := r.URL.Query().Get("direction"); dir != "" {
		d := TransactionCategoryDirection(dir)
		if d != TransactionCategoryDirectionIncome && d != TransactionCategoryDirectionExpense {
			core.SendValidationError(w, "Invalid direction. Use 'income' or 'expense'.")
			return
		}
		direction = &d
	}

	categories, total, err := h.repository.GetCategories(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
		direction,
		search,
	)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get categories: %s", err.Error()))
		return
	}

	response := map[string]interface{}{
		"categories": categories,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Categories retrieved successfully.")
}

func (h *Handlers) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Парсим тело запроса
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация
	if strings.TrimSpace(req.Name) == "" {
		core.SendError(w, http.StatusBadRequest, "Category name is required.")
		return
	}
	if req.Direction == "" {
		core.SendError(w, http.StatusBadRequest, "Category direction is required.")
		return
	}
	if req.Direction != TransactionCategoryDirectionIncome && req.Direction != TransactionCategoryDirectionExpense {
		core.SendValidationError(w, "Invalid direction. Use 'income' or 'expense'.")
		return
	}

	category, err := h.repository.CreateCategory(r.Context(), req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to create category: %s", err.Error()))
		return
	}

	core.SendSuccess(w, category, "Category created successfully.")
}

func (h *Handlers) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация направления если указано
	if req.Direction != nil {
		if *req.Direction != TransactionCategoryDirectionIncome && *req.Direction != TransactionCategoryDirectionExpense {
			core.SendValidationError(w, "Invalid direction. Use 'income' or 'expense'.")
			return
		}
	}

	category, err := h.repository.UpdateCategory(r.Context(), categoryID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "category not found"):
			core.SendNotFound(w, "Category not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to update category: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, category, "Category updated successfully.")
}

func (h *Handlers) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	ok, err := h.repository.DeleteCategory(r.Context(), categoryID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "cannot delete category: used in"):
			core.SendValidationError(w, errorMsg)
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to delete category: %s", errorMsg))
		}
		return
	}

	if !ok {
		core.SendNotFound(w, "Category not found.")
		return
	}

	core.SendSuccess(w, map[string]interface{}{
		"category_id": categoryID,
		"deleted":     true,
	}, "Category deleted successfully.")
}

// --------
// ANALYSIS
// --------

func (h *Handlers) GetAnalysisSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Парсим параметры дат
	var startDate, endDate *time.Time

	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			startDate = &t
		}
	}

	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			endDate = &t
		}
	}

	summary, err := h.repository.GetAnalysisSummary(r.Context(), startDate, endDate)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get analysis summary: %s", err.Error()))
		return
	}

	core.SendSuccess(w, summary, "Analysis summary retrieved successfully.")
}

// func (h *Handlers) GetDailyStats(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
// 		return
// 	}

// 	var startDate, endDate *time.Time

// 	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
// 		t, err := time.Parse(time.RFC3339, startStr)
// 		if err == nil {
// 			startDate = &t
// 		}
// 	}

// 	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
// 		t, err := time.Parse(time.RFC3339, endStr)
// 		if err == nil {
// 			endDate = &t
// 		}
// 	}

// 	stats, err := h.repository.GetDailyStats(r.Context(), startDate, endDate)
// 	if err != nil {
// 		core.SendInternalError(w, fmt.Sprintf("Failed to get daily stats: %s", err.Error()))
// 		return
// 	}

// 	core.SendSuccess(w, stats, "Daily stats retrieved successfully.")
// }

func (h *Handlers) GetGroupedTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	groupBy := GroupBy(r.URL.Query().Get("group_by"))
	if groupBy == "" {
		core.SendValidationError(w, "group_by parameter is required (category/employee/day/month)")
		return
	}

	var startDate, endDate *time.Time
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		t, _ := time.Parse(time.RFC3339, startStr)
		startDate = &t
	}
	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		t, _ := time.Parse(time.RFC3339, endStr)
		endDate = &t
	}

	stats, err := h.repository.GetGroupedTransactions(r.Context(), groupBy, startDate, endDate)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get grouped stats: %s", err.Error()))
		return
	}

	core.SendSuccess(w, stats, "Grouped stats retrieved successfully.")
}

// --------
// COUNTERPARTIES
// --------

func (h *Handlers) GetCounterparty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	counterparty, err := h.repository.GetCounterpartyByID(r.Context(), counterpartyID)
	if err != nil {
		core.SendNotFound(w, "Counterparty not found.")
		return
	}

	core.SendSuccess(w, counterparty, "Counterparty retrieved successfully.")
}

func (h *Handlers) GetCounterparties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Парсим query параметры в структуру фильтров
	var filters GetCounterpartiesRequest
	filters.Page = pagination.Page
	filters.Limit = pagination.Limit

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := CounterpartyType(typeStr)
		if t != CounterpartyTypeBank && t != CounterpartyTypeOrganization && t != CounterpartyTypePerson {
			core.SendValidationError(w, "Invalid type. Use 'bank', 'organization' or 'person'.")
			return
		}
		filters.Type = &t
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := CounterpartyStatus(statusStr)
		if s != CounterpartyStatusActive && s != CounterpartyStatusInactive {
			core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
			return
		}
		filters.Status = &s
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	counterparties, total, err := h.repository.GetCounterparties(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
		filters,
	)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get counterparties: %s", err.Error()))
		return
	}

	response := map[string]interface{}{
		"counterparties": counterparties,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Counterparties retrieved successfully.")
}

func (h *Handlers) CreateCounterparty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Парсим тело запроса
	var req CreateCounterpartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация
	if strings.TrimSpace(req.Name) == "" {
		core.SendError(w, http.StatusBadRequest, "Counterparty name is required.")
		return
	}
	if req.Type == "" {
		core.SendError(w, http.StatusBadRequest, "Counterparty type is required.")
		return
	}
	if req.Type != CounterpartyTypeBank && req.Type != CounterpartyTypeOrganization && req.Type != CounterpartyTypePerson {
		core.SendValidationError(w, "Invalid type. Use 'bank', 'organization' or 'person'.")
		return
	}
	if req.Status != "" && req.Status != CounterpartyStatusActive && req.Status != CounterpartyStatusInactive {
		core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
		return
	}

	counterparty, err := h.repository.CreateCounterparty(r.Context(), req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to create counterparty: %s", err.Error()))
		return
	}

	core.SendSuccess(w, counterparty, "Counterparty created successfully.")
}

func (h *Handlers) UpdateCounterparty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateCounterpartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация типа если указан
	if req.Type != nil {
		if *req.Type != CounterpartyTypeBank && *req.Type != CounterpartyTypeOrganization && *req.Type != CounterpartyTypePerson {
			core.SendValidationError(w, "Invalid type. Use 'bank', 'organization' or 'person'.")
			return
		}
	}

	counterparty, err := h.repository.UpdateCounterparty(r.Context(), counterpartyID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			core.SendNotFound(w, "Counterparty not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to update counterparty: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, counterparty, "Counterparty updated successfully.")
}

func (h *Handlers) ActivateCounterparty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	counterparty, err := h.repository.ActivateCounterparty(r.Context(), counterpartyID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			core.SendNotFound(w, "Counterparty not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to activate counterparty: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, counterparty, "Counterparty activated successfully.")
}

func (h *Handlers) DeactivateCounterparty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	counterparty, err := h.repository.DeactivateCounterparty(r.Context(), counterpartyID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			core.SendNotFound(w, "Counterparty not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate counterparty: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, counterparty, "Counterparty deactivated successfully.")
}

// --------
// CREDITS
// --------

func (h *Handlers) GetCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	credit, err := h.repository.GetCreditByID(r.Context(), creditID)
	if err != nil {
		core.SendNotFound(w, "Credit not found.")
		return
	}

	core.SendSuccess(w, credit, "Credit retrieved successfully.")
}

func (h *Handlers) GetCredits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Парсим query параметры в структуру фильтров
	var filters GetCreditsRequest
	filters.Page = pagination.Page
	filters.Limit = pagination.Limit

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := CreditType(typeStr)
		if t != CreditTypeDebt && t != CreditTypeCredit {
			core.SendValidationError(w, "Invalid type. Use 'debt' or 'credit'.")
			return
		}
		filters.Type = &t
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := CreditStatus(statusStr)
		if s != CreditStatusActive && s != CreditStatusClosed {
			core.SendValidationError(w, "Invalid status. Use 'active' or 'closed'.")
			return
		}
		filters.Status = &s
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	credits, total, err := h.repository.GetCredits(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
		filters,
	)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get credits: %s", err.Error()))
		return
	}

	response := map[string]interface{}{
		"credits": credits,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Credits retrieved successfully.")
}

func (h *Handlers) CreateCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Парсим тело запроса
	var req CreateCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация
	if strings.TrimSpace(req.Name) == "" {
		core.SendError(w, http.StatusBadRequest, "Credit name is required.")
		return
	}
	if req.Type == "" {
		core.SendError(w, http.StatusBadRequest, "Credit type is required.")
		return
	}
	if req.Type != CreditTypeDebt && req.Type != CreditTypeCredit {
		core.SendValidationError(w, "Invalid type. Use 'debt' or 'credit'.")
		return
	}
	if req.TotalAmount <= 0 {
		core.SendError(w, http.StatusBadRequest, "Total amount must be greater than 0.")
		return
	}
	if req.Currency == "" {
		core.SendError(w, http.StatusBadRequest, "Currency is required.")
		return
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		core.SendError(w, http.StatusBadRequest, "Start date and end date are required.")
		return
	}
	if req.EndDate.Before(req.StartDate) {
		core.SendError(w, http.StatusBadRequest, "End date must be after start date.")
		return
	}
	if req.CounterpartyID == "" {
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}
	if req.InterestRate < 0 || req.InterestRate > 100 {
		core.SendError(w, http.StatusBadRequest, "Interest rate must be between 0 and 100.")
		return
	}

	credit, err := h.repository.CreateCredit(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			core.SendNotFound(w, "Counterparty not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to create credit: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, credit, "Credit created successfully.")
}

func (h *Handlers) UpdateCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация типа если указан
	if req.Type != nil {
		if *req.Type != CreditTypeDebt && *req.Type != CreditTypeCredit {
			core.SendValidationError(w, "Invalid type. Use 'debt' or 'credit'.")
			return
		}
	}

	// Валидация дат если обе указаны
	if req.StartDate != nil && req.EndDate != nil {
		if req.EndDate.Before(*req.StartDate) {
			core.SendError(w, http.StatusBadRequest, "End date must be after start date.")
			return
		}
	}

	// Валидация процентной ставки если указана
	if req.InterestRate != nil {
		if *req.InterestRate < 0 || *req.InterestRate > 100 {
			core.SendError(w, http.StatusBadRequest, "Interest rate must be between 0 and 100.")
			return
		}
	}

	credit, err := h.repository.UpdateCredit(r.Context(), creditID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			core.SendNotFound(w, "Credit not found.")
		case strings.Contains(errorMsg, "counterparty not found"):
			core.SendNotFound(w, "Counterparty not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to update credit: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, credit, "Credit updated successfully.")
}

func (h *Handlers) ActivateCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	credit, err := h.repository.ActivateCredit(r.Context(), creditID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			core.SendNotFound(w, "Credit not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to activate credit: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, credit, "Credit activated successfully.")
}

func (h *Handlers) DeactivateCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	credit, err := h.repository.DeactivateCredit(r.Context(), creditID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			core.SendNotFound(w, "Credit not found.")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate credit: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, credit, "Credit deactivated successfully.")
}

// --------
// CREDIT PAYMENTS
// --------

func (h *Handlers) PayCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	// Парсим тело запроса
	var req PayCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	req.CreditID = creditID

	// Валидация
	if req.Amount <= 0 {
		core.SendError(w, http.StatusBadRequest, "Payment amount must be greater than 0.")
		return
	}
	if req.EmployeeID == "" {
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}
	if req.PaidAt.IsZero() {
		core.SendError(w, http.StatusBadRequest, "Payment date is required.")
		return
	}

	transaction, err := h.repository.PayCredit(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			core.SendNotFound(w, "Credit not found.")
		case strings.Contains(errorMsg, "credit is not active"):
			core.SendValidationError(w, "Credit is not active.")
		case strings.Contains(errorMsg, "exceeds remaining debt"):
			core.SendValidationError(w, errorMsg)
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to process payment: %s", errorMsg))
		}
		return
	}

	core.SendSuccess(w, transaction, "Payment processed successfully.")
}

func (h *Handlers) GetCreditTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	transactions, total, err := h.repository.GetCreditTransactions(
		r.Context(),
		creditID,
		pagination.Offset,
		pagination.Limit,
	)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get credit transactions: %s", err.Error()))
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"pagination": core.NewPagination(
			int(total),
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Credit transactions retrieved successfully.")
}

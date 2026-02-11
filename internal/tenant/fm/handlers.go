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

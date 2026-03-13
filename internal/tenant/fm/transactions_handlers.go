package fm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"
	"time"
)

// --------
// TRANSACTIONS
// --------

func (h *Handlers) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим тело запроса
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация обязательных полей
	if req.BaseAmount <= 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Amount must be greater than 0"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Amount must be greater than 0.")
		return
	}
	if req.EmployeeID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}
	if req.Direction == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Transaction direction is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Transaction direction is required.")
		return
	}
	if req.Currency == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Currency is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Currency is required.")
		return
	}

	transaction, err := h.repository.CreateTransaction(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "invalid employee_id"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("employee_id", req.EmployeeID),
			)
			core.SendNotFound(w, "Employee not found.")
		case strings.Contains(errorMsg, "invalid transaction direction"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("direction", req.Direction),
			)
			core.SendValidationError(w, "Invalid transaction direction. Use 'income' or 'expense'.")
		case strings.Contains(errorMsg, "invalid currency"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("currency", req.Currency),
			)
			core.SendValidationError(w, "Invalid currency. Supported: RUB, USD, EUR, KZT.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create transaction: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("transaction_id", transaction.ID),
		logs.WithMetadata("amount", req.BaseAmount),
		logs.WithMetadata("direction", req.Direction),
		logs.WithMetadata("currency", req.Currency),
		logs.WithMetadata("employee_id", req.EmployeeID),
	)

	core.SendSuccess(w, transaction, "Transaction created successfully.")
}

func (h *Handlers) GetTransaction(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID транзакции из URL
	transactionID := r.PathValue("transactionId")
	if transactionID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Transaction ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Transaction ID is required.")
		return
	}

	transaction, err := h.repository.GetTransactionByID(r.Context(), transactionID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Transaction not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("transaction_id", transactionID),
		)
		core.SendNotFound(w, "Transaction not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("transaction_id", transactionID),
	)

	core.SendSuccess(w, transaction, "Transaction retrieved successfully.")
}

func (h *Handlers) GetTransactions(w http.ResponseWriter, r *http.Request) {
	// Получаем accountID из контекста (после JWT middleware)
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
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
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid direction"),
				logs.WithMetadata("direction", dir),
				logs.WithMetadata("path", r.URL.Path),
			)
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
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid status"),
				logs.WithMetadata("status", status),
				logs.WithMetadata("path", r.URL.Path),
			)
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
		// Логируем ошибку
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get transactions: %s", err.Error()))
		return
	}

	// Логируем успешный просмотр
	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"start_date":  filters.StartDate,
			"end_date":    filters.EndDate,
			"direction":   filters.Direction,
			"status":      filters.Status,
			"category_id": filters.CategoryID,
			"employee_id": filters.EmployeeID,
			"search":      filters.Search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(transactions)),
	)

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
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID транзакции из URL
	transactionID := r.PathValue("transactionId")
	if transactionID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_REVERSE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Transaction ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Transaction ID is required.")
		return
	}

	transaction, err := h.repository.CreateReverseTransaction(r.Context(), transactionID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_REVERSE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("transaction_id", transactionID),
		)
		core.SendNotFound(w, fmt.Sprintf("Failed create reverse transaction: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_REVERSE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("original_transaction_id", transactionID),
		logs.WithMetadata("reverse_transaction_id", transaction.ID),
	)

	core.SendSuccess(w, transaction, "Reverse transaction created successfully.")
}

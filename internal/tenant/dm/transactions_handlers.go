package dm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// GetDealTransactions возвращает список транзакций сделки
func (h *Handlers) GetDealTransactions(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dealID := chi.URLParam(r, "dealId")
	if dealID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required.")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	var filters fm.GetTransactionsRequest

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
	if dir := r.URL.Query().Get("direction"); dir != "" {
		d := fm.TransactionDirection(dir)
		if d == fm.TransactionDirectionIncome || d == fm.TransactionDirectionExpense {
			filters.Direction = &d
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		s := fm.TransactionStatus(status)
		filters.Status = &s
	}
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filters.CategoryID = &categoryID
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		filters.EmployeeID = &employeeID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	transactions, total, err := h.repository.fmRepository.GetDealTransactions(
		r.Context(),
		dealID,
		pagination.Offset,
		pagination.Limit,
		filters,
	)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("deal_id", dealID),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deal transactions: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("deal_id", dealID),
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

	core.SendSuccess(w, response, "Deal transactions retrieved successfully.")
}

// GetDealTransactionsSummary возвращает сводку по транзакциям сделки
func (h *Handlers) GetDealTransactionsSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dealID := chi.URLParam(r, "dealId")
	if dealID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_SUMMARY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required.")
		return
	}

	var filters fm.GetTransactionsRequest

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
	if status := r.URL.Query().Get("status"); status != "" {
		s := fm.TransactionStatus(status)
		filters.Status = &s
	}
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filters.CategoryID = &categoryID
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		filters.EmployeeID = &employeeID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	summary, err := h.repository.fmRepository.GetDealTransactionsSummary(r.Context(), dealID, filters)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_SUMMARY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("deal_id", dealID),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deal transactions summary: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_SUMMARY, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("deal_id", dealID),
	)

	core.SendSuccess(w, summary, "Deal transactions summary retrieved successfully.")
}

// CreateDealTransaction создает транзакцию для сделки
func (h *Handlers) CreateDealTransaction(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dealID := chi.URLParam(r, "dealId")
	if dealID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required.")
		return
	}

	var req fm.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if req.BaseAmount <= 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Amount must be greater than 0"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Amount must be greater than 0.")
		return
	}

	if req.Direction == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Transaction direction is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Transaction direction is required.")
		return
	}

	if req.Currency == "" {
		req.Currency = fm.CurrencyRUB
	}

	transaction, err := h.repository.fmRepository.CreateDealTransaction(r.Context(), dealID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "deal not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("deal_id", dealID),
			)
			core.SendNotFound(w, "Deal not found.")
		case strings.Contains(errorMsg, "invalid employee_id"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("employee_id", req.EmployeeID),
			)
			core.SendNotFound(w, "Employee not found.")
		case strings.Contains(errorMsg, "invalid transaction direction"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("direction", req.Direction),
			)
			core.SendValidationError(w, "Invalid transaction direction. Use 'income' or 'expense'.")
		case strings.Contains(errorMsg, "invalid currency"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("currency", req.Currency),
			)
			core.SendValidationError(w, "Invalid currency. Supported: RUB.")
		case strings.Contains(errorMsg, "deal category not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, "Deal categories not configured. Please create 'deal-income' and 'deal-expense' categories.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create deal transaction: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_TRANSACTIONS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("deal_id", dealID),
		logs.WithMetadata("transaction_id", transaction.ID),
		logs.WithMetadata("amount", req.BaseAmount),
		logs.WithMetadata("direction", req.Direction),
	)

	core.SendSuccess(w, transaction, "Deal transaction created successfully.")
}

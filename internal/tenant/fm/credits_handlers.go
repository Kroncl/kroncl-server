package fm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"
)

// --------
// CREDITS
// --------

func (h *Handlers) GetCredit(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	credit, err := h.repository.GetCreditByID(r.Context(), creditID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("credit_id", creditID),
		)
		core.SendNotFound(w, "Credit not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("credit_id", creditID),
	)

	core.SendSuccess(w, credit, "Credit retrieved successfully.")
}

func (h *Handlers) GetCredits(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
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
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid type"),
				logs.WithMetadata("type", typeStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid type. Use 'debt' or 'credit'.")
			return
		}
		filters.Type = &t
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := CreditStatus(statusStr)
		if s != CreditStatusActive && s != CreditStatusClosed {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid status"),
				logs.WithMetadata("status", statusStr),
				logs.WithMetadata("path", r.URL.Path),
			)
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
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get credits: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"type":   filters.Type,
			"status": filters.Status,
			"search": filters.Search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(credits)),
	)

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
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим тело запроса
	var req CreateCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация
	if strings.TrimSpace(req.Name) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Credit name is required.")
		return
	}
	if req.Type == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit type is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Credit type is required.")
		return
	}
	if req.Type != CreditTypeDebt && req.Type != CreditTypeCredit {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid type"),
			logs.WithMetadata("type", req.Type),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid type. Use 'debt' or 'credit'.")
		return
	}
	if req.TotalAmount <= 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Total amount must be greater than 0"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Total amount must be greater than 0.")
		return
	}
	if req.Currency == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Currency is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Currency is required.")
		return
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Start date and end date are required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Start date and end date are required.")
		return
	}
	if req.EndDate.Before(req.StartDate) {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "End date must be after start date"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "End date must be after start date.")
		return
	}
	if req.CounterpartyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}
	if req.InterestRate < 0 || req.InterestRate > 100 {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Interest rate must be between 0 and 100"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Interest rate must be between 0 and 100.")
		return
	}

	credit, err := h.repository.CreateCredit(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Counterparty not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("counterparty_id", req.CounterpartyID),
			)
			core.SendNotFound(w, "Counterparty not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create credit: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("credit_id", credit.ID),
		logs.WithMetadata("name", req.Name),
		logs.WithMetadata("type", req.Type),
		logs.WithMetadata("total_amount", req.TotalAmount),
		logs.WithMetadata("currency", req.Currency),
	)

	core.SendSuccess(w, credit, "Credit created successfully.")
}

func (h *Handlers) UpdateCredit(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация типа если указан
	if req.Type != nil {
		if *req.Type != CreditTypeDebt && *req.Type != CreditTypeCredit {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid type"),
				logs.WithMetadata("type", *req.Type),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid type. Use 'debt' or 'credit'.")
			return
		}
	}

	// Валидация дат если обе указаны
	if req.StartDate != nil && req.EndDate != nil {
		if req.EndDate.Before(*req.StartDate) {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "End date must be after start date"),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendError(w, http.StatusBadRequest, "End date must be after start date.")
			return
		}
	}

	// Валидация процентной ставки если указана
	if req.InterestRate != nil {
		if *req.InterestRate < 0 || *req.InterestRate > 100 {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Interest rate must be between 0 and 100"),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendError(w, http.StatusBadRequest, "Interest rate must be between 0 and 100.")
			return
		}
	}

	credit, err := h.repository.UpdateCredit(r.Context(), creditID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Credit not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("credit_id", creditID),
			)
			core.SendNotFound(w, "Credit not found.")
		case strings.Contains(errorMsg, "counterparty not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Counterparty not found"),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendNotFound(w, "Counterparty not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update credit: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("credit_id", creditID),
	)

	core.SendSuccess(w, credit, "Credit updated successfully.")
}

func (h *Handlers) ActivateCredit(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	credit, err := h.repository.ActivateCredit(r.Context(), creditID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Credit not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("credit_id", creditID),
			)
			core.SendNotFound(w, "Credit not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to activate credit: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("credit_id", creditID),
		logs.WithMetadata("action", "activate"),
	)

	core.SendSuccess(w, credit, "Credit activated successfully.")
}

func (h *Handlers) DeactivateCredit(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	credit, err := h.repository.DeactivateCredit(r.Context(), creditID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Credit not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("credit_id", creditID),
			)
			core.SendNotFound(w, "Credit not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate credit: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("credit_id", creditID),
		logs.WithMetadata("action", "deactivate"),
	)

	core.SendSuccess(w, credit, "Credit deactivated successfully.")
}

// --------
// CREDIT PAYMENTS
// --------

func (h *Handlers) PayCredit(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Credit ID is required.")
		return
	}

	// Парсим тело запроса
	var req PayCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	req.CreditID = creditID

	// Валидация
	if req.Amount <= 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Payment amount must be greater than 0"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Payment amount must be greater than 0.")
		return
	}
	if req.EmployeeID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Employee ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Employee ID is required.")
		return
	}
	if req.PaidAt.IsZero() {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Payment date is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Payment date is required.")
		return
	}

	transaction, err := h.repository.PayCredit(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "credit not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Credit not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("credit_id", creditID),
			)
			core.SendNotFound(w, "Credit not found.")
		case strings.Contains(errorMsg, "credit is not active"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Credit is not active"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("credit_id", creditID),
			)
			core.SendValidationError(w, "Credit is not active.")
		case strings.Contains(errorMsg, "exceeds remaining debt"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("credit_id", creditID),
				logs.WithMetadata("amount", req.Amount),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to process payment: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_PAY, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("credit_id", creditID),
		logs.WithMetadata("transaction_id", transaction.ID),
		logs.WithMetadata("amount", req.Amount),
		logs.WithMetadata("employee_id", req.EmployeeID),
	)

	core.SendSuccess(w, transaction, "Payment processed successfully.")
}

func (h *Handlers) GetCreditTransactions(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID кредита из URL
	creditID := r.PathValue("creditId")
	if creditID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_TRANSACTIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Credit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
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
		h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_TRANSACTIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get credit transactions: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_CREDITS_TRANSACTIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("credit_id", creditID),
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

	core.SendSuccess(w, response, "Credit transactions retrieved successfully.")
}

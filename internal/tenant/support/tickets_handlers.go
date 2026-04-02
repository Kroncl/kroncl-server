package support

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetTickets возвращает список тикетов компании
func (h *Handlers) GetTickets(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Company ID required")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	tickets, total, err := h.service.GetTickets(r.Context(), companyID, pagination.Page, pagination.Limit)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get tickets: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("result_count", len(tickets)),
	)

	response := map[string]interface{}{
		"tickets": tickets,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Tickets retrieved successfully")
}

// GetTicketByID возвращает один тикет
func (h *Handlers) GetTicketByID(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Company ID required")
		return
	}

	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Ticket ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Ticket ID required")
		return
	}

	// Проверяем доступ
	if err := h.service.CheckTicketAccess(r.Context(), companyID, ticketID); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendNotFound(w, "Ticket not found")
		return
	}

	ticket, err := h.service.GetTicketByID(r.Context(), companyID, ticketID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendNotFound(w, "Ticket not found")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("ticket_id", ticketID),
	)

	core.SendSuccess(w, ticket, "Ticket retrieved successfully")
}

// UpdateTicketStatus обновляет статус тикета
func (h *Handlers) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Company ID required")
		return
	}

	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Ticket ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Ticket ID required")
		return
	}

	// Проверяем доступ
	if err := h.service.CheckTicketAccess(r.Context(), companyID, ticketID); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendNotFound(w, "Ticket not found")
		return
	}

	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid request body")
		return
	}

	if req.Status == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Status is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Status is required")
		return
	}

	ticket, err := h.service.UpdateTicketStatus(r.Context(), companyID, ticketID, req.Status)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendValidationError(w, err.Error())
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("ticket_id", ticketID),
		logs.WithMetadata("new_status", string(req.Status)),
	)

	core.SendSuccess(w, ticket, "Ticket status updated successfully")
}

// CreateTicket создаёт новый тикет
func (h *Handlers) CreateTicket(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Company ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Company ID required")
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid request body")
		return
	}

	// Валидация
	if req.Theme == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Theme is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Theme is required")
		return
	}

	if req.Text == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Text is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Text is required")
		return
	}

	if len(req.Text) < 10 || len(req.Text) > 3000 {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Text must be between 10 and 3000 characters"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Text must be between 10 and 3000 characters")
		return
	}

	ticket, err := h.service.CreateTicket(r.Context(), companyID, accountID, &req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create ticket: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("ticket_id", ticket.ID),
		logs.WithMetadata("theme", req.Theme),
	)

	core.SendCreated(w, ticket, "Ticket created successfully")
}

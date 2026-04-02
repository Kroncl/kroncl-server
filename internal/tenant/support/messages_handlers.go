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

// GetMessages возвращает сообщения тикета
func (h *Handlers) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	pagination := core.GetDefaultPaginationParams(r)

	messages, total, err := h.service.GetMessages(r.Context(), ticketID, pagination.Page, pagination.Limit)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get messages: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("ticket_id", ticketID),
		logs.WithMetadata("result_count", len(messages)),
	)

	response := map[string]interface{}{
		"messages": messages,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Messages retrieved successfully")
}

// UpdateMessageReadStatus обновляет статус прочтения сообщения
func (h *Handlers) UpdateMessageReadStatus(w http.ResponseWriter, r *http.Request) {
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

	messageID := chi.URLParam(r, "messageId")
	if messageID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Message ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Message ID required")
		return
	}

	var req struct {
		Read bool `json:"read"`
	}
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

	// Проверяем доступ к тикету
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

	// Проверяем, что сообщение принадлежит этому тикету
	if err := h.service.CheckMessageAccess(r.Context(), ticketID, messageID); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
			logs.WithMetadata("message_id", messageID),
		)
		core.SendNotFound(w, "Message not found")
		return
	}

	if err := h.service.UpdateMessageReadStatus(r.Context(), messageID, req.Read); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
			logs.WithMetadata("message_id", messageID),
		)
		core.SendNotFound(w, err.Error())
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("ticket_id", ticketID),
		logs.WithMetadata("message_id", messageID),
		logs.WithMetadata("read", req.Read),
	)

	core.SendSuccess(w, nil, "Message read status updated successfully")
}

// CreateMessage создаёт новое сообщение в тикете
func (h *Handlers) CreateMessage(w http.ResponseWriter, r *http.Request) {
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

	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Ticket ID required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Ticket ID required")
		return
	}

	// Проверяем доступ к тикету
	if err := h.service.CheckTicketAccess(r.Context(), companyID, ticketID); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendNotFound(w, "Ticket not found")
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid request body")
		return
	}

	// Валидация
	if req.Text == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Text is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Text is required")
		return
	}

	if len(req.Text) < 10 || len(req.Text) > 3000 {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Text must be between 10 and 3000 characters"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Text must be between 10 and 3000 characters")
		return
	}

	message, err := h.service.CreateMessage(r.Context(), ticketID, accountID, req.Text)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create message: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("ticket_id", ticketID),
		logs.WithMetadata("message_id", message.ID),
	)

	core.SendCreated(w, message, "Message created successfully")
}

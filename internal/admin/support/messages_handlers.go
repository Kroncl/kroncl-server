package adminsupport

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetTicketMessages(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		core.SendValidationError(w, "ticketId is required")
		return
	}

	messages, err := h.service.GetTicketMessages(r.Context(), ticketID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get messages: %v", err))
		return
	}

	core.SendSuccess(w, messages, "Ticket messages.")
}

func (h *Handlers) CreateAdminMessage(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		core.SendValidationError(w, "ticketId is required")
		return
	}

	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Unauthorized")
		return
	}

	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	// Валидация длины
	if len(req.Text) < 10 {
		core.SendValidationError(w, "Message must be at least 10 characters")
		return
	}
	if len(req.Text) > 3000 {
		core.SendValidationError(w, "Message must not exceed 3000 characters")
		return
	}

	message, err := h.service.CreateAdminMessage(r.Context(), ticketID, account.UserID, req.Text)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to create message: %v", err))
		return
	}

	core.SendSuccess(w, message, "Message created.")
}

func (h *Handlers) UpdateAdminMessage(w http.ResponseWriter, r *http.Request) {
	messageID := chi.URLParam(r, "messageId")
	if messageID == "" {
		core.SendValidationError(w, "messageId is required")
		return
	}

	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Unauthorized")
		return
	}

	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	if req.Text == "" {
		core.SendValidationError(w, "Message text is required")
		return
	}

	if len(req.Text) < 10 || len(req.Text) > 3000 {
		core.SendValidationError(w, "Message must be between 10 and 3000 characters")
		return
	}

	message, err := h.service.UpdateAdminMessage(r.Context(), messageID, account.UserID, req.Text)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to update message: %v", err))
		return
	}

	core.SendSuccess(w, message, "Message updated.")
}

func (h *Handlers) DeleteAdminMessage(w http.ResponseWriter, r *http.Request) {
	messageID := chi.URLParam(r, "messageId")
	if messageID == "" {
		core.SendValidationError(w, "messageId is required")
		return
	}

	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Unauthorized")
		return
	}

	err := h.service.DeleteAdminMessage(r.Context(), messageID, account.UserID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to delete message: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Message deleted.")
}

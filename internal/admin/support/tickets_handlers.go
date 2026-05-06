package adminsupport

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/support"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetAllTickets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var status *support.TicketStatus
	if s := query.Get("status"); s != "" {
		statusVal := support.TicketStatus(s)
		status = &statusVal
	}

	params := core.GetDefaultPaginationParams(r)

	tickets, pagination, err := h.service.GetAllTickets(r.Context(), status, params)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get tickets: %v", err))
		return
	}

	response := map[string]interface{}{
		"tickets":    tickets,
		"pagination": pagination,
	}

	core.SendSuccess(w, response, "Support tickets list.")
}

func (h *Handlers) GetTicketByID(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		core.SendValidationError(w, "ticketId is required")
		return
	}

	ticket, err := h.service.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get ticket: %v", err))
		return
	}

	core.SendSuccess(w, ticket, "Ticket details.")
}

func (h *Handlers) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		core.SendValidationError(w, "ticketId is required")
		return
	}

	var req struct {
		Status support.TicketStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request body")
		return
	}

	if req.Status != support.TicketStatusClosed && req.Status != support.TicketStatusRevoked {
		core.SendValidationError(w, "Status must be 'closed' or 'revoked'")
		return
	}

	err := h.service.UpdateTicketStatus(r.Context(), ticketID, req.Status)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to update ticket status: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Ticket status updated.")
}

func (h *Handlers) AssignTicket(w http.ResponseWriter, r *http.Request) {
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

	err := h.service.AssignTicketWithCheck(r.Context(), ticketID, account.UserID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to assign ticket: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Ticket assigned to you.")
}

func (h *Handlers) UnassignTicket(w http.ResponseWriter, r *http.Request) {
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

	err := h.service.UnassignTicket(r.Context(), ticketID, account.UserID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to unassign ticket: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Ticket unassigned.")
}

func (h *Handlers) CloseTicket(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		core.SendValidationError(w, "ticketId is required")
		return
	}

	err := h.service.UpdateTicketStatus(r.Context(), ticketID, support.TicketStatusClosed)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to close ticket: %v", err))
		return
	}

	core.SendSuccess(w, nil, "Ticket closed.")
}

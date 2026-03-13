package accounts

import (
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/core"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// GetAccountInvitations возвращает приглашения для текущего пользователя
func (h *Handlers) GetAccountInvitations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Получаем параметры запроса
	query := r.URL.Query()

	// Фильтр по статусу
	status := query.Get("status")

	// Пагинация
	paginationParams := core.GetDefaultPaginationParams(r)

	// Формируем запрос
	req := companies.GetInvitationsByEmailRequest{
		Status:           status,
		PaginationParams: paginationParams,
	}

	// Получаем приглашения
	response, err := h.service.GetAccountInvitations(r.Context(), claims.UserID, req)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get invitations: %v", err))
		return
	}

	// Отправляем ответ
	core.SendSuccess(w, response, "Invitations retrieved successfully")
}

// AcceptAccountInvitation принимает приглашение
func (h *Handlers) AcceptAccountInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Извлекаем ID приглашения из URL параметра
	invitationID := chi.URLParam(r, "invitationId")

	if invitationID == "" {
		core.SendValidationError(w, "Invitation ID is required")
		return
	}

	// Принимаем приглашение
	invitation, err := h.service.AcceptInvitation(r.Context(), claims.UserID, invitationID)
	if err != nil {
		// Определяем тип ошибки для соответствующего HTTP статуса
		switch {
		case strings.Contains(err.Error(), "account must be confirmed"):
			core.SendValidationError(w, "You must confirm your email before accepting invitations")
		case strings.Contains(err.Error(), "invitation does not belong"):
			core.SendUnauthorized(w, "This invitation does not belong to you")
		case strings.Contains(err.Error(), "invitation not found"):
			core.SendNotFound(w, "Invitation not found")
		case strings.Contains(err.Error(), "invitation is not in waiting status"):
			core.SendValidationError(w, "This invitation is no longer valid")
		case strings.Contains(err.Error(), "is already a member"):
			core.SendValidationError(w, "You are already a member of this company")
		case strings.Contains(err.Error(), "user is already a member"):
			core.SendValidationError(w, "You are already a member of this company")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to accept invitation: %v", err))
		}
		return
	}

	core.SendSuccess(w, invitation, "Invitation accepted successfully")
}

// RejectAccountInvitation отклоняет приглашение
func (h *Handlers) RejectAccountInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем пользователя из контекста
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Извлекаем ID приглашения из URL параметра
	invitationID := chi.URLParam(r, "invitationId")

	if invitationID == "" {
		core.SendValidationError(w, "Invitation ID is required")
		return
	}

	// Отклоняем приглашение
	invitation, err := h.service.RejectInvitation(r.Context(), claims.UserID, invitationID)
	if err != nil {
		// Определяем тип ошибки
		switch {
		case strings.Contains(err.Error(), "account must be confirmed"):
			core.SendValidationError(w, "You must confirm your email before rejecting invitations")
		case strings.Contains(err.Error(), "invitation does not belong"):
			core.SendUnauthorized(w, "This invitation does not belong to you")
		case strings.Contains(err.Error(), "invitation not found"):
			core.SendNotFound(w, "Invitation not found")
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to reject invitation: %v", err))
		}
		return
	}

	core.SendSuccess(w, invitation, "Invitation rejected successfully")
}

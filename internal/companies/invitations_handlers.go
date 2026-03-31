package companies

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/core"
	"kroncl-server/utils"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ---------
// INVITATIONS
// ---------

// возвращает список приглашений в компанию с пагинацией и фильтрацией
func (h *Handlers) GetCompanyInvitations(w http.ResponseWriter, r *http.Request) {
	// Получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Парсим параметры запроса
	query := r.URL.Query()

	// Параметры пагинации
	paginationParams := core.GetDefaultPaginationParams(r)

	// Поиск по email
	search := query.Get("search")

	// Фильтр по статусу
	status := query.Get("status")

	// Валидация статуса
	if status != "" {
		err := ValidateInvitationStatus(status)
		if err != nil {
			core.SendValidationError(w, err.Error())
			return
		}
	}

	// Формируем запрос
	req := GetInvitationsRequest{
		Search:           search,
		Status:           status,
		PaginationParams: paginationParams,
	}

	// Получаем приглашения через сервис
	response, err := h.service.GetCompanyInvitations(r.Context(), companyID, req)
	if err != nil {
		// Проверяем тип ошибки для соответствующего HTTP статуса
		if strings.Contains(err.Error(), "invalid status filter") {
			core.SendValidationError(w, err.Error())
		} else {
			core.SendInternalError(w, fmt.Sprintf("Failed to get company invitations: %v", err))
		}
		return
	}

	// Отправляем успешный ответ
	core.SendSuccess(w, response, "Company invitations retrieved successfully")
}

// создает новое приглашение в компанию
func (h *Handlers) CreateCompanyInvitation(w http.ResponseWriter, r *http.Request) {

	// Получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Проверяем авторизацию
	account, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		core.SendUnauthorized(w, "Authentication required")
		return
	}

	// Парсим тело запроса
	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendValidationError(w, "Invalid request format")
		return
	}

	// Валидация email
	if !utils.IsValidEmail(req.Email) {
		core.SendValidationError(w, "Invalid email format")
		return
	}

	// Создаем приглашение через сервис
	response, err := h.service.CreateInvitationAtomic(
		r.Context(),
		companyID,
		account.UserID,
		&req,
	)
	if err != nil {
		// Проверяем тип ошибки для соответствующего HTTP статуса
		switch {
		case strings.Contains(err.Error(), "invalid email format"):
			core.SendValidationError(w, err.Error())
		case strings.Contains(err.Error(), "company not found"):
			core.SendNotFound(w, err.Error())
		case strings.Contains(err.Error(), "is already a member"):
			core.SendValidationError(w, err.Error())
		case strings.Contains(err.Error(), "Invitation already exists"):
			// Если приглашение уже существует, возвращаем 200 с информацией
			core.SendSuccess(w, response, response.Message)
			return
		default:
			core.SendInternalError(w, fmt.Sprintf("Failed to create invitation: %v", err))
		}
		return
	}

	// Отправляем успешный ответ
	core.SendCreated(w, response.Invitation, response.Message)
}

// RevokeInvitation отзывает (удаляет) приглашение
// + проверка принадлежности
func (h *Handlers) RevokeInvitation(w http.ResponseWriter, r *http.Request) {

	// Получаем ID компании из URL
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	// Получаем ID приглашения из URL
	invitationID := chi.URLParam(r, "invitationId")
	if invitationID == "" {
		core.SendValidationError(w, "Invitation ID required")
		return
	}

	// Дополнительно: проверяем, что приглашение принадлежит этой компании
	invitation, err := h.service.GetInvitationByID(r.Context(), invitationID)
	if err != nil {
		core.SendNotFound(w, "Invitation not found")
		return
	}

	if invitation.CompanyID != companyID {
		core.SendUnauthorized(w, "Invitation does not belong to this company")
		return
	}

	// Отзываем приглашение
	err = h.service.WithdrawInvitation(r.Context(), invitationID)
	if err != nil {
		if strings.Contains(err.Error(), "invitation not found") {
			core.SendNotFound(w, err.Error())
		} else {
			core.SendInternalError(w, fmt.Sprintf("Failed to revoke invitation: %v", err))
		}
		return
	}

	// Отправляем успешный ответ
	core.SendSuccess(w, map[string]interface{}{
		"invitation_id": invitationID,
		"company_id":    companyID,
		"revoked":       true,
	}, "Invitation revoked successfully")
}

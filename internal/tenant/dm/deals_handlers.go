package dm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"
)

// ---------
// DEAL TYPES
// ---------

func (h *Handlers) GetDealType(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	typeID := r.PathValue("typeId")
	if typeID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Type ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Type ID is required.")
		return
	}

	dealType, err := h.repository.GetDealTypeByID(r.Context(), typeID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal type not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("type_id", typeID),
		)
		core.SendNotFound(w, "Deal type not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("type_id", typeID),
	)

	core.SendSuccess(w, dealType, "Deal type retrieved successfully.")
}

func (h *Handlers) GetDealTypes(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)
	search := r.URL.Query().Get("search")

	dealTypes, total, err := h.repository.GetDealTypes(r.Context(), pagination.Page, pagination.Limit, search)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deal types: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"search": search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(dealTypes)),
	)

	response := map[string]interface{}{
		"deal_types": dealTypes,
		"pagination": core.NewPagination(
			int(total),
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Deal types retrieved successfully.")
}

func (h *Handlers) CreateDealType(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateDealTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Name is required.")
		return
	}

	dealType, err := h.repository.CreateDealType(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create deal type: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("type_id", dealType.ID),
		logs.WithMetadata("name", req.Name),
	)

	core.SendSuccess(w, dealType, "Deal type created successfully.")
}

func (h *Handlers) UpdateDealType(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	typeID := r.PathValue("typeId")
	if typeID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Type ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Type ID is required.")
		return
	}

	var req UpdateDealTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	dealType, err := h.repository.UpdateDealType(r.Context(), typeID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "deal type not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Deal type not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("type_id", typeID),
			)
			core.SendNotFound(w, "Deal type not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update deal type: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("type_id", typeID),
	)

	core.SendSuccess(w, dealType, "Deal type updated successfully.")
}

func (h *Handlers) DeleteDealType(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	typeID := r.PathValue("typeId")
	if typeID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_DELETE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Type ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Type ID is required.")
		return
	}

	err := h.repository.DeleteDealType(r.Context(), typeID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "deal type not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Deal type not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("type_id", typeID),
			)
			core.SendNotFound(w, "Deal type not found.")
		case strings.Contains(errorMsg, "cannot delete deal type that is used"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("type_id", typeID),
			)
			core.SendValidationError(w, "Cannot delete deal type that is used in deals.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to delete deal type: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_TYPES_DELETE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("type_id", typeID),
		logs.WithMetadata("action", "delete"),
	)

	core.SendSuccess(w, map[string]interface{}{
		"type_id": typeID,
		"deleted": true,
	}, "Deal type deleted successfully.")
}

// ---------
// DEAL STATUSES
// ---------

func (h *Handlers) GetDealStatus(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	statusID := r.PathValue("statusId")
	if statusID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Status ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Status ID is required.")
		return
	}

	status, err := h.repository.GetDealStatusByID(r.Context(), statusID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal status not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("status_id", statusID),
		)
		core.SendNotFound(w, "Deal status not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("status_id", statusID),
	)

	core.SendSuccess(w, status, "Deal status retrieved successfully.")
}

func (h *Handlers) GetDealStatuses(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)
	search := r.URL.Query().Get("search")

	statuses, total, err := h.repository.GetDealStatuses(r.Context(), pagination.Page, pagination.Limit, search)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deal statuses: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"search": search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(statuses)),
	)

	response := map[string]interface{}{
		"statuses": statuses,
		"pagination": core.NewPagination(
			int(total),
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Deal statuses retrieved successfully.")
}

func (h *Handlers) CreateDealStatus(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateDealStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Name is required.")
		return
	}

	if req.SortOrder < 1 {
		req.SortOrder = 1
	}

	status, err := h.repository.CreateDealStatus(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create deal status: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("status_id", status.ID),
		logs.WithMetadata("name", req.Name),
	)

	core.SendSuccess(w, status, "Deal status created successfully.")
}

func (h *Handlers) UpdateDealStatus(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	statusID := r.PathValue("statusId")
	if statusID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Status ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Status ID is required.")
		return
	}

	var req UpdateDealStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	status, err := h.repository.UpdateDealStatus(r.Context(), statusID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "deal status not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Deal status not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("status_id", statusID),
			)
			core.SendNotFound(w, "Deal status not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update deal status: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("status_id", statusID),
	)

	core.SendSuccess(w, status, "Deal status updated successfully.")
}

func (h *Handlers) DeleteDealStatus(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	statusID := r.PathValue("statusId")
	if statusID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_DELETE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Status ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Status ID is required.")
		return
	}

	err := h.repository.DeleteDealStatus(r.Context(), statusID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "deal status not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Deal status not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("status_id", statusID),
			)
			core.SendNotFound(w, "Deal status not found.")
		case strings.Contains(errorMsg, "cannot delete deal status that is used"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("status_id", statusID),
			)
			core.SendValidationError(w, "Cannot delete deal status that is used in deals.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to delete deal status: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_STATUSES_DELETE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("status_id", statusID),
		logs.WithMetadata("action", "delete"),
	)

	core.SendSuccess(w, map[string]interface{}{
		"status_id": statusID,
		"deleted":   true,
	}, "Deal status deleted successfully.")
}

// ----------
// DEALS
// ----------

func (h *Handlers) CreateDeal(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateDealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	deal, err := h.repository.CreateDeal(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create deal: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("deal_id", deal.ID),
	)

	core.SendSuccess(w, deal, "Deal created successfully.")
}

func (h *Handlers) GetDeal(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dealID := r.PathValue("dealId")
	if dealID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required.")
		return
	}

	deal, err := h.repository.GetDealWithDetails(r.Context(), dealID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("deal_id", dealID),
		)
		core.SendNotFound(w, "Deal not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("deal_id", dealID),
	)

	core.SendSuccess(w, deal, "Deal retrieved successfully.")
}

func (h *Handlers) UpdateDeal(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dealID := r.PathValue("dealId")
	if dealID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required.")
		return
	}

	var req UpdateDealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	err := h.repository.UpdateDeal(r.Context(), dealID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "deal not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Deal not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("deal_id", dealID),
			)
			core.SendNotFound(w, "Deal not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update deal: %s", errorMsg))
		}
		return
	}

	// Получаем обновленную сделку для ответа
	updatedDeal, err := h.repository.GetDealWithDetails(r.Context(), dealID)
	if err != nil {
		// Даже если не смогли загрузить детали, обновление прошло успешно
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusSuccess),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("deal_id", dealID),
		)
		core.SendSuccess(w, map[string]interface{}{"id": dealID}, "Deal updated successfully.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("deal_id", dealID),
	)

	core.SendSuccess(w, updatedDeal, "Deal updated successfully.")
}

func (h *Handlers) DeleteDeal(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dealID := r.PathValue("dealId")
	if dealID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_DELETE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Deal ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Deal ID is required.")
		return
	}

	err := h.repository.DeleteDeal(r.Context(), dealID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "deal not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Deal not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("deal_id", dealID),
			)
			core.SendNotFound(w, "Deal not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to delete deal: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS_DELETE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("deal_id", dealID),
		logs.WithMetadata("action", "delete"),
	)

	core.SendSuccess(w, map[string]interface{}{
		"deal_id": dealID,
		"deleted": true,
	}, "Deal deleted successfully.")
}

func (h *Handlers) GetDeals(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	var req GetDealsParams
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	// Фильтры
	if typeID := r.URL.Query().Get("type_id"); typeID != "" {
		req.TypeID = &typeID
	}
	if statusID := r.URL.Query().Get("status_id"); statusID != "" {
		req.StatusID = &statusID
	}
	if clientID := r.URL.Query().Get("client_id"); clientID != "" {
		req.ClientID = &clientID
	}
	if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
		req.EmployeeID = &employeeID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	// Группировка
	if groupBy := r.URL.Query().Get("group_by"); groupBy != "" {
		req.GroupBy = &groupBy
	}

	result, total, err := h.repository.GetDealsWithDetails(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get deals: %s", err.Error()))
		return
	}

	// Логируем
	filters := map[string]interface{}{
		"type_id":     req.TypeID,
		"status_id":   req.StatusID,
		"client_id":   req.ClientID,
		"employee_id": req.EmployeeID,
		"search":      req.Search,
		"group_by":    req.GroupBy,
	}

	if req.GroupBy != nil && *req.GroupBy == "status" {
		// Для группировки возвращаем группы
		groups := result.([]DealGroup)
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS, accountID,
			logs.WithStatus(logs.LogStatusSuccess),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("filters", filters),
			logs.WithMetadata("result_count", total),
		)
		core.SendSuccess(w, groups, "Deals grouped by status retrieved successfully.")
	} else {
		// Для обычной пагинации
		deals := result.([]Deal)
		h.logsService.Log(r.Context(), config.PERMISSION_DM_DEALS, accountID,
			logs.WithStatus(logs.LogStatusSuccess),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("filters", filters),
			logs.WithMetadata("pagination", map[string]int{
				"page":  pagination.Page,
				"limit": pagination.Limit,
			}),
			logs.WithMetadata("result_count", len(deals)),
		)

		response := map[string]interface{}{
			"deals": deals,
			"pagination": core.NewPagination(
				int(total),
				pagination.Page,
				pagination.Limit,
			),
		}
		core.SendSuccess(w, response, "Deals retrieved successfully.")
	}
}

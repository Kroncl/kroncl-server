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
// COUNTERPARTIES
// --------

func (h *Handlers) GetCounterparty(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	counterparty, err := h.repository.GetCounterpartyByID(r.Context(), counterpartyID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("counterparty_id", counterpartyID),
		)
		core.SendNotFound(w, "Counterparty not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("counterparty_id", counterpartyID),
	)

	core.SendSuccess(w, counterparty, "Counterparty retrieved successfully.")
}

func (h *Handlers) GetCounterparties(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Парсим query параметры в структуру фильтров
	var filters GetCounterpartiesRequest
	filters.Page = pagination.Page
	filters.Limit = pagination.Limit

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := CounterpartyType(typeStr)
		if t != CounterpartyTypeBank && t != CounterpartyTypeOrganization && t != CounterpartyTypePerson {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid type"),
				logs.WithMetadata("type", typeStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid type. Use 'bank', 'organization' or 'person'.")
			return
		}
		filters.Type = &t
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := CounterpartyStatus(statusStr)
		if s != CounterpartyStatusActive && s != CounterpartyStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid status"),
				logs.WithMetadata("status", statusStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
			return
		}
		filters.Status = &s
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	counterparties, total, err := h.repository.GetCounterparties(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
		filters,
	)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get counterparties: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES, accountID,
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
		logs.WithMetadata("result_count", len(counterparties)),
	)

	response := map[string]interface{}{
		"counterparties": counterparties,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Counterparties retrieved successfully.")
}

func (h *Handlers) CreateCounterparty(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим тело запроса
	var req CreateCounterpartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_CREATE, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Counterparty name is required.")
		return
	}
	if req.Type == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty type is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Counterparty type is required.")
		return
	}
	if req.Type != CounterpartyTypeBank && req.Type != CounterpartyTypeOrganization && req.Type != CounterpartyTypePerson {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid type"),
			logs.WithMetadata("type", req.Type),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid type. Use 'bank', 'organization' or 'person'.")
		return
	}
	if req.Status != "" && req.Status != CounterpartyStatusActive && req.Status != CounterpartyStatusInactive {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid status"),
			logs.WithMetadata("status", req.Status),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
		return
	}

	counterparty, err := h.repository.CreateCounterparty(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create counterparty: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("counterparty_id", counterparty.ID),
		logs.WithMetadata("name", req.Name),
		logs.WithMetadata("type", req.Type),
	)

	core.SendSuccess(w, counterparty, "Counterparty created successfully.")
}

func (h *Handlers) UpdateCounterparty(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateCounterpartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
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
		if *req.Type != CounterpartyTypeBank && *req.Type != CounterpartyTypeOrganization && *req.Type != CounterpartyTypePerson {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid type"),
				logs.WithMetadata("type", *req.Type),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid type. Use 'bank', 'organization' or 'person'.")
			return
		}
	}

	counterparty, err := h.repository.UpdateCounterparty(r.Context(), counterpartyID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Counterparty not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("counterparty_id", counterpartyID),
			)
			core.SendNotFound(w, "Counterparty not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update counterparty: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("counterparty_id", counterpartyID),
	)

	core.SendSuccess(w, counterparty, "Counterparty updated successfully.")
}

func (h *Handlers) ActivateCounterparty(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	counterparty, err := h.repository.ActivateCounterparty(r.Context(), counterpartyID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Counterparty not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("counterparty_id", counterpartyID),
			)
			core.SendNotFound(w, "Counterparty not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to activate counterparty: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("counterparty_id", counterpartyID),
		logs.WithMetadata("action", "activate"),
	)

	core.SendSuccess(w, counterparty, "Counterparty activated successfully.")
}

func (h *Handlers) DeactivateCounterparty(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID контрагента из URL
	counterpartyID := r.PathValue("counterpartyId")
	if counterpartyID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Counterparty ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Counterparty ID is required.")
		return
	}

	counterparty, err := h.repository.DeactivateCounterparty(r.Context(), counterpartyID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "counterparty not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Counterparty not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("counterparty_id", counterpartyID),
			)
			core.SendNotFound(w, "Counterparty not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate counterparty: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_COUNTERPARTIES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("counterparty_id", counterpartyID),
		logs.WithMetadata("action", "deactivate"),
	)

	core.SendSuccess(w, counterparty, "Counterparty deactivated successfully.")
}

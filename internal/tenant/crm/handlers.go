package crm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"
)

type Handlers struct {
	repository  *Repository
	logsService *logs.Service
}

func NewHandlers(repository *Repository, logsService *logs.Service) *Handlers {
	return &Handlers{
		repository:  repository,
		logsService: logsService,
	}
}

// ---------
// SOURCES
// ---------

func (h *Handlers) GetClientSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID источника из URL
	sourceID := r.PathValue("sourceId")
	if sourceID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Source ID is required.")
		return
	}

	source, err := h.repository.GetClientSourceByID(r.Context(), sourceID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("source_id", sourceID),
		)
		core.SendNotFound(w, "Source not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("source_id", sourceID),
	)

	core.SendSuccess(w, source, "Source retrieved successfully.")
}

func (h *Handlers) GetClientSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Формируем запрос с фильтрами
	var req GetSourcesRequest
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	// Type filter
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := SourceType(typeStr)
		switch t {
		case SourceTypeOrganic, SourceTypeSocial, SourceTypeReferral, SourceTypePaid, SourceTypeEmail, SourceTypeOther:
			req.Type = &t
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid source type"),
				logs.WithMetadata("type", typeStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid source type. Use: organic, social, referral, paid, email, other.")
			return
		}
	}

	// Status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := SourceStatus(statusStr)
		if s != SourceStatusActive && s != SourceStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid status"),
				logs.WithMetadata("status", statusStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
			return
		}
		req.Status = &s
	}

	// System filter
	if systemStr := r.URL.Query().Get("system"); systemStr != "" {
		system := systemStr == "true"
		req.System = &system
	}

	// Search filter
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	sources, total, err := h.repository.GetClientSources(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get sources: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"type":   req.Type,
			"status": req.Status,
			"system": req.System,
			"search": req.Search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(sources)),
	)

	response := map[string]interface{}{
		"sources": sources,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Sources retrieved successfully.")
}

func (h *Handlers) CreateClientSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим тело запроса
	var req CreateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Source name is required.")
		return
	}

	if req.Type == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source type is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Source type is required.")
		return
	}

	switch req.Type {
	case SourceTypeOrganic, SourceTypeSocial, SourceTypeReferral, SourceTypePaid, SourceTypeEmail, SourceTypeOther:
		// valid
	default:
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid source type"),
			logs.WithMetadata("type", req.Type),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid source type. Use: organic, social, referral, paid, email, other.")
		return
	}

	if req.Status != "" && req.Status != SourceStatusActive && req.Status != SourceStatusInactive {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid status"),
			logs.WithMetadata("status", req.Status),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
		return
	}

	source, err := h.repository.CreateClientSource(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "already exists"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("name", req.Name),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create source: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("source_id", source.ID),
		logs.WithMetadata("name", req.Name),
		logs.WithMetadata("type", req.Type),
	)

	core.SendSuccess(w, source, "Source created successfully.")
}

func (h *Handlers) UpdateClientSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID источника из URL
	sourceID := r.PathValue("sourceId")
	if sourceID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Source ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
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
		switch *req.Type {
		case SourceTypeOrganic, SourceTypeSocial, SourceTypeReferral, SourceTypePaid, SourceTypeEmail, SourceTypeOther:
			// valid
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid source type"),
				logs.WithMetadata("type", *req.Type),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid source type. Use: organic, social, referral, paid, email, other.")
			return
		}
	}

	// Валидация статуса если указан
	if req.Status != nil {
		if *req.Status != SourceStatusActive && *req.Status != SourceStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid status"),
				logs.WithMetadata("status", *req.Status),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
			return
		}
	}

	source, err := h.repository.UpdateClientSource(r.Context(), sourceID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "source not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Source not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("source_id", sourceID),
			)
			core.SendNotFound(w, "Source not found.")
		case strings.Contains(errorMsg, "cannot update system source"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("source_id", sourceID),
			)
			core.SendValidationError(w, "Cannot update system source.")
		case strings.Contains(errorMsg, "already exists"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update source: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("source_id", sourceID),
	)

	core.SendSuccess(w, source, "Source updated successfully.")
}

func (h *Handlers) ActivateClientSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID источника из URL
	sourceID := r.PathValue("sourceId")
	if sourceID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Source ID is required.")
		return
	}

	source, err := h.repository.ActivateClientSource(r.Context(), sourceID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "source not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Source not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("source_id", sourceID),
			)
			core.SendNotFound(w, "Source not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to activate source: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("source_id", sourceID),
		logs.WithMetadata("action", "activate"),
	)

	core.SendSuccess(w, source, "Source activated successfully.")
}

func (h *Handlers) DeactivateClientSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID источника из URL
	sourceID := r.PathValue("sourceId")
	if sourceID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Source ID is required.")
		return
	}

	// Для деактивации используем Update с установкой статуса Inactive
	req := UpdateSourceRequest{
		Status: &[]SourceStatus{SourceStatusInactive}[0],
	}

	source, err := h.repository.UpdateClientSource(r.Context(), sourceID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "source not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Source not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("source_id", sourceID),
			)
			core.SendNotFound(w, "Source not found.")
		case strings.Contains(errorMsg, "cannot update system source"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("source_id", sourceID),
			)
			core.SendValidationError(w, "Cannot deactivate system source.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate source: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_SOURCES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("source_id", sourceID),
		logs.WithMetadata("action", "deactivate"),
	)

	core.SendSuccess(w, source, "Source deactivated successfully.")
}

// ---------
// CLIENTS
// ---------

// ---------
// CLIENTS
// ---------

func (h *Handlers) GetClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID клиента из URL
	clientID := r.PathValue("clientId")
	if clientID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Client ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Client ID is required.")
		return
	}

	client, err := h.repository.GetClientByID(r.Context(), clientID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Client not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("client_id", clientID),
		)
		core.SendNotFound(w, "Client not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("client_id", clientID),
	)

	core.SendSuccess(w, client, "Client retrieved successfully.")
}

func (h *Handlers) GetClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Формируем запрос с фильтрами
	var req GetClientsRequest
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	// Type filter
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := ClientType(typeStr)
		if t != ClientTypeIndividual && t != ClientTypeLegal {
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid client type"),
				logs.WithMetadata("type", typeStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid client type. Use 'individual' or 'legal'.")
			return
		}
		req.Type = &t
	}

	// Status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := ClientStatus(statusStr)
		if s != ClientStatusActive && s != ClientStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid status"),
				logs.WithMetadata("status", statusStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
			return
		}
		req.Status = &s
	}

	// Search filter
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	clients, total, err := h.repository.GetClients(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get clients: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"type":   req.Type,
			"status": req.Status,
			"search": req.Search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(clients)),
	)

	response := map[string]interface{}{
		"clients": clients,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Clients retrieved successfully.")
}

func (h *Handlers) CreateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим тело запроса
	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация
	if strings.TrimSpace(req.FirstName) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "First name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "First name is required.")
		return
	}

	// Проверяем source_id
	if strings.TrimSpace(req.SourceID) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Source ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Source ID is required.")
		return
	}

	if req.Type == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Client type is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Client type is required.")
		return
	}

	if req.Type != ClientTypeIndividual && req.Type != ClientTypeLegal {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid client type"),
			logs.WithMetadata("type", req.Type),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid client type. Use 'individual' or 'legal'.")
		return
	}

	if req.Status != "" && req.Status != ClientStatusActive && req.Status != ClientStatusInactive {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid status"),
			logs.WithMetadata("status", req.Status),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
		return
	}

	client, err := h.repository.CreateClient(r.Context(), req, req.SourceID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "source not found or not active"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("source_id", req.SourceID),
			)
			core.SendValidationError(w, "Selected source is not available or not active.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create client: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("client_id", client.ID),
		logs.WithMetadata("first_name", req.FirstName),
		logs.WithMetadata("type", req.Type),
		logs.WithMetadata("source_id", req.SourceID),
	)

	core.SendSuccess(w, client, "Client created successfully.")
}

func (h *Handlers) UpdateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID клиента из URL
	clientID := r.PathValue("clientId")
	if clientID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Client ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Client ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
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
		if *req.Type != ClientTypeIndividual && *req.Type != ClientTypeLegal {
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid client type"),
				logs.WithMetadata("type", *req.Type),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid client type. Use 'individual' or 'legal'.")
			return
		}
	}

	// Валидация статуса если указан
	if req.Status != nil {
		if *req.Status != ClientStatusActive && *req.Status != ClientStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid status"),
				logs.WithMetadata("status", *req.Status),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid status. Use 'active' or 'inactive'.")
			return
		}
	}

	client, err := h.repository.UpdateClient(r.Context(), clientID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "client not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Client not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("client_id", clientID),
			)
			core.SendNotFound(w, "Client not found.")
		case strings.Contains(errorMsg, "source with id") && strings.Contains(errorMsg, "not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("source_id", *req.SourceID),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update client: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("client_id", clientID),
	)

	core.SendSuccess(w, client, "Client updated successfully.")
}

func (h *Handlers) ActivateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID клиента из URL
	clientID := r.PathValue("clientId")
	if clientID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Client ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Client ID is required.")
		return
	}

	client, err := h.repository.ActivateClient(r.Context(), clientID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "client not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Client not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("client_id", clientID),
			)
			core.SendNotFound(w, "Client not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to activate client: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("client_id", clientID),
		logs.WithMetadata("action", "activate"),
	)

	core.SendSuccess(w, client, "Client activated successfully.")
}

func (h *Handlers) DeactivateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID клиента из URL
	clientID := r.PathValue("clientId")
	if clientID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Client ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Client ID is required.")
		return
	}

	client, err := h.repository.DeactivateClient(r.Context(), clientID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "client not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Client not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("client_id", clientID),
			)
			core.SendNotFound(w, "Client not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate client: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_CRM_CLIENTS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("client_id", clientID),
		logs.WithMetadata("action", "deactivate"),
	)

	core.SendSuccess(w, client, "Client deactivated successfully.")
}

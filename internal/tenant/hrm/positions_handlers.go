package hrm

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// GetPositions возвращает список должностей с пагинацией
func (h *Handlers) GetPositions(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)
	search := r.URL.Query().Get("search")

	positions, total, err := h.repository.GetPositions(
		r.Context(),
		pagination.Page,
		pagination.Limit,
		search,
	)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get positions: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"search": search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(positions)),
	)

	response := map[string]interface{}{
		"positions": positions,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Positions retrieved successfully.")
}

// GetPositionByID возвращает должность по ID
func (h *Handlers) GetPositionByID(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	positionID := chi.URLParam(r, "positionId")
	if positionID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Position ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Position ID is required.")
		return
	}

	position, err := h.repository.GetPositionByID(r.Context(), positionID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Position not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("position_id", positionID),
		)
		core.SendNotFound(w, "Position not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("position_id", positionID),
	)

	core.SendSuccess(w, position, "Position retrieved successfully.")
}

// CreatePosition создаёт новую должность
func (h *Handlers) CreatePosition(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_CREATE, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Name is required.")
		return
	}

	position, err := h.repository.CreatePosition(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "invalid permissions"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create position: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("position_id", position.ID),
		logs.WithMetadata("name", position.Name),
	)

	core.SendCreated(w, position, "Position created successfully.")
}

// UpdatePosition обновляет должность
func (h *Handlers) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	positionID := chi.URLParam(r, "positionId")
	if positionID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Position ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Position ID is required.")
		return
	}

	var req UpdatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	position, err := h.repository.UpdatePosition(r.Context(), positionID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "position not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Position not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("position_id", positionID),
			)
			core.SendNotFound(w, "Position not found.")
		case strings.Contains(errorMsg, "invalid permissions"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update position: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("position_id", positionID),
	)

	core.SendSuccess(w, position, "Position updated successfully.")
}

// DeletePosition удаляет должность
func (h *Handlers) DeletePosition(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	positionID := chi.URLParam(r, "positionId")
	if positionID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_DELETE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Position ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Position ID is required.")
		return
	}

	err := h.repository.DeletePosition(r.Context(), positionID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "position not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Position not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("position_id", positionID),
			)
			core.SendNotFound(w, "Position not found.")
		case strings.Contains(errorMsg, "employee(s) assigned"):
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("position_id", positionID),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to delete position: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_HRM_POSITIONS_DELETE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("position_id", positionID),
		logs.WithMetadata("action", "delete"),
	)

	core.SendSuccess(w, nil, "Position deleted successfully.")
}

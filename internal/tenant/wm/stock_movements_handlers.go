package wm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
)

// CreateStockMovement создаёт движение (списание)
func (h *Handlers) CreateStockMovement(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateStockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if !IsValidStockMovementType(req.Type) {
		core.SendValidationError(w, "Invalid movement type.")
		return
	}

	if req.Quantity <= 0 {
		core.SendValidationError(w, "Quantity must be greater than zero.")
		return
	}

	movement, err := h.repository.CreateMovement(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "insufficient stock"):
			core.SendValidationError(w, "Недостаточно товара на складе.")
		case strings.Contains(errorMsg, "failed to lock position"):
			core.SendNotFound(w, "Stock position not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_MOVEMENTS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create movement: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_MOVEMENTS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("outcome_batch_id", movement.OutcomeBatchID),
		logs.WithMetadata("position_id", movement.PositionID),
		logs.WithMetadata("quantity", movement.Quantity),
	)

	core.SendSuccess(w, movement, "Movement created successfully.")
}

// CreateStockMovementsBatch создаёт несколько движений атомарно (отгрузка)
func (h *Handlers) CreateStockMovementsBatch(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var requests []CreateStockMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if len(requests) == 0 {
		core.SendValidationError(w, "At least one movement is required.")
		return
	}

	for _, req := range requests {
		if !IsValidStockMovementType(req.Type) {
			core.SendValidationError(w, "Invalid movement type.")
			return
		}
		if req.Quantity <= 0 {
			core.SendValidationError(w, "Quantity must be greater than zero.")
			return
		}
	}

	movements, err := h.repository.CreateMovementsBatch(r.Context(), requests)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "insufficient stock"):
			core.SendValidationError(w, "Недостаточно товара на складе.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_MOVEMENTS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create movements: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_MOVEMENTS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("count", len(movements)),
	)

	core.SendSuccess(w, movements, "Movements created successfully.")
}

// GetMovements возвращает список движений
func (h *Handlers) GetMovements(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	var req GetMovementsParams
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := StockMovementType(typeStr)
		if !IsValidStockMovementType(t) {
			core.SendValidationError(w, "Invalid movement type.")
			return
		}
		req.Type = &t
	}

	if bid := r.URL.Query().Get("outcome_batch_id"); bid != "" {
		req.OutcomeBatchID = &bid
	}

	if pid := r.URL.Query().Get("position_id"); pid != "" {
		req.PositionID = &pid
	}

	movements, total, err := h.repository.GetMovements(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get movements: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("result_count", len(movements)),
	)

	core.SendSuccess(w, map[string]interface{}{
		"movements":  movements,
		"pagination": core.NewPagination(int(total), pagination.Page, pagination.Limit),
	}, "Movements retrieved successfully.")
}

package wm

import (
	"fmt"
	"net/http"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
)

// GetStockPosition возвращает позицию с деталями
func (h *Handlers) GetStockPosition(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	positionID := r.PathValue("positionId")
	if positionID == "" {
		core.SendError(w, http.StatusBadRequest, "Position ID is required.")
		return
	}

	position, err := h.repository.GetStockPositionWithDetails(r.Context(), positionID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Position not found"),
			logs.WithMetadata("position_id", positionID),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendNotFound(w, "Stock position not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("position_id", positionID),
	)

	core.SendSuccess(w, position, "Stock position retrieved successfully.")
}

// GetStockPositions возвращает список позиций с пагинацией
func (h *Handlers) GetStockPositions(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	var req GetStockPositionsParams
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := StockPositionType(typeStr)
		if !IsValidStockPositionType(t) {
			core.SendValidationError(w, "Invalid position type. Use 'batch' or 'serial'.")
			return
		}
		req.Type = &t
	}

	if unitID := r.URL.Query().Get("unit_id"); unitID != "" {
		req.UnitID = &unitID
	}

	if batchID := r.URL.Query().Get("income_batch_id"); batchID != "" {
		req.IncomeBatchID = &batchID
	}

	if inStockStr := r.URL.Query().Get("in_stock"); inStockStr != "" {
		inStock := inStockStr == "true"
		req.InStock = &inStock
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	positions, total, err := h.repository.GetStockPositions(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get stock positions: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
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

	core.SendSuccess(w, response, "Stock positions retrieved successfully.")
}

// GetStockBalance возвращает остатки
func (h *Handlers) GetStockBalance(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var unitID *string
	if id := r.URL.Query().Get("unit_id"); id != "" {
		unitID = &id
	}

	items, err := h.repository.GetStockBalance(r.Context(), unitID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BALANCE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get stock balance: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BALANCE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("result_count", len(items)),
	)

	core.SendSuccess(w, items, "Stock balance retrieved successfully.")
}

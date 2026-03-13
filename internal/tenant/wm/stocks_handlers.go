package wm

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
// STOCK BATCHES
// ---------

func (h *Handlers) CreateStockBatch(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateStockBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if req.Direction == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Direction is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Direction is required.")
		return
	}

	if len(req.Positions) == 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "At least one position is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "At least one position is required.")
		return
	}

	result, err := h.repository.CreateStockBatchWithPositions(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "unit with id") && strings.Contains(errorMsg, "not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Одна из указанных товарных позиций не найдена.")
		case strings.Contains(errorMsg, "too many serial positions"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Слишком много поштучных позиций в одной партии. Максимум: 1000.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create stock batch: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("batch_id", result.BatchID),
		logs.WithMetadata("direction", result.Direction),
		logs.WithMetadata("positions_count", len(result.Positions)),
	)

	core.SendSuccess(w, result, "Stock batch created successfully.")
}

func (h *Handlers) GetStockBatch(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	batchID := r.PathValue("batchId")
	if batchID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Batch ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Batch ID is required.")
		return
	}

	batch, err := h.repository.GetStockBatchWithPositions(r.Context(), batchID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Batch not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("batch_id", batchID),
		)
		core.SendNotFound(w, "Stock batch not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("batch_id", batchID),
	)

	core.SendSuccess(w, batch, "Stock batch retrieved successfully.")
}

func (h *Handlers) GetStockBatches(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	var req GetStockBatchesParams
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	if dirStr := r.URL.Query().Get("direction"); dirStr != "" {
		dir := StockDirection(dirStr)
		if dir != StockDirectionIncome && dir != StockDirectionOutcome {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid direction"),
				logs.WithMetadata("direction", dirStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid direction. Use 'income' or 'outcome'.")
			return
		}
		req.Direction = &dir
	}

	if unitID := r.URL.Query().Get("unit_id"); unitID != "" {
		req.UnitID = &unitID
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	batches, total, err := h.repository.GetStockBatches(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get stock batches: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"direction": req.Direction,
			"unit_id":   req.UnitID,
			"search":    req.Search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(batches)),
	)

	response := map[string]interface{}{
		"batches": batches,
		"pagination": core.NewPagination(
			int(total),
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Stock batches retrieved successfully.")
}

// ---------
// STOCK POSITIONS
// ---------

func (h *Handlers) GetStockPosition(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	positionID := r.PathValue("positionId")
	if positionID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Position ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Position ID is required.")
		return
	}

	position, err := h.repository.GetStockPositionWithDetails(r.Context(), positionID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Position not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("position_id", positionID),
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
		if t != StockPositionTypeBatch && t != StockPositionTypeSerial {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_POSITIONS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid position type"),
				logs.WithMetadata("type", typeStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid position type. Use 'batch' or 'serial'.")
			return
		}
		req.Type = &t
	}

	if unitID := r.URL.Query().Get("unit_id"); unitID != "" {
		req.UnitID = &unitID
	}

	if batchID := r.URL.Query().Get("batch_id"); batchID != "" {
		req.BatchID = &batchID
	}

	if inStockStr := r.URL.Query().Get("in_stock"); inStockStr != "" {
		inStock := inStockStr == "true"
		req.InStock = &inStock
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
		logs.WithMetadata("filters", map[string]interface{}{
			"type":     req.Type,
			"unit_id":  req.UnitID,
			"batch_id": req.BatchID,
			"in_stock": req.InStock,
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
			int(total),
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Stock positions retrieved successfully.")
}

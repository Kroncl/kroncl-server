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

// CreateStockBatch создаёт батч с позициями
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

	if !IsValidStockDirection(req.Direction) {
		core.SendValidationError(w, "Invalid direction. Use 'income' or 'outcome'.")
		return
	}

	if len(req.Positions) == 0 {
		core.SendValidationError(w, "At least one position is required.")
		return
	}

	result, err := h.repository.CreateStockBatchWithPositions(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "unit with id") && strings.Contains(errorMsg, "not found"):
			core.SendValidationError(w, "Одна из указанных товарных позиций не найдена.")
		case strings.Contains(errorMsg, "too many serial positions"):
			core.SendValidationError(w, "Слишком много поштучных позиций в одной партии.")
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

// CreateStockBatchOnly создаёт пустой батч (черновик)
func (h *Handlers) CreateStockBatchOnly(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateStockBatchOnlyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if !IsValidStockDirection(req.Direction) {
		core.SendValidationError(w, "Invalid direction.")
		return
	}

	batch, err := h.repository.CreateStockBatchOnly(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create stock batch: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("batch_id", batch.ID),
	)

	core.SendSuccess(w, batch, "Stock batch created successfully.")
}

// UpdateStockBatchStatus обновляет статус батча
func (h *Handlers) UpdateStockBatchStatus(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	batchID := r.PathValue("batchId")
	if batchID == "" {
		core.SendError(w, http.StatusBadRequest, "Batch ID is required.")
		return
	}

	var req UpdateStockBatchStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if !IsValidStockBatchStatus(req.Status) {
		core.SendValidationError(w, "Invalid status.")
		return
	}

	// проверяем существование + валидность перехода
	current, err := h.repository.GetStockBatchByID(r.Context(), batchID)
	if err != nil {
		core.SendNotFound(w, "Stock batch not found.")
		return
	}

	if !current.Status.CanTransitionTo(req.Status) {
		core.SendValidationError(w,
			fmt.Sprintf("Cannot transition from '%s' to '%s'.", current.Status, req.Status))
		return
	}

	batch, err := h.repository.UpdateStockBatchStatus(r.Context(), batchID, req.Status)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to update stock batch status: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("batch_id", batchID),
		logs.WithMetadata("status", req.Status),
	)

	core.SendSuccess(w, batch, "Stock batch status updated successfully.")
}

// GetStockBatch возвращает батч с позициями
func (h *Handlers) GetStockBatch(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	batchID := r.PathValue("batchId")
	if batchID == "" {
		core.SendError(w, http.StatusBadRequest, "Batch ID is required.")
		return
	}

	batch, err := h.repository.GetStockBatchWithPositions(r.Context(), batchID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_STOCKS_BATCHES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Batch not found"),
			logs.WithMetadata("batch_id", batchID),
			logs.WithMetadata("path", r.URL.Path),
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

// GetStockBatches возвращает список батчей с пагинацией
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
		if !IsValidStockDirection(dir) {
			core.SendValidationError(w, "Invalid direction. Use 'income' or 'outcome'.")
			return
		}
		req.Direction = &dir
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := StockBatchStatus(statusStr)
		if !IsValidStockBatchStatus(status) {
			core.SendValidationError(w, "Invalid status.")
			return
		}
		req.Status = &status
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
		logs.WithMetadata("result_count", len(batches)),
	)

	response := map[string]interface{}{
		"batches": batches,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Stock batches retrieved successfully.")
}

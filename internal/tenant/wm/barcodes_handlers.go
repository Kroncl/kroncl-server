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

// CreateBarcode создаёт баркод в словаре
func (h *Handlers) CreateBarcode(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateBarcodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if strings.TrimSpace(req.Barcode) == "" {
		core.SendValidationError(w, "Barcode is required.")
		return
	}

	exists, err := h.repository.BarcodeExists(r.Context(), req.Barcode)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to check barcode: %s", err.Error()))
		return
	}
	if exists {
		core.SendValidationError(w, "Barcode already exists.")
		return
	}

	barcode, err := h.repository.CreateBarcode(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create barcode: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("barcode", req.Barcode),
	)

	core.SendSuccess(w, barcode, "Barcode created successfully.")
}

// GetBarcode возвращает баркод по значению (query: ?barcode=...)
func (h *Handlers) GetBarcode(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	value := r.URL.Query().Get("barcode")
	if value == "" {
		core.SendValidationError(w, "Barcode is required.")
		return
	}

	barcode, err := h.repository.GetBarcodeByValue(r.Context(), value)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Barcode not found"),
			logs.WithMetadata("barcode", value),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendNotFound(w, "Barcode not found.")
		return
	}

	core.SendSuccess(w, barcode, "Barcode retrieved successfully.")
}

// GetBarcodeByID возвращает баркод по ID (path: /{barcodeId})
func (h *Handlers) GetBarcodeByID(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id := r.PathValue("barcodeId")
	if id == "" {
		core.SendError(w, http.StatusBadRequest, "Barcode ID is required.")
		return
	}

	barcode, err := h.repository.GetBarcodeByID(r.Context(), id)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Barcode not found"),
			logs.WithMetadata("barcode_id", id),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendNotFound(w, "Barcode not found.")
		return
	}

	core.SendSuccess(w, barcode, "Barcode retrieved successfully.")
}

// GetBarcodes возвращает список баркодов
func (h *Handlers) GetBarcodes(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	var req GetBarcodesParams
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	if unitID := r.URL.Query().Get("catalog_unit_id"); unitID != "" {
		req.CatalogUnitID = &unitID
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	barcodes, total, err := h.repository.GetBarcodes(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get barcodes: %s", err.Error()))
		return
	}

	response := map[string]interface{}{
		"barcodes": barcodes,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Barcodes retrieved successfully.")
}

// UpdateBarcode обновляет баркод
func (h *Handlers) UpdateBarcode(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id := r.PathValue("barcodeId")
	if id == "" {
		core.SendError(w, http.StatusBadRequest, "Barcode ID is required.")
		return
	}

	var req UpdateBarcodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	barcode, err := h.repository.UpdateBarcode(r.Context(), id, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "not found"):
			core.SendNotFound(w, "Barcode not found.")
		case strings.Contains(errorMsg, "cannot be empty"):
			core.SendValidationError(w, "Barcode cannot be empty.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("barcode_id", id),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update barcode: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("barcode_id", id),
	)

	core.SendSuccess(w, barcode, "Barcode updated successfully.")
}

// DeleteBarcode удаляет баркод
func (h *Handlers) DeleteBarcode(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id := r.PathValue("barcodeId")
	if id == "" {
		core.SendError(w, http.StatusBadRequest, "Barcode ID is required.")
		return
	}

	deleted, err := h.repository.DeleteBarcode(r.Context(), id)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES_DELETE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("barcode_id", id),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to delete barcode: %s", err.Error()))
		return
	}

	if !deleted {
		core.SendNotFound(w, "Barcode not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_BARCODES_DELETE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("barcode_id", id),
	)

	core.SendSuccess(w, nil, "Barcode deleted successfully.")
}

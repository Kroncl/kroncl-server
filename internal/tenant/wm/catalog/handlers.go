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
// CATEGORIES
// ---------

func (h *Handlers) GetCatalogCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	category, err := h.repository.GetCatalogCategoryByID(r.Context(), categoryID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("category_id", categoryID),
		)
		core.SendNotFound(w, "Category not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", categoryID),
	)

	core.SendSuccess(w, category, "Category retrieved successfully.")
}

func (h *Handlers) GetCatalogCategories(w http.ResponseWriter, r *http.Request) {
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
	var req GetCategoriesRequest
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	// Status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := CategoryStatus(statusStr)
		if s != CategoryStatusActive && s != CategoryStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES, accountID,
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

	// Parent filter
	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		req.ParentID = &parentID
	}

	// Search filter
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	categories, total, err := h.repository.GetCatalogCategories(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get categories: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"status":    req.Status,
			"parent_id": req.ParentID,
			"search":    req.Search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(categories)),
	)

	response := map[string]interface{}{
		"categories": categories,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Categories retrieved successfully.")
}

func (h *Handlers) CreateCatalogCategory(w http.ResponseWriter, r *http.Request) {
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
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_CREATE, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category name is required.")
		return
	}

	category, err := h.repository.CreateCatalogCategory(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "already exists"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("name", req.Name),
			)
			core.SendValidationError(w, errorMsg)
		case strings.Contains(errorMsg, "parent category with id") && strings.Contains(errorMsg, "not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("parent_id", req.ParentID),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create category: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", category.ID),
		logs.WithMetadata("name", req.Name),
	)

	core.SendSuccess(w, category, "Category created successfully.")
}

func (h *Handlers) UpdateCatalogCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация статуса если указан
	if req.Status != nil {
		if *req.Status != CategoryStatusActive && *req.Status != CategoryStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
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

	category, err := h.repository.UpdateCatalogCategory(r.Context(), categoryID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "category not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Category not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("category_id", categoryID),
			)
			core.SendNotFound(w, "Category not found.")
		case strings.Contains(errorMsg, "already exists"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, errorMsg)
		case strings.Contains(errorMsg, "parent category with id") && strings.Contains(errorMsg, "not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("parent_id", req.ParentID),
			)
			core.SendValidationError(w, errorMsg)
		case strings.Contains(errorMsg, "category cannot be parent of itself"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update category: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", categoryID),
	)

	core.SendSuccess(w, category, "Category updated successfully.")
}

func (h *Handlers) ActivateCatalogCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	category, err := h.repository.ActivateCatalogCategory(r.Context(), categoryID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "category not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Category not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("category_id", categoryID),
			)
			core.SendNotFound(w, "Category not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to activate category: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", categoryID),
		logs.WithMetadata("action", "activate"),
	)

	core.SendSuccess(w, category, "Category activated successfully.")
}

func (h *Handlers) DeactivateCatalogCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	category, err := h.repository.DeactivateCatalogCategory(r.Context(), categoryID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "category not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Category not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("category_id", categoryID),
			)
			core.SendNotFound(w, "Category not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate category: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_CATEGORIES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", categoryID),
		logs.WithMetadata("action", "deactivate"),
	)

	core.SendSuccess(w, category, "Category deactivated successfully.")
}

// ---------
// UNITS
// ---------

func (h *Handlers) GetCatalogUnit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID юнита из URL
	unitID := r.PathValue("unitId")
	if unitID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Unit ID is required.")
		return
	}

	unit, err := h.repository.GetCatalogUnitByID(r.Context(), unitID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("unit_id", unitID),
		)
		core.SendNotFound(w, "Unit not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("unit_id", unitID),
	)

	core.SendSuccess(w, unit, "Unit retrieved successfully.")
}

func (h *Handlers) GetCatalogUnits(w http.ResponseWriter, r *http.Request) {
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
	var req GetUnitsRequest
	req.Page = pagination.Page
	req.Limit = pagination.Limit

	// Type filter
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		t := UnitType(typeStr)
		if t != UnitTypeProduct && t != UnitTypeService {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid unit type"),
				logs.WithMetadata("type", typeStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid unit type. Use 'product' or 'service'.")
			return
		}
		req.Type = &t
	}

	// Status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s := UnitStatus(statusStr)
		if s != UnitStatusActive && s != UnitStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
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

	// Inventory type filter
	if invTypeStr := r.URL.Query().Get("inventory_type"); invTypeStr != "" {
		it := InventoryType(invTypeStr)
		if it != InventoryTypeTracked && it != InventoryTypeUntracked {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid inventory type"),
				logs.WithMetadata("inventory_type", invTypeStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid inventory type. Use 'tracked' or 'untracked'.")
			return
		}
		req.InventoryType = &it
	}

	// Tracking detail filter (НОВЫЙ!)
	if trackingDetailStr := r.URL.Query().Get("tracking_detail"); trackingDetailStr != "" {
		td := TrackingDetail(trackingDetailStr)
		if td != TrackingDetailBatch && td != TrackingDetailSerial {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid tracking detail"),
				logs.WithMetadata("tracking_detail", trackingDetailStr),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid tracking detail. Use 'batch' or 'serial'.")
			return
		}
		req.TrackingDetail = &td
	}

	// Category filter
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		req.CategoryID = &categoryID
	}

	// Search filter
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	units, total, err := h.repository.GetCatalogUnits(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get units: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"type":            req.Type,
			"status":          req.Status,
			"inventory_type":  req.InventoryType,
			"tracking_detail": req.TrackingDetail, // НОВЫЙ!
			"category_id":     req.CategoryID,
			"search":          req.Search,
		}),
		logs.WithMetadata("pagination", map[string]int{
			"page":  pagination.Page,
			"limit": pagination.Limit,
		}),
		logs.WithMetadata("result_count", len(units)),
	)

	response := map[string]interface{}{
		"units": units,
		"pagination": core.NewPagination(
			total,
			pagination.Page,
			pagination.Limit,
		),
	}

	core.SendSuccess(w, response, "Units retrieved successfully.")
}

func (h *Handlers) CreateCatalogUnit(w http.ResponseWriter, r *http.Request) {
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
	var req CreateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Проверяем обязательность категории
	if strings.TrimSpace(req.CategoryID) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category is required.")
		return
	}

	// Валидация
	if strings.TrimSpace(req.Name) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Unit name is required.")
		return
	}

	if req.Type == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit type is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Unit type is required.")
		return
	}

	if req.InventoryType == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Inventory type is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Inventory type is required.")
		return
	}

	if strings.TrimSpace(req.Unit) == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit of measurement is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Unit of measurement is required.")
		return
	}

	if req.SalePrice < 0 {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Sale price cannot be negative"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Sale price cannot be negative.")
		return
	}

	if req.Currency == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Currency is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Currency is required.")
		return
	}

	unit, err := h.repository.CreateCatalogUnit(r.Context(), req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "service cannot be tracked"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("type", req.Type),
				logs.WithMetadata("inventory_type", req.InventoryType),
			)
			core.SendValidationError(w, "Service cannot be tracked.")

		// НОВЫЕ ОШИБКИ ДЛЯ TRACKING_DETAIL
		case strings.Contains(errorMsg, "tracking detail (batch/serial) is required"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для складских товаров необходимо указать тип учета: batch (партионный) или serial (поштучный).")

		case strings.Contains(errorMsg, "tracked type (fifo/lifo) is required for batch tracking"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для партионного учета (batch) необходимо указать метод списания: FIFO или LIFO.")

		case strings.Contains(errorMsg, "tracked type must be nil for serial tracking"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для поштучного учета (serial) метод списания (FIFO/LIFO) не применяется.")

		case strings.Contains(errorMsg, "tracking detail must be nil for untracked items"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для товаров без учета (untracked) нельзя указывать детализацию учета.")

		case strings.Contains(errorMsg, "tracked type (fifo/lifo) is required"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Tracked type (FIFO/LIFO) is required for tracked items.")

		case strings.Contains(errorMsg, "purchase price is required"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Purchase price is required for tracked items.")

		case strings.Contains(errorMsg, "service cannot have purchase price"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Service cannot have purchase price.")

		case strings.Contains(errorMsg, "category with id") && strings.Contains(errorMsg, "not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("category_id", req.CategoryID),
			)
			core.SendValidationError(w, errorMsg)

		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to create unit: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("unit_id", unit.ID),
		logs.WithMetadata("name", req.Name),
		logs.WithMetadata("type", req.Type),
		logs.WithMetadata("inventory_type", req.InventoryType),
		logs.WithMetadata("tracking_detail", req.TrackingDetail), // НОВЫЙ!
	)

	core.SendSuccess(w, unit, "Unit created successfully.")
}

func (h *Handlers) UpdateCatalogUnit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID юнита из URL
	unitID := r.PathValue("unitId")
	if unitID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Unit ID is required.")
		return
	}

	// Парсим тело запроса
	var req UpdateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
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
		if *req.Type != UnitTypeProduct && *req.Type != UnitTypeService {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid unit type"),
				logs.WithMetadata("type", *req.Type),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid unit type. Use 'product' or 'service'.")
			return
		}
	}

	// Валидация статуса если указан
	if req.Status != nil {
		if *req.Status != UnitStatusActive && *req.Status != UnitStatusInactive {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
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

	// Валидация tracking_detail если указан (НОВЫЙ!)
	if req.TrackingDetail != nil {
		if *req.TrackingDetail != TrackingDetailBatch && *req.TrackingDetail != TrackingDetailSerial {
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid tracking detail"),
				logs.WithMetadata("tracking_detail", *req.TrackingDetail),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid tracking detail. Use 'batch' or 'serial'.")
			return
		}
	}

	unit, err := h.repository.UpdateCatalogUnit(r.Context(), unitID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "unit not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Unit not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("unit_id", unitID),
			)
			core.SendNotFound(w, "Unit not found.")

		case strings.Contains(errorMsg, "service cannot be tracked"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Service cannot be tracked.")

		// НОВЫЕ ОШИБКИ ДЛЯ TRACKING_DETAIL
		case strings.Contains(errorMsg, "tracking detail (batch/serial) is required"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для складских товаров необходимо указать тип учета: batch (партионный) или serial (поштучный).")

		case strings.Contains(errorMsg, "tracked type (fifo/lifo) is required for batch tracking"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для партионного учета (batch) необходимо указать метод списания: FIFO или LIFO.")

		case strings.Contains(errorMsg, "tracked type must be nil for serial tracking"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для поштучного учета (serial) метод списания (FIFO/LIFO) не применяется.")

		case strings.Contains(errorMsg, "tracking detail must be nil for untracked items"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Для товаров без учета (untracked) нельзя указывать детализацию учета.")

		case strings.Contains(errorMsg, "category with id") && strings.Contains(errorMsg, "not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("category_id", req.CategoryID),
			)
			core.SendValidationError(w, errorMsg)

		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update unit: %s", errorMsg))
		}
		return
	}

	// Формируем метаданные для лога с учетом новых полей
	logMetadata := map[string]interface{}{
		"unit_id": unitID,
	}
	if req.TrackingDetail != nil {
		logMetadata["tracking_detail"] = *req.TrackingDetail
	}
	if req.TrackedType != nil {
		logMetadata["tracked_type"] = *req.TrackedType
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadataMap(logMetadata),
	)

	core.SendSuccess(w, unit, "Unit updated successfully.")
}

func (h *Handlers) ActivateCatalogUnit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID юнита из URL
	unitID := r.PathValue("unitId")
	if unitID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Unit ID is required.")
		return
	}

	unit, err := h.repository.ActivateCatalogUnit(r.Context(), unitID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "unit not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Unit not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("unit_id", unitID),
			)
			core.SendNotFound(w, "Unit not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to activate unit: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("unit_id", unitID),
		logs.WithMetadata("action", "activate"),
	)

	core.SendSuccess(w, unit, "Unit activated successfully.")
}

func (h *Handlers) DeactivateCatalogUnit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		core.SendError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID юнита из URL
	unitID := r.PathValue("unitId")
	if unitID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Unit ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Unit ID is required.")
		return
	}

	unit, err := h.repository.DeactivateCatalogUnit(r.Context(), unitID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "unit not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Unit not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("unit_id", unitID),
			)
			core.SendNotFound(w, "Unit not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to deactivate unit: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_WM_CATALOG_UNITS_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("unit_id", unitID),
		logs.WithMetadata("action", "deactivate"),
	)

	core.SendSuccess(w, unit, "Unit deactivated successfully.")
}

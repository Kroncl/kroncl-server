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
// CATEGORIES
// базовый круд без хуйни
// --------

func (h *Handlers) GetCategory(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	category, err := h.repository.GetCategoryByID(r.Context(), categoryID)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("category_id", categoryID),
		)
		core.SendNotFound(w, "Category not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", categoryID),
	)

	core.SendSuccess(w, category, "Category retrieved successfully.")
}

func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Параметры пагинации
	pagination := core.GetDefaultPaginationParams(r)

	// Фильтры
	search := r.URL.Query().Get("search")

	var direction *TransactionCategoryDirection
	if dir := r.URL.Query().Get("direction"); dir != "" {
		d := TransactionCategoryDirection(dir)
		if d != TransactionCategoryDirectionIncome && d != TransactionCategoryDirectionExpense {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid direction"),
				logs.WithMetadata("direction", dir),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid direction. Use 'income' or 'expense'.")
			return
		}
		direction = &d
	}

	categories, total, err := h.repository.GetCategories(
		r.Context(),
		pagination.Offset,
		pagination.Limit,
		direction,
		search,
	)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to get categories: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("filters", map[string]interface{}{
			"direction": direction,
			"search":    search,
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

func (h *Handlers) CreateCategory(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Парсим тело запроса
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category name is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category name is required.")
		return
	}
	if req.Direction == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category direction is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category direction is required.")
		return
	}
	if req.Direction != TransactionCategoryDirectionIncome && req.Direction != TransactionCategoryDirectionExpense {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid direction"),
			logs.WithMetadata("direction", req.Direction),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendValidationError(w, "Invalid direction. Use 'income' or 'expense'.")
		return
	}

	category, err := h.repository.CreateCategory(r.Context(), req)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendInternalError(w, fmt.Sprintf("Failed to create category: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", category.ID),
		logs.WithMetadata("name", req.Name),
		logs.WithMetadata("direction", req.Direction),
	)

	core.SendSuccess(w, category, "Category created successfully.")
}

func (h *Handlers) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE, accountID,
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
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Invalid request body"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Валидация направления если указано
	if req.Direction != nil {
		if *req.Direction != TransactionCategoryDirectionIncome && *req.Direction != TransactionCategoryDirectionExpense {
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Invalid direction"),
				logs.WithMetadata("direction", *req.Direction),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendValidationError(w, "Invalid direction. Use 'income' or 'expense'.")
			return
		}
	}

	category, err := h.repository.UpdateCategory(r.Context(), categoryID, req)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "category not found"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", "Category not found"),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("category_id", categoryID),
			)
			core.SendNotFound(w, "Category not found.")
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to update category: %s", errorMsg))
		}
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_UPDATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", categoryID),
	)

	core.SendSuccess(w, category, "Category updated successfully.")
}

func (h *Handlers) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем ID категории из URL
	categoryID := r.PathValue("categoryId")
	if categoryID == "" {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category ID is required"),
			logs.WithMetadata("path", r.URL.Path),
		)
		core.SendError(w, http.StatusBadRequest, "Category ID is required.")
		return
	}

	ok, err := h.repository.DeleteCategory(r.Context(), categoryID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "cannot delete category: used in"):
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
				logs.WithMetadata("category_id", categoryID),
			)
			core.SendValidationError(w, errorMsg)
		default:
			h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE, accountID,
				logs.WithStatus(logs.LogStatusError),
				logs.WithUserAgent(r.UserAgent()),
				logs.WithMetadata("error", errorMsg),
				logs.WithMetadata("path", r.URL.Path),
			)
			core.SendInternalError(w, fmt.Sprintf("Failed to delete category: %s", errorMsg))
		}
		return
	}

	if !ok {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", "Category not found"),
			logs.WithMetadata("path", r.URL.Path),
			logs.WithMetadata("category_id", categoryID),
		)
		core.SendNotFound(w, "Category not found.")
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_CATEGORIES_DELETE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("category_id", categoryID),
	)

	core.SendSuccess(w, map[string]interface{}{
		"category_id": categoryID,
		"deleted":     true,
	}, "Category deleted successfully.")
}

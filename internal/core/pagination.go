package core

import (
	"net/http"
	"strconv"
)

// Pagination представляет метаданные пагинации
type Pagination struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

// PaginationParams содержит параметры пагинации из запроса
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// GetPaginationParams извлекает параметры пагинации из HTTP запроса
func GetPaginationParams(r *http.Request, defaultLimit, maxLimit int) PaginationParams {
	query := r.URL.Query()

	// Страница (по умолчанию 1)
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	// Лимит на страницу
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}

	// Вычисляем смещение
	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

// NewPagination создает структуру Pagination на основе результатов
func NewPagination(total, page, limit int) Pagination {
	pages := 0
	if limit > 0 {
		pages = (total + limit - 1) / limit // Округление вверх
	}
	if pages == 0 {
		pages = 1
	}

	return Pagination{
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	}
}

// GetDefaultPaginationParams удобная функция со стандартными настройками
func GetDefaultPaginationParams(r *http.Request) PaginationParams {
	return GetPaginationParams(r, 20, 100)
}

// ValidatePaginationParams валидация параметров пагинации
func ValidatePaginationParams(page, limit, maxLimit int) (int, int) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}

	return page, limit
}

package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// GetCategoryByID возвращает категорию по ID
func (r *Repository) GetCategoryByID(ctx context.Context, id string) (*TransactionCategory, error) {
	query := `
		SELECT 
			id, name, description, direction, system, slug, created_at, updated_at
		FROM transaction_categories
		WHERE id = $1
	`

	var category TransactionCategory
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Direction,
		&category.System,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// GetCategoryBySlug возвращает категорию по slug
func (r *Repository) GetCategoryBySlug(ctx context.Context, slug string) (*TransactionCategory, error) {
	query := `
		SELECT 
			id, name, description, direction, system, slug, created_at, updated_at
		FROM transaction_categories
		WHERE slug = $1
	`

	var category TransactionCategory
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Direction,
		&category.System,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return &category, nil
}

// GetCategories возвращает список категорий с пагинацией и фильтрацией
func (r *Repository) GetCategories(ctx context.Context, offset, limit int, direction *TransactionCategoryDirection, search string) ([]TransactionCategory, int, error) {
	var whereClause string
	var args []interface{}
	var whereConditions []string
	argIndex := 1

	if direction != nil {
		whereConditions = append(whereConditions, "direction = $"+strconv.Itoa(argIndex))
		args = append(args, *direction)
		argIndex++
	}

	if search != "" {
		searchConditions := []string{
			"name ILIKE $" + strconv.Itoa(argIndex),
			"description ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM transaction_categories ` + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	// Получаем категории с пагинацией
	query := `
		SELECT 
			id, name, description, direction, system, slug, created_at, updated_at
		FROM transaction_categories
		` + whereClause + `
		ORDER BY 
			system DESC, -- системные сверху
			name ASC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []TransactionCategory
	for rows.Next() {
		var category TransactionCategory
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Direction,
			&category.System,
			&category.Slug,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, total, nil
}

// CreateCategory создает новую категорию
func (r *Repository) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*TransactionCategory, error) {
	// Валидация
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	description := strings.TrimSpace(req.Description)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	var slug string
	id := uuid.New().String()
	slug = "category-" + id

	query := `
		INSERT INTO transaction_categories (
			id, name, description, direction, system, slug, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, name, description, direction, system, slug, created_at, updated_at
	`

	var category TransactionCategory
	err := r.pool.QueryRow(ctx, query,
		id,
		name,
		descriptionPtr,
		req.Direction,
		req.System,
		slug,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.Direction,
		&category.System,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

// UpdateCategory обновляет категорию
func (r *Repository) UpdateCategory(ctx context.Context, id string, req UpdateCategoryRequest) (*TransactionCategory, error) {
	// Проверяем существование и системный флаг
	existing, err := r.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Запрещаем изменение системных категорий
	if existing.System {
		return nil, fmt.Errorf("cannot update system category")
	}

	updater := core.NewUpdater("transaction_categories")

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name != "" {
			updater.SetString("name", name)
			// Сбрасываем slug при изменении имени (будет сгенерирован заново)
			updater.SetString("slug", "")
		}
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description == "" {
			updater.SetNull("description")
		} else {
			updater.SetString("description", description)
		}
	}

	if req.Direction != nil {
		updater.SetString("direction", string(*req.Direction))
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetCategoryByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return r.GetCategoryByID(ctx, id)
}

// DeleteCategory удаляет категорию
func (r *Repository) DeleteCategory(ctx context.Context, id string) (bool, error) {
	// Проверяем существование и системный флаг
	existing, err := r.GetCategoryByID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("category not found: %w", err)
	}

	// Запрещаем удаление системных категорий
	if existing.System {
		return false, fmt.Errorf("cannot delete system category")
	}

	// Проверяем, не используется ли категория в транзакциях
	checkQuery := `SELECT COUNT(*) FROM transaction_category WHERE category_id = $1`
	var count int
	err = r.pool.QueryRow(ctx, checkQuery, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check category usage: %w", err)
	}

	if count > 0 {
		return false, fmt.Errorf("cannot delete category: used in %d transactions", count)
	}

	query := `DELETE FROM transaction_categories WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return rowsAffected > 0, nil
}

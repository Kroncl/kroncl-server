package wm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CatalogCategoryExists проверяет существование категории по ID
func (r *Repository) CatalogCategoryExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM catalog_categories WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check catalog category existence: %w", err)
	}

	return exists, nil
}

// GetCatalogCategoryByID возвращает категорию по ID
func (r *Repository) GetCatalogCategoryByID(ctx context.Context, id string) (*CatalogCategory, error) {
	query := `
		SELECT 
			id, name, comment, status, parent_id, metadata, created_at, updated_at
		FROM catalog_categories
		WHERE id = $1
	`

	var category CatalogCategory
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Comment,
		&category.Status,
		&category.ParentID,
		&category.Metadata,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get catalog category: %w", err)
	}

	return &category, nil
}

// GetCatalogCategoryByName возвращает категорию по имени (для проверки уникальности)
func (r *Repository) GetCatalogCategoryByName(ctx context.Context, name string) (*CatalogCategory, error) {
	query := `
		SELECT 
			id, name, comment, status, parent_id, metadata, created_at, updated_at
		FROM catalog_categories
		WHERE name = $1
	`

	var category CatalogCategory
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&category.ID,
		&category.Name,
		&category.Comment,
		&category.Status,
		&category.ParentID,
		&category.Metadata,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get catalog category by name: %w", err)
	}

	return &category, nil
}

// GetCatalogCategories возвращает список категорий с пагинацией и фильтрацией
func (r *Repository) GetCatalogCategories(ctx context.Context, req GetCategoriesRequest) ([]CatalogCategory, int, error) {
	var whereClause string
	var args []interface{}
	var whereConditions []string
	argIndex := 1

	// Вычисляем offset
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	if req.Status != nil {
		whereConditions = append(whereConditions, "status = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.ParentID != nil {
		if *req.ParentID == "" || *req.ParentID == "null" {
			whereConditions = append(whereConditions, "parent_id IS NULL")
		} else {
			whereConditions = append(whereConditions, "parent_id = $"+strconv.Itoa(argIndex))
			args = append(args, *req.ParentID)
			argIndex++
		}
	}

	if req.Search != nil && *req.Search != "" {
		searchConditions := []string{
			"name ILIKE $" + strconv.Itoa(argIndex),
			"comment ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM catalog_categories ` + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count catalog categories: %w", err)
	}

	// Получаем категории с пагинацией
	query := `
		SELECT 
			id, name, comment, status, parent_id, metadata, created_at, updated_at
		FROM catalog_categories
		` + whereClause + `
		ORDER BY 
			name ASC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query catalog categories: %w", err)
	}
	defer rows.Close()

	var categories []CatalogCategory
	for rows.Next() {
		var category CatalogCategory
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Comment,
			&category.Status,
			&category.ParentID,
			&category.Metadata,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan catalog category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, total, nil
}

// CreateCatalogCategory создает новую категорию
func (r *Repository) CreateCatalogCategory(ctx context.Context, req CreateCategoryRequest) (*CatalogCategory, error) {
	// Валидация
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	// Проверяем уникальность имени
	existing, _ := r.GetCatalogCategoryByName(ctx, name)
	if existing != nil {
		return nil, fmt.Errorf("category with name '%s' already exists", name)
	}

	// Проверяем существование родительской категории, если указана
	if req.ParentID != nil && *req.ParentID != "" {
		parentExists, err := r.CatalogCategoryExists(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to check parent category existence: %w", err)
		}
		if !parentExists {
			return nil, fmt.Errorf("parent category with id '%s' not found", *req.ParentID)
		}
	}

	id := uuid.New().String()

	query := `
		INSERT INTO catalog_categories (
			id, name, comment, status, parent_id, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, name, comment, status, parent_id, metadata, created_at, updated_at
	`

	var category CatalogCategory
	err := r.pool.QueryRow(ctx, query,
		id,
		name,
		req.Comment,
		CategoryStatusActive, // новые категории всегда active по умолчанию
		req.ParentID,
		req.Metadata,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Comment,
		&category.Status,
		&category.ParentID,
		&category.Metadata,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create catalog category: %w", err)
	}

	return &category, nil
}

// UpdateCatalogCategory обновляет категорию
func (r *Repository) UpdateCatalogCategory(ctx context.Context, id string, req UpdateCategoryRequest) (*CatalogCategory, error) {
	// Проверяем существование
	existing, err := r.GetCatalogCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	updater := core.NewUpdater("catalog_categories")

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("category name cannot be empty")
		}

		// Проверяем уникальность имени, если оно меняется
		if name != existing.Name {
			existingWithName, _ := r.GetCatalogCategoryByName(ctx, name)
			if existingWithName != nil && existingWithName.ID != id {
				return nil, fmt.Errorf("category with name '%s' already exists", name)
			}
		}

		updater.SetString("name", name)
	}

	if req.Comment != nil {
		if *req.Comment == "" {
			updater.SetNull("comment")
		} else {
			comment := strings.TrimSpace(*req.Comment)
			updater.SetString("comment", comment)
		}
	}

	if req.ParentID != nil {
		if *req.ParentID == "" {
			updater.SetNull("parent_id")
		} else {
			// Проверяем существование родительской категории
			parentExists, err := r.CatalogCategoryExists(ctx, *req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("failed to check parent category existence: %w", err)
			}
			if !parentExists {
				return nil, fmt.Errorf("parent category with id '%s' not found", *req.ParentID)
			}
			// Проверяем, что не пытаемся сделать категорию родителем самой себя
			if *req.ParentID == id {
				return nil, fmt.Errorf("category cannot be parent of itself")
			}
			updater.SetString("parent_id", *req.ParentID)
		}
	}

	if req.Status != nil {
		updater.SetString("status", string(*req.Status))
	}

	if req.Metadata != nil {
		updater.SetJSONB("metadata", *req.Metadata)
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetCatalogCategoryByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update catalog category: %w", err)
	}

	return r.GetCatalogCategoryByID(ctx, id)
}

// ActivateCatalogCategory активирует категорию
func (r *Repository) ActivateCatalogCategory(ctx context.Context, id string) (*CatalogCategory, error) {
	// Проверяем существование
	exists, err := r.CatalogCategoryExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("category not found")
	}

	query := `
		UPDATE catalog_categories 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
		RETURNING 
			id, name, comment, status, parent_id, metadata, created_at, updated_at
	`

	var category CatalogCategory
	err = r.pool.QueryRow(ctx, query, CategoryStatusActive, id).Scan(
		&category.ID,
		&category.Name,
		&category.Comment,
		&category.Status,
		&category.ParentID,
		&category.Metadata,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to activate catalog category: %w", err)
	}

	return &category, nil
}

// DeactivateCatalogCategory деактивирует категорию
func (r *Repository) DeactivateCatalogCategory(ctx context.Context, id string) (*CatalogCategory, error) {
	// Проверяем существование
	exists, err := r.CatalogCategoryExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("category not found")
	}

	query := `
		UPDATE catalog_categories 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
		RETURNING 
			id, name, comment, status, parent_id, metadata, created_at, updated_at
	`

	var category CatalogCategory
	err = r.pool.QueryRow(ctx, query, CategoryStatusInactive, id).Scan(
		&category.ID,
		&category.Name,
		&category.Comment,
		&category.Status,
		&category.ParentID,
		&category.Metadata,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to deactivate catalog category: %w", err)
	}

	return &category, nil
}

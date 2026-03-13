package wm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ----------
// CATEGORIES
// ----------

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

// -----------
// UNITS
// -----------

// CatalogUnitExists проверяет существование юнита по ID
func (r *Repository) CatalogUnitExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM catalog_units WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check catalog unit existence: %w", err)
	}

	return exists, nil
}

// GetCatalogUnitByID возвращает юнит по ID
func (r *Repository) GetCatalogUnitByID(ctx context.Context, id string) (*CatalogUnit, error) {
	query := `
		SELECT 
			id, name, comment, type, status, inventory_type, tracking_detail, tracked_type, 
			unit, sale_price, purchase_price, currency, metadata, created_at, updated_at
		FROM catalog_units
		WHERE id = $1
	`

	var unit CatalogUnit
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Comment,
		&unit.Type,
		&unit.Status,
		&unit.InventoryType,
		&unit.TrackingDetail,
		&unit.TrackedType,
		&unit.Unit,
		&unit.SalePrice,
		&unit.PurchasePrice,
		&unit.Currency,
		&unit.Metadata,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get catalog unit: %w", err)
	}

	// Загружаем категорию из связующей таблицы
	categoryID, err := r.getUnitCategoryID(ctx, id)
	if err == nil {
		unit.CategoryID = categoryID
	}

	return &unit, nil
}

// getUnitCategoryID возвращает ID категории для юнита
func (r *Repository) getUnitCategoryID(ctx context.Context, unitID string) (string, error) {
	query := `SELECT category_id FROM catalog_unit_category WHERE unit_id = $1`

	var categoryID string
	err := r.pool.QueryRow(ctx, query, unitID).Scan(&categoryID)
	if err != nil {
		return "", fmt.Errorf("failed to get unit category: %w", err)
	}

	return categoryID, nil
}

// GetCatalogUnits возвращает список юнитов с пагинацией и фильтрацией
func (r *Repository) GetCatalogUnits(ctx context.Context, req GetUnitsRequest) ([]CatalogUnit, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	// Вычисляем offset
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	if req.Type != nil {
		conditions = append(conditions, "type = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Type)
		argIndex++
	}

	if req.Status != nil {
		conditions = append(conditions, "status = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.InventoryType != nil {
		conditions = append(conditions, "inventory_type = $"+strconv.Itoa(argIndex))
		args = append(args, *req.InventoryType)
		argIndex++
	}

	if req.TrackingDetail != nil {
		conditions = append(conditions, "tracking_detail = $"+strconv.Itoa(argIndex))
		args = append(args, *req.TrackingDetail)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		searchConditions := []string{
			"name ILIKE $" + strconv.Itoa(argIndex),
			"comment ILIKE $" + strconv.Itoa(argIndex),
		}
		conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	// Базовый запрос
	fromClause := `FROM catalog_units`

	// Если есть фильтр по категории, делаем JOIN с pivot таблицей
	if req.CategoryID != nil && *req.CategoryID != "" {
		fromClause += ` u INNER JOIN catalog_unit_category uc ON u.id = uc.unit_id`
		conditions = append(conditions, "uc.category_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.CategoryID)
		argIndex++
	} else {
		fromClause += ` u`
	}

	// Формируем WHERE
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(DISTINCT u.id) " + fromClause + " " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count catalog units: %w", err)
	}

	// Получаем юниты с пагинацией
	query := `
		SELECT DISTINCT
			u.id, u.name, u.comment, u.type, u.status, u.inventory_type, u.tracking_detail, u.tracked_type, 
			u.unit, u.sale_price, u.purchase_price, u.currency, u.metadata, u.created_at, u.updated_at
	` + fromClause + " " + whereClause + `
		ORDER BY u.name ASC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query catalog units: %w", err)
	}
	defer rows.Close()

	var units []CatalogUnit
	for rows.Next() {
		var unit CatalogUnit
		err := rows.Scan(
			&unit.ID,
			&unit.Name,
			&unit.Comment,
			&unit.Type,
			&unit.Status,
			&unit.InventoryType,
			&unit.TrackingDetail,
			&unit.TrackedType,
			&unit.Unit,
			&unit.SalePrice,
			&unit.PurchasePrice,
			&unit.Currency,
			&unit.Metadata,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan catalog unit: %w", err)
		}

		// Загружаем категорию для каждого юнита
		categoryID, err := r.getUnitCategoryID(ctx, unit.ID)
		if err == nil {
			unit.CategoryID = categoryID
		}

		units = append(units, unit)
	}

	return units, total, nil
}

// CreateCatalogUnit создает новый юнит
func (r *Repository) CreateCatalogUnit(ctx context.Context, req CreateUnitRequest) (*CatalogUnit, error) {
	// Валидация
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("unit name is required")
	}

	if req.SalePrice < 0 {
		return nil, fmt.Errorf("sale price cannot be negative")
	}

	// Проверяем обязательность категории
	if strings.TrimSpace(req.CategoryID) == "" {
		return nil, fmt.Errorf("category is required")
	}

	// Проверяем существование категории
	catExists, err := r.CatalogCategoryExists(ctx, req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if !catExists {
		return nil, fmt.Errorf("category with id '%s' not found", req.CategoryID)
	}

	// Проверяем ограничения для услуг
	if req.Type == UnitTypeService {
		if req.InventoryType == InventoryTypeTracked {
			return nil, fmt.Errorf("service cannot be tracked")
		}
		if req.PurchasePrice != nil {
			return nil, fmt.Errorf("service cannot have purchase price")
		}
		if req.TrackingDetail != nil {
			return nil, fmt.Errorf("service cannot have tracking detail")
		}
	}

	// Проверяем ограничения для tracked
	if req.InventoryType == InventoryTypeTracked {
		if req.TrackingDetail == nil {
			return nil, fmt.Errorf("tracking detail (batch/serial) is required for tracked items")
		}
		if req.PurchasePrice == nil {
			return nil, fmt.Errorf("purchase price is required for tracked items")
		}
		if *req.PurchasePrice < 0 {
			return nil, fmt.Errorf("purchase price cannot be negative")
		}

		// Для batch-учета требуется tracked_type (FIFO/LIFO)
		if *req.TrackingDetail == TrackingDetailBatch {
			if req.TrackedType == nil {
				return nil, fmt.Errorf("tracked type (fifo/lifo) is required for batch tracking")
			}
		} else {
			// Для serial-учета tracked_type должен быть nil
			if req.TrackedType != nil {
				return nil, fmt.Errorf("tracked type must be nil for serial tracking")
			}
		}
	} else {
		// Для untracked tracking_detail и tracked_type должны быть nil
		if req.TrackingDetail != nil {
			return nil, fmt.Errorf("tracking detail must be nil for untracked items")
		}
		if req.TrackedType != nil {
			return nil, fmt.Errorf("tracked type must be nil for untracked items")
		}
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Устанавливаем статус по умолчанию
	status := UnitStatusActive
	if req.Status != nil {
		status = *req.Status
	}

	id := uuid.New().String()

	// Вставляем юнит
	query := `
		INSERT INTO catalog_units (
			id, name, comment, type, status, inventory_type, tracking_detail, tracked_type, 
			unit, sale_price, purchase_price, currency, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, name, comment, type, status, inventory_type, tracking_detail, tracked_type, 
			unit, sale_price, purchase_price, currency, metadata, created_at, updated_at
	`

	var unit CatalogUnit
	err = tx.QueryRow(ctx, query,
		id,
		strings.TrimSpace(req.Name),
		req.Comment,
		req.Type,
		status,
		req.InventoryType,
		req.TrackingDetail,
		req.TrackedType,
		req.Unit,
		req.SalePrice,
		req.PurchasePrice,
		req.Currency,
		req.Metadata,
	).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Comment,
		&unit.Type,
		&unit.Status,
		&unit.InventoryType,
		&unit.TrackingDetail,
		&unit.TrackedType,
		&unit.Unit,
		&unit.SalePrice,
		&unit.PurchasePrice,
		&unit.Currency,
		&unit.Metadata,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create catalog unit: %w", err)
	}

	// Привязываем категорию
	linkQuery := `
		INSERT INTO catalog_unit_category (id, unit_id, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	linkID := uuid.New().String()
	_, err = tx.Exec(ctx, linkQuery, linkID, unit.ID, req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to link unit to category: %w", err)
	}
	unit.CategoryID = req.CategoryID

	// Коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &unit, nil
}

// UpdateCatalogUnit обновляет юнит
func (r *Repository) UpdateCatalogUnit(ctx context.Context, id string, req UpdateUnitRequest) (*CatalogUnit, error) {
	// Проверяем существование
	existing, err := r.GetCatalogUnitByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Валидация изменений
	if req.Type != nil && req.InventoryType != nil {
		if *req.Type == UnitTypeService && *req.InventoryType == InventoryTypeTracked {
			return nil, fmt.Errorf("service cannot be tracked")
		}
	}

	if req.PurchasePrice != nil && *req.PurchasePrice < 0 {
		return nil, fmt.Errorf("purchase price cannot be negative")
	}

	if req.SalePrice != nil && *req.SalePrice < 0 {
		return nil, fmt.Errorf("sale price cannot be negative")
	}

	// Определяем конечные значения для валидации
	finalType := existing.Type
	if req.Type != nil {
		finalType = *req.Type
	}

	finalInventoryType := existing.InventoryType
	if req.InventoryType != nil {
		finalInventoryType = *req.InventoryType
	}

	finalTrackingDetail := existing.TrackingDetail
	if req.TrackingDetail != nil {
		finalTrackingDetail = req.TrackingDetail
	}

	finalTrackedType := existing.TrackedType
	if req.TrackedType != nil {
		finalTrackedType = req.TrackedType
	}

	finalPurchasePrice := existing.PurchasePrice
	if req.PurchasePrice != nil {
		finalPurchasePrice = req.PurchasePrice
	}

	// Валидация для услуг
	if finalType == UnitTypeService {
		if finalInventoryType == InventoryTypeTracked {
			return nil, fmt.Errorf("service cannot be tracked")
		}
		if finalPurchasePrice != nil {
			return nil, fmt.Errorf("service cannot have purchase price")
		}
		if finalTrackingDetail != nil {
			return nil, fmt.Errorf("service cannot have tracking detail")
		}
	}

	// Валидация для tracked
	if finalInventoryType == InventoryTypeTracked {
		if finalTrackingDetail == nil {
			return nil, fmt.Errorf("tracking detail (batch/serial) is required for tracked items")
		}
		if finalPurchasePrice == nil {
			return nil, fmt.Errorf("purchase price is required for tracked items")
		}

		// Для batch-учета требуется tracked_type
		if *finalTrackingDetail == TrackingDetailBatch {
			if finalTrackedType == nil {
				return nil, fmt.Errorf("tracked type (fifo/lifo) is required for batch tracking")
			}
		} else {
			// Для serial-учета tracked_type должен быть nil
			if finalTrackedType != nil {
				return nil, fmt.Errorf("tracked type must be nil for serial tracking")
			}
		}
	} else {
		// Для untracked tracking_detail и tracked_type должны быть nil
		if finalTrackingDetail != nil {
			return nil, fmt.Errorf("tracking detail must be nil for untracked items")
		}
		if finalTrackedType != nil {
			return nil, fmt.Errorf("tracked type must be nil for untracked items")
		}
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем данные юнита
	updater := core.NewUpdater("catalog_units")

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("unit name cannot be empty")
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

	if req.Type != nil {
		updater.SetString("type", string(*req.Type))
	}

	if req.Status != nil {
		updater.SetString("status", string(*req.Status))
	}

	if req.InventoryType != nil {
		updater.SetString("inventory_type", string(*req.InventoryType))
	}

	if req.TrackingDetail != nil {
		updater.SetString("tracking_detail", string(*req.TrackingDetail))
	}

	if req.TrackedType != nil {
		updater.SetString("tracked_type", string(*req.TrackedType))
	}

	if req.Unit != nil {
		updater.SetString("unit", *req.Unit)
	}

	if req.SalePrice != nil {
		updater.SetFloat("sale_price", *req.SalePrice)
	}

	if req.PurchasePrice != nil {
		if *req.PurchasePrice == 0 {
			updater.SetNull("purchase_price")
		} else {
			updater.SetFloat("purchase_price", *req.PurchasePrice)
		}
	}

	if req.Currency != nil {
		updater.SetString("currency", string(*req.Currency))
	}

	if req.Metadata != nil {
		updater.SetJSONB("metadata", *req.Metadata)
	}

	// Применяем обновления
	unitQuery, unitArgs := updater.Where("id = $1", id).Build()
	if unitQuery != "" {
		_, err = tx.Exec(ctx, unitQuery, unitArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to update catalog unit: %w", err)
		}
	}

	// Обновляем категорию, если указана
	if req.CategoryID != nil {
		// Проверяем существование категории
		if *req.CategoryID != "" {
			catExists, err := r.CatalogCategoryExists(ctx, *req.CategoryID)
			if err != nil {
				return nil, fmt.Errorf("failed to check category existence: %w", err)
			}
			if !catExists {
				return nil, fmt.Errorf("category with id '%s' not found", *req.CategoryID)
			}
		}

		// Удаляем старую связь
		_, err = tx.Exec(ctx, "DELETE FROM catalog_unit_category WHERE unit_id = $1", id)
		if err != nil {
			return nil, fmt.Errorf("failed to remove old category link: %w", err)
		}

		// Если категория не пустая, создаем новую связь
		if *req.CategoryID != "" {
			linkQuery := `
				INSERT INTO catalog_unit_category (id, unit_id, category_id, created_at, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`
			linkID := uuid.New().String()
			_, err = tx.Exec(ctx, linkQuery, linkID, id, *req.CategoryID)
			if err != nil {
				return nil, fmt.Errorf("failed to link unit to category: %w", err)
			}
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetCatalogUnitByID(ctx, id)
}

// ActivateCatalogUnit активирует юнит
func (r *Repository) ActivateCatalogUnit(ctx context.Context, id string) (*CatalogUnit, error) {
	// Проверяем существование
	exists, err := r.CatalogUnitExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check unit existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("unit not found")
	}

	query := `
		UPDATE catalog_units 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
		RETURNING 
			id, name, comment, type, status, inventory_type, tracking_detail, tracked_type, 
			unit, sale_price, purchase_price, currency, metadata, created_at, updated_at
	`

	var unit CatalogUnit
	err = r.pool.QueryRow(ctx, query, UnitStatusActive, id).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Comment,
		&unit.Type,
		&unit.Status,
		&unit.InventoryType,
		&unit.TrackingDetail,
		&unit.TrackedType,
		&unit.Unit,
		&unit.SalePrice,
		&unit.PurchasePrice,
		&unit.Currency,
		&unit.Metadata,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to activate catalog unit: %w", err)
	}

	// Загружаем категорию
	categoryID, err := r.getUnitCategoryID(ctx, id)
	if err == nil {
		unit.CategoryID = categoryID
	}

	return &unit, nil
}

// DeactivateCatalogUnit деактивирует юнит
func (r *Repository) DeactivateCatalogUnit(ctx context.Context, id string) (*CatalogUnit, error) {
	// Проверяем существование
	exists, err := r.CatalogUnitExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check unit existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("unit not found")
	}

	query := `
		UPDATE catalog_units 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
		RETURNING 
			id, name, comment, type, status, inventory_type, tracking_detail, tracked_type, 
			unit, sale_price, purchase_price, currency, metadata, created_at, updated_at
	`

	var unit CatalogUnit
	err = r.pool.QueryRow(ctx, query, UnitStatusInactive, id).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Comment,
		&unit.Type,
		&unit.Status,
		&unit.InventoryType,
		&unit.TrackingDetail,
		&unit.TrackedType,
		&unit.Unit,
		&unit.SalePrice,
		&unit.PurchasePrice,
		&unit.Currency,
		&unit.Metadata,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to deactivate catalog unit: %w", err)
	}

	// Загружаем категорию
	categoryID, err := r.getUnitCategoryID(ctx, id)
	if err == nil {
		unit.CategoryID = categoryID
	}

	return &unit, nil
}

// GetCatalogUnitsByIDs возвращает список единиц каталога по их ID
func (r *Repository) GetCatalogUnitsByIDs(ctx context.Context, ids []string) ([]CatalogUnit, error) {
	if len(ids) == 0 {
		return []CatalogUnit{}, nil
	}

	// Создаем плейсхолдеры для IN запроса
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT 
			id, name, comment, type, status, inventory_type, tracking_detail, tracked_type, 
			unit, sale_price, purchase_price, currency, metadata, created_at, updated_at
		FROM catalog_units
		WHERE id IN (%s)
		ORDER BY name ASC
	`, strings.Join(placeholders, ", "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query catalog units by ids: %w", err)
	}
	defer rows.Close()

	var units []CatalogUnit
	for rows.Next() {
		var unit CatalogUnit
		err := rows.Scan(
			&unit.ID,
			&unit.Name,
			&unit.Comment,
			&unit.Type,
			&unit.Status,
			&unit.InventoryType,
			&unit.TrackingDetail,
			&unit.TrackedType,
			&unit.Unit,
			&unit.SalePrice,
			&unit.PurchasePrice,
			&unit.Currency,
			&unit.Metadata,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan catalog unit: %w", err)
		}

		// Загружаем категорию для каждого юнита (опционально)
		categoryID, err := r.getUnitCategoryID(ctx, unit.ID)
		if err == nil {
			unit.CategoryID = categoryID
		}

		units = append(units, unit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating catalog units: %w", err)
	}

	return units, nil
}

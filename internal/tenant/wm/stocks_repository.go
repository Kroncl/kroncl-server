package wm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Константы
const (
	// MaxSerialPositionsPerBatch максимальное количество serial-позиций в одном батче
	MaxSerialPositionsPerBatch = 400
)

// --------
// ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ
// --------

// StockBatchExists проверяет существование батча по ID
func (r *Repository) StockBatchExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM stock_batches WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check stock batch existence: %w", err)
	}

	return exists, nil
}

// StockPositionExists проверяет существование позиции по ID
func (r *Repository) StockPositionExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM stock_positions WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check stock position existence: %w", err)
	}

	return exists, nil
}

// getUnitTrackingDetail возвращает tracking_detail юнита
func (r *Repository) getUnitTrackingDetail(ctx context.Context, unitID string) (*TrackingDetail, error) {
	query := `SELECT tracking_detail FROM catalog_units WHERE id = $1`

	var trackingDetail *TrackingDetail
	err := r.pool.QueryRow(ctx, query, unitID).Scan(&trackingDetail)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit tracking detail: %w", err)
	}

	return trackingDetail, nil
}

// getUnitByID возвращает юнит по ID
func (r *Repository) getUnitByID(ctx context.Context, unitID string) (*CatalogUnit, error) {
	return r.GetCatalogUnitByID(ctx, unitID)
}

// --------
// STOCK BATCHES
// --------

// GetStockBatchByID возвращает батч по ID
func (r *Repository) GetStockBatchByID(ctx context.Context, id string) (*StockBatch, error) {
	query := `
		SELECT id, direction, comment, metadata, created_at, updated_at
		FROM stock_batches
		WHERE id = $1
	`

	var batch StockBatch
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&batch.ID,
		&batch.Direction,
		&batch.Comment,
		&batch.Metadata,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stock batch: %w", err)
	}

	return &batch, nil
}

// GetStockBatches возвращает список батчей с пагинацией и фильтрацией
func (r *Repository) GetStockBatches(ctx context.Context, req GetStockBatchesParams) ([]StockBatch, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	if req.Direction != nil {
		conditions = append(conditions, "direction = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Direction)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		conditions = append(conditions, "comment ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	// Если есть фильтр по unit_id, делаем JOIN с позициями
	if req.UnitID != nil && *req.UnitID != "" {
		fromClause := `FROM stock_batches sb
			INNER JOIN stock_position_batch spb ON sb.id = spb.batch_id
			INNER JOIN stock_positions sp ON spb.position_id = sp.id`
		conditions = append(conditions, "sp.unit_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.UnitID)
		argIndex++

		// Получаем общее количество
		countQuery := "SELECT COUNT(DISTINCT sb.id) " + fromClause + " WHERE " + strings.Join(conditions, " AND ")
		var total int
		err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count stock batches: %w", err)
		}

		// Получаем батчи с пагинацией
		query := `
			SELECT DISTINCT sb.id, sb.direction, sb.comment, sb.metadata, sb.created_at, sb.updated_at
		` + fromClause + " WHERE " + strings.Join(conditions, " AND ") + `
			ORDER BY sb.created_at DESC
			LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

		allArgs := append(args, req.Limit, offset)

		rows, err := r.pool.Query(ctx, query, allArgs...)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query stock batches: %w", err)
		}
		defer rows.Close()

		var batches []StockBatch
		for rows.Next() {
			var batch StockBatch
			err := rows.Scan(
				&batch.ID,
				&batch.Direction,
				&batch.Comment,
				&batch.Metadata,
				&batch.CreatedAt,
				&batch.UpdatedAt,
			)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to scan stock batch: %w", err)
			}
			batches = append(batches, batch)
		}

		return batches, total, nil
	}

	// Без фильтра по unit_id
	fromClause := `FROM stock_batches`
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(*) " + fromClause + " " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count stock batches: %w", err)
	}

	// Получаем батчи с пагинацией
	query := `
		SELECT id, direction, comment, metadata, created_at, updated_at
	` + fromClause + " " + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query stock batches: %w", err)
	}
	defer rows.Close()

	var batches []StockBatch
	for rows.Next() {
		var batch StockBatch
		err := rows.Scan(
			&batch.ID,
			&batch.Direction,
			&batch.Comment,
			&batch.Metadata,
			&batch.CreatedAt,
			&batch.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan stock batch: %w", err)
		}
		batches = append(batches, batch)
	}

	return batches, total, nil
}

// GetStockBatchWithPositions возвращает батч со всеми позициями (монолитно)
func (r *Repository) GetStockBatchWithPositions(ctx context.Context, batchID string) (*BatchWithPositionsResponse, error) {
	// Получаем батч
	batch, err := r.GetStockBatchByID(ctx, batchID)
	if err != nil {
		return nil, err
	}

	// Получаем позиции в батче с деталями юнитов
	query := `
		SELECT 
			sp.id, sp.type, sp.unit_id, sp.quantity, sp.created_at,
			u.id, u.name, u.comment, u.type, u.status, u.inventory_type, 
			u.tracking_detail, u.tracked_type, u.unit, u.sale_price, 
			u.purchase_price, u.currency, u.metadata, u.created_at, u.updated_at
		FROM stock_positions sp
		INNER JOIN stock_position_batch spb ON sp.id = spb.position_id
		INNER JOIN catalog_units u ON sp.unit_id = u.id
		WHERE spb.batch_id = $1
		ORDER BY sp.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query batch positions: %w", err)
	}
	defer rows.Close()

	var positions []PositionWithUnitResponse
	for rows.Next() {
		var pos PositionWithUnitResponse
		var unit CatalogUnit

		err := rows.Scan(
			&pos.ID,
			&pos.Type,
			&pos.UnitID,
			&pos.Quantity,
			&pos.CreatedAt,
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
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}

		pos.BatchID = batchID
		pos.Unit = unit
		positions = append(positions, pos)
	}

	return &BatchWithPositionsResponse{
		ID:        batch.ID,
		Direction: batch.Direction,
		Comment:   batch.Comment,
		Metadata:  batch.Metadata,
		CreatedAt: batch.CreatedAt,
		UpdatedAt: batch.UpdatedAt,
		Positions: positions,
	}, nil
}

// --------
// STOCK POSITIONS (только просмотр)
// --------

// GetStockPositionByID возвращает позицию по ID (без деталей)
func (r *Repository) GetStockPositionByID(ctx context.Context, id string) (*StockPosition, error) {
	query := `
		SELECT id, type, unit_id, quantity, created_at
		FROM stock_positions
		WHERE id = $1
	`

	var pos StockPosition
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pos.ID,
		&pos.Type,
		&pos.UnitID,
		&pos.Quantity,
		&pos.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stock position: %w", err)
	}

	return &pos, nil
}

// GetStockPositionWithDetails возвращает позицию с деталями (монолитно)
func (r *Repository) GetStockPositionWithDetails(ctx context.Context, id string) (*PositionWithUnitResponse, error) {
	query := `
		SELECT 
			sp.id, sp.type, sp.unit_id, sp.quantity, sp.created_at,
			u.id, u.name, u.comment, u.type, u.status, u.inventory_type, 
			u.tracking_detail, u.tracked_type, u.unit, u.sale_price, 
			u.purchase_price, u.currency, u.metadata, u.created_at, u.updated_at,
			spb.batch_id
		FROM stock_positions sp
		INNER JOIN catalog_units u ON sp.unit_id = u.id
		LEFT JOIN stock_position_batch spb ON sp.id = spb.position_id
		WHERE sp.id = $1
	`

	var pos PositionWithUnitResponse
	var unit CatalogUnit
	var batchID *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pos.ID,
		&pos.Type,
		&pos.UnitID,
		&pos.Quantity,
		&pos.CreatedAt,
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
		&batchID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stock position with details: %w", err)
	}

	pos.Unit = unit
	if batchID != nil {
		pos.BatchID = *batchID
	}

	return &pos, nil
}

// GetStockPositions возвращает список позиций с пагинацией и фильтрацией (монолитно)
func (r *Repository) GetStockPositions(ctx context.Context, req GetStockPositionsParams) ([]PositionWithUnitResponse, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	// Базовый запрос
	fromClause := `FROM stock_positions sp
		INNER JOIN catalog_units u ON sp.unit_id = u.id
		LEFT JOIN stock_position_batch spb ON sp.id = spb.position_id`

	if req.Type != nil {
		conditions = append(conditions, "sp.type = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Type)
		argIndex++
	}

	if req.UnitID != nil {
		conditions = append(conditions, "sp.unit_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.UnitID)
		argIndex++
	}

	if req.BatchID != nil {
		conditions = append(conditions, "spb.batch_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.BatchID)
		argIndex++
	}

	if req.InStock != nil && *req.InStock {
		conditions = append(conditions, "sp.quantity > 0")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(DISTINCT sp.id) " + fromClause + " " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count stock positions: %w", err)
	}

	// Получаем позиции с пагинацией - ИСПРАВЛЕНО: batch_id как *string
	query := `
		SELECT DISTINCT
			sp.id, sp.type, sp.unit_id, sp.quantity, sp.created_at,
			u.id, u.name, u.comment, u.type, u.status, u.inventory_type, 
			u.tracking_detail, u.tracked_type, u.unit, u.sale_price, 
			u.purchase_price, u.currency, u.metadata, u.created_at, u.updated_at,
			spb.batch_id
	` + fromClause + " " + whereClause + `
		ORDER BY sp.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query stock positions: %w", err)
	}
	defer rows.Close()

	var positions []PositionWithUnitResponse
	for rows.Next() {
		var pos PositionWithUnitResponse
		var unit CatalogUnit
		var batchID *string // ИСПРАВЛЕНО: используем указатель

		err := rows.Scan(
			&pos.ID,
			&pos.Type,
			&pos.UnitID,
			&pos.Quantity,
			&pos.CreatedAt,
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
			&batchID, // ИСПРАВЛЕНО: сканируем в указатель
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan stock position: %w", err)
		}

		pos.Unit = unit
		if batchID != nil {
			pos.BatchID = *batchID // ИСПРАВЛЕНО: разыменовываем если не nil
		}
		positions = append(positions, pos)
	}

	return positions, total, nil
}

// --------
// STOCK BATCHES WITH POSITIONS (ОСНОВНОЙ МЕТОД)
// --------

// CreateStockBatchWithPositions создает батч с позициями (атомарно) и возвращает монолитный ответ
func (r *Repository) CreateStockBatchWithPositions(ctx context.Context, req CreateStockBatchRequest) (*CreateStockBatchResponse, error) {
	// Валидация
	if len(req.Positions) == 0 {
		return nil, fmt.Errorf("at least one position is required")
	}

	// Подсчет serial позиций
	serialCount := 0
	for _, pos := range req.Positions {
		trackingDetail, err := r.getUnitTrackingDetail(ctx, pos.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tracking detail for unit %s: %w", pos.UnitID, err)
		}

		if trackingDetail != nil && *trackingDetail == TrackingDetailSerial {
			// Для serial считаем каждую единицу отдельно
			serialCount += int(pos.Quantity)
		}
	}

	// Проверка лимита для serial позиций
	if serialCount > MaxSerialPositionsPerBatch {
		return nil, fmt.Errorf("too many serial positions in one batch: %d (max %d)",
			serialCount, MaxSerialPositionsPerBatch)
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Создаем батч
	batchID := uuid.New().String()
	batchQuery := `
		INSERT INTO stock_batches (id, direction, comment, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, direction, comment, metadata, created_at, updated_at
	`

	var batch StockBatch
	err = tx.QueryRow(ctx, batchQuery,
		batchID,
		req.Direction,
		req.Comment,
		req.Metadata,
	).Scan(
		&batch.ID,
		&batch.Direction,
		&batch.Comment,
		&batch.Metadata,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create stock batch: %w", err)
	}

	// 2. Создаем позиции
	var positions []PositionWithUnitResponse

	for _, posReq := range req.Positions {
		// Проверяем существование юнита
		unitExists, err := r.CatalogUnitExists(ctx, posReq.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to check unit existence: %w", err)
		}
		if !unitExists {
			return nil, fmt.Errorf("unit with id '%s' not found", posReq.UnitID)
		}

		// Получаем юнит для ответа
		unit, err := r.getUnitByID(ctx, posReq.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get unit details: %w", err)
		}

		// Получаем tracking_detail юнита
		trackingDetail, err := r.getUnitTrackingDetail(ctx, posReq.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tracking detail for unit %s: %w", posReq.UnitID, err)
		}

		// Определяем тип позиции
		posType := StockPositionTypeBatch
		if trackingDetail != nil && *trackingDetail == TrackingDetailSerial {
			posType = StockPositionTypeSerial
		}

		// Для serial позиций количество должно быть целым и создаем отдельные записи
		if posType == StockPositionTypeSerial {
			if posReq.Quantity != float64(int(posReq.Quantity)) {
				return nil, fmt.Errorf("serial position quantity must be integer")
			}

			// Создаем отдельную запись для каждой единицы
			for i := 0; i < int(posReq.Quantity); i++ {
				posID := uuid.New().String()
				posQuery := `
					INSERT INTO stock_positions (id, type, unit_id, quantity, created_at)
					VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
					RETURNING id, type, unit_id, quantity, created_at
				`

				var stockPos StockPosition
				err = tx.QueryRow(ctx, posQuery,
					posID,
					posType,
					posReq.UnitID,
					1, // всегда 1 для serial
				).Scan(
					&stockPos.ID,
					&stockPos.Type,
					&stockPos.UnitID,
					&stockPos.Quantity,
					&stockPos.CreatedAt,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to create serial position: %w", err)
				}

				// Создаем связь
				linkQuery := `
					INSERT INTO stock_position_batch (id, position_id, batch_id, created_at)
					VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
				`
				linkID := uuid.New().String()
				_, err = tx.Exec(ctx, linkQuery, linkID, stockPos.ID, batch.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to link position to batch: %w", err)
				}

				// Добавляем в ответ
				positions = append(positions, PositionWithUnitResponse{
					ID:        stockPos.ID,
					Type:      stockPos.Type,
					UnitID:    stockPos.UnitID,
					Quantity:  stockPos.Quantity,
					CreatedAt: stockPos.CreatedAt,
					BatchID:   batch.ID,
					Unit:      *unit,
				})
			}
		} else {
			// Batch позиция - одна запись
			posID := uuid.New().String()
			posQuery := `
				INSERT INTO stock_positions (id, type, unit_id, quantity, created_at)
				VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
				RETURNING id, type, unit_id, quantity, created_at
			`

			var stockPos StockPosition
			err = tx.QueryRow(ctx, posQuery,
				posID,
				posType,
				posReq.UnitID,
				posReq.Quantity,
			).Scan(
				&stockPos.ID,
				&stockPos.Type,
				&stockPos.UnitID,
				&stockPos.Quantity,
				&stockPos.CreatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create batch position: %w", err)
			}

			// Создаем связь
			linkQuery := `
				INSERT INTO stock_position_batch (id, position_id, batch_id, created_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
			`
			linkID := uuid.New().String()
			_, err = tx.Exec(ctx, linkQuery, linkID, stockPos.ID, batch.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to link position to batch: %w", err)
			}

			// Добавляем в ответ
			positions = append(positions, PositionWithUnitResponse{
				ID:        stockPos.ID,
				Type:      stockPos.Type,
				UnitID:    stockPos.UnitID,
				Quantity:  stockPos.Quantity,
				CreatedAt: stockPos.CreatedAt,
				BatchID:   batch.ID,
				Unit:      *unit,
			})
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CreateStockBatchResponse{
		BatchID:   batch.ID,
		Direction: batch.Direction,
		Comment:   batch.Comment,
		Metadata:  batch.Metadata,
		CreatedAt: batch.CreatedAt,
		UpdatedAt: batch.UpdatedAt,
		Positions: positions,
	}, nil
}

// GetStockPositionsByIDs возвращает список складских позиций по их ID (без информации о юнитах)
func (r *Repository) GetStockPositionsByIDs(ctx context.Context, ids []string) ([]StockPosition, error) {
	if len(ids) == 0 {
		return []StockPosition{}, nil
	}

	// Создаем плейсхолдеры для IN запроса
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, type, unit_id, quantity, created_at
		FROM stock_positions
		WHERE id IN (%s)
		ORDER BY created_at DESC
	`, strings.Join(placeholders, ", "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock positions by ids: %w", err)
	}
	defer rows.Close()

	var positions []StockPosition
	for rows.Next() {
		var pos StockPosition
		err := rows.Scan(
			&pos.ID,
			&pos.Type,
			&pos.UnitID,
			&pos.Quantity,
			&pos.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock position: %w", err)
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock positions: %w", err)
	}

	return positions, nil
}

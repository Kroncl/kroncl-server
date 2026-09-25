package wm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// MovementExists проверяет существование движения
func (r *Repository) MovementExists(ctx context.Context, outcomeBatchID, positionID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM stock_position_movements
			WHERE outcome_batch_id = $1 AND position_id = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, outcomeBatchID, positionID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check movement existence: %w", err)
	}
	return exists, nil
}

// GetMovementsByPosition возвращает движения по позиции
func (r *Repository) GetMovementsByPosition(ctx context.Context, positionID string) ([]StockPositionMovement, error) {
	query := `
		SELECT outcome_batch_id, position_id, type, quantity, comment, metadata, created_at
		FROM stock_position_movements
		WHERE position_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query movements: %w", err)
	}
	defer rows.Close()

	var movements []StockPositionMovement
	for rows.Next() {
		var m StockPositionMovement
		err := rows.Scan(
			&m.OutcomeBatchID,
			&m.PositionID,
			&m.Type,
			&m.Quantity,
			&m.Comment,
			&m.Metadata,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movement: %w", err)
		}
		movements = append(movements, m)
	}

	return movements, nil
}

// GetMovementsByOutcomeBatch возвращает движения по батчу отгрузки
func (r *Repository) GetMovementsByOutcomeBatch(ctx context.Context, batchID string) ([]StockPositionMovement, error) {
	query := `
		SELECT outcome_batch_id, position_id, type, quantity, comment, metadata, created_at
		FROM stock_position_movements
		WHERE outcome_batch_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query movements by batch: %w", err)
	}
	defer rows.Close()

	var movements []StockPositionMovement
	for rows.Next() {
		var m StockPositionMovement
		err := rows.Scan(
			&m.OutcomeBatchID,
			&m.PositionID,
			&m.Type,
			&m.Quantity,
			&m.Comment,
			&m.Metadata,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movement: %w", err)
		}
		movements = append(movements, m)
	}

	return movements, nil
}

// CreateMovement создаёт движение (списание) с проверкой остатка
func (r *Repository) CreateMovement(ctx context.Context, req CreateStockMovementRequest) (*StockPositionMovement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Блокируем позицию
	var positionQuantity float64
	err = tx.QueryRow(ctx, `
		SELECT quantity FROM stock_positions WHERE id = $1 FOR UPDATE
	`, req.PositionID).Scan(&positionQuantity)
	if err != nil {
		return nil, fmt.Errorf("failed to lock position: %w", err)
	}

	// Считаем уже списанное
	var alreadyWrittenOff float64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity), 0)
		FROM stock_position_movements
		WHERE position_id = $1 AND type = 'write_off'
	`, req.PositionID).Scan(&alreadyWrittenOff)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate written off: %w", err)
	}

	remaining := positionQuantity - alreadyWrittenOff
	if req.Quantity > remaining {
		return nil, fmt.Errorf("insufficient stock: requested %v, remaining %v", req.Quantity, remaining)
	}

	// Вставляем движение
	var m StockPositionMovement
	err = tx.QueryRow(ctx, `
		INSERT INTO stock_position_movements (
			outcome_batch_id, position_id, type, quantity, comment, metadata, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		RETURNING outcome_batch_id, position_id, type, quantity, comment, metadata, created_at
	`,
		req.OutcomeBatchID,
		req.PositionID,
		req.Type,
		req.Quantity,
		req.Comment,
		req.Metadata,
	).Scan(
		&m.OutcomeBatchID,
		&m.PositionID,
		&m.Type,
		&m.Quantity,
		&m.Comment,
		&m.Metadata,
		&m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create movement: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &m, nil
}

// CreateMovementsBatch создаёт несколько движений атомарно (для отгрузки)
func (r *Repository) CreateMovementsBatch(ctx context.Context, requests []CreateStockMovementRequest) ([]StockPositionMovement, error) {
	if len(requests) == 0 {
		return []StockPositionMovement{}, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var movements []StockPositionMovement

	for _, req := range requests {
		// Блокируем позицию
		var positionQuantity float64
		err = tx.QueryRow(ctx, `
			SELECT quantity FROM stock_positions WHERE id = $1 FOR UPDATE
		`, req.PositionID).Scan(&positionQuantity)
		if err != nil {
			return nil, fmt.Errorf("failed to lock position %s: %w", req.PositionID, err)
		}

		// Считаем уже списанное
		var alreadyWrittenOff float64
		err = tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(quantity), 0)
			FROM stock_position_movements
			WHERE position_id = $1 AND type = 'write_off'
		`, req.PositionID).Scan(&alreadyWrittenOff)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate written off: %w", err)
		}

		remaining := positionQuantity - alreadyWrittenOff
		if req.Quantity > remaining {
			return nil, fmt.Errorf("insufficient stock for position %s: requested %v, remaining %v",
				req.PositionID, req.Quantity, remaining)
		}

		// Вставляем
		var m StockPositionMovement
		err = tx.QueryRow(ctx, `
			INSERT INTO stock_position_movements (
				outcome_batch_id, position_id, type, quantity, comment, metadata, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
			RETURNING outcome_batch_id, position_id, type, quantity, comment, metadata, created_at
		`,
			req.OutcomeBatchID,
			req.PositionID,
			req.Type,
			req.Quantity,
			req.Comment,
			req.Metadata,
		).Scan(
			&m.OutcomeBatchID,
			&m.PositionID,
			&m.Type,
			&m.Quantity,
			&m.Comment,
			&m.Metadata,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create movement for position %s: %w", req.PositionID, err)
		}

		movements = append(movements, m)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return movements, nil
}

// GetStockBalance возвращает остатки по всем позициям (агрегированно по unit_id)
func (r *Repository) GetStockBalance(ctx context.Context, unitID *string) ([]StockBalanceItem, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	if unitID != nil && *unitID != "" {
		conditions = append(conditions, "sp.unit_id = $"+strconv.Itoa(argIndex))
		args = append(args, *unitID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT
			sp.unit_id,
			u.name AS unit_name,
			COALESCE(SUM(sp.quantity), 0) AS total_quantity,
			0 AS reserved,
			COALESCE(SUM(
				sp.quantity - COALESCE((
					SELECT SUM(m.quantity)
					FROM stock_position_movements m
					WHERE m.position_id = sp.id AND m.type = 'write_off'
				), 0)
			), 0) AS available
		FROM stock_positions sp
		INNER JOIN catalog_units u ON sp.unit_id = u.id
		` + whereClause + `
		GROUP BY sp.unit_id, u.name
		ORDER BY u.name ASC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock balance: %w", err)
	}
	defer rows.Close()

	var items []StockBalanceItem
	for rows.Next() {
		var item StockBalanceItem
		err := rows.Scan(
			&item.UnitID,
			&item.UnitName,
			&item.Quantity,
			&item.Reserved,
			&item.Available,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock balance: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) GetMovements(ctx context.Context, req GetMovementsParams) ([]StockPositionMovement, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	if req.Type != nil {
		conditions = append(conditions, "type = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Type)
		argIndex++
	}
	if req.OutcomeBatchID != nil {
		conditions = append(conditions, "outcome_batch_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.OutcomeBatchID)
		argIndex++
	}
	if req.PositionID != nil {
		conditions = append(conditions, "position_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.PositionID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM stock_position_movements " + whereClause
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count movements: %w", err)
	}

	query := `
		SELECT outcome_batch_id, position_id, type, quantity, comment, metadata, created_at
		FROM stock_position_movements
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query movements: %w", err)
	}
	defer rows.Close()

	var movements []StockPositionMovement
	for rows.Next() {
		var m StockPositionMovement
		if err := rows.Scan(
			&m.OutcomeBatchID, &m.PositionID, &m.Type, &m.Quantity,
			&m.Comment, &m.Metadata, &m.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan movement: %w", err)
		}
		movements = append(movements, m)
	}

	return movements, total, nil
}

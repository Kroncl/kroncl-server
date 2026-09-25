package wm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// StockPositionExists проверяет существование позиции
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

// positionScanFields — единый набор полей для SELECT-сканов
const positionSelectFields = `
	sp.id, sp.type, sp.income_batch_id, sp.unit_id, sp.quantity, sp.unit_price,
	sp.maker, sp.barcode_id, sp.created_at, sp.updated_at,
	u.id, u.name, u.comment, u.type, u.status, u.inventory_type,
	u.tracking_detail, u.tracked_type, u.unit, u.sale_price,
	u.purchase_price, u.currency, u.metadata, u.created_at, u.updated_at
`

// scanPositionWithUnit — хелпер для сканирования
func scanPositionWithUnit(rows interface {
	Scan(dest ...interface{}) error
}) (*PositionWithUnitResponse, error) {
	var pos PositionWithUnitResponse
	var unit CatalogUnit

	err := rows.Scan(
		&pos.ID, &pos.Type, &pos.IncomeBatchID, &pos.UnitID, &pos.Quantity, &pos.UnitPrice,
		&pos.Maker, &pos.BarcodeID, &pos.CreatedAt, &pos.UpdatedAt,
		&unit.ID, &unit.Name, &unit.Comment, &unit.Type, &unit.Status, &unit.InventoryType,
		&unit.TrackingDetail, &unit.TrackedType, &unit.Unit, &unit.SalePrice,
		&unit.PurchasePrice, &unit.Currency, &unit.Metadata, &unit.CreatedAt, &unit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	pos.Unit = unit
	return &pos, nil
}

// GetStockPositionByID возвращает позицию без деталей
func (r *Repository) GetStockPositionByID(ctx context.Context, id string) (*StockPosition, error) {
	query := `
		SELECT id, type, income_batch_id, unit_id, quantity, unit_price, maker, barcode_id, created_at, updated_at
		FROM stock_positions
		WHERE id = $1
	`

	var pos StockPosition
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pos.ID, &pos.Type, &pos.IncomeBatchID, &pos.UnitID, &pos.Quantity, &pos.UnitPrice,
		&pos.Maker, &pos.BarcodeID, &pos.CreatedAt, &pos.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock position: %w", err)
	}
	return &pos, nil
}

// GetStockPositionWithDetails возвращает позицию с деталями
func (r *Repository) GetStockPositionWithDetails(ctx context.Context, id string) (*PositionWithUnitResponse, error) {
	query := `
		SELECT ` + positionSelectFields + `
		FROM stock_positions sp
		INNER JOIN catalog_units u ON sp.unit_id = u.id
		WHERE sp.id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	pos, err := scanPositionWithUnit(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock position with details: %w", err)
	}

	// Считаем остаток
	remaining, err := r.GetPositionRemaining(ctx, id)
	if err == nil {
		pos.Remaining = remaining
	} else {
		pos.Remaining = pos.Quantity
	}

	return pos, nil
}

// GetStockPositions возвращает список позиций с пагинацией
func (r *Repository) GetStockPositions(ctx context.Context, req GetStockPositionsParams) ([]PositionWithUnitResponse, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	fromClause := `FROM stock_positions sp
		INNER JOIN catalog_units u ON sp.unit_id = u.id`

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

	if req.IncomeBatchID != nil {
		conditions = append(conditions, "sp.income_batch_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IncomeBatchID)
		argIndex++
	}

	if req.InStock != nil && *req.InStock {
		// позиция в наличии, если quantity > сумма движений
		conditions = append(conditions, `
			sp.quantity > COALESCE((
				SELECT SUM(m.quantity)
				FROM stock_position_movements m
				WHERE m.position_id = sp.id AND m.type = 'write_off'
			), 0)
		`)
	}

	if req.Search != nil && *req.Search != "" {
		searchPattern := "%" + strings.ToLower(*req.Search) + "%"
		conditions = append(conditions,
			"(sp.id::text ILIKE $"+strconv.Itoa(argIndex)+" OR LOWER(u.name) ILIKE $"+strconv.Itoa(argIndex)+")")
		args = append(args, searchPattern)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) " + fromClause + " " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count stock positions: %w", err)
	}

	query := `
		SELECT ` + positionSelectFields + `
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
		pos, err := scanPositionWithUnit(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan stock position: %w", err)
		}

		remaining, err := r.GetPositionRemaining(ctx, pos.ID)
		if err == nil {
			pos.Remaining = remaining
		} else {
			pos.Remaining = pos.Quantity
		}

		positions = append(positions, *pos)
	}

	return positions, total, nil
}

// GetPositionsByIncomeBatch возвращает все позиции батча
func (r *Repository) GetPositionsByIncomeBatch(ctx context.Context, batchID string) ([]PositionWithUnitResponse, error) {
	query := `
		SELECT ` + positionSelectFields + `
		FROM stock_positions sp
		INNER JOIN catalog_units u ON sp.unit_id = u.id
		WHERE sp.income_batch_id = $1
		ORDER BY sp.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query batch positions: %w", err)
	}
	defer rows.Close()

	var positions []PositionWithUnitResponse
	for rows.Next() {
		pos, err := scanPositionWithUnit(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}

		remaining, err := r.GetPositionRemaining(ctx, pos.ID)
		if err == nil {
			pos.Remaining = remaining
		} else {
			pos.Remaining = pos.Quantity
		}

		positions = append(positions, *pos)
	}

	return positions, nil
}

// GetStockPositionsByIDs возвращает позиции по списку ID
func (r *Repository) GetStockPositionsByIDs(ctx context.Context, ids []string) ([]StockPosition, error) {
	if len(ids) == 0 {
		return []StockPosition{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, type, income_batch_id, unit_id, quantity, unit_price, maker, barcode_id, created_at, updated_at
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
			&pos.ID, &pos.Type, &pos.IncomeBatchID, &pos.UnitID, &pos.Quantity, &pos.UnitPrice,
			&pos.Maker, &pos.BarcodeID, &pos.CreatedAt, &pos.UpdatedAt,
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

// GetPositionRemaining считает остаток позиции = quantity - сумма write_off движений
func (r *Repository) GetPositionRemaining(ctx context.Context, positionID string) (float64, error) {
	query := `
		SELECT sp.quantity - COALESCE((
			SELECT SUM(m.quantity)
			FROM stock_position_movements m
			WHERE m.position_id = sp.id AND m.type = 'write_off'
		), 0)
		FROM stock_positions sp
		WHERE sp.id = $1
	`

	var remaining float64
	err := r.pool.QueryRow(ctx, query, positionID).Scan(&remaining)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate position remaining: %w", err)
	}
	return remaining, nil
}

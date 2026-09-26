package wm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// MaxSerialPositionsPerBatch максимальное количество serial-позиций в одном батче
const MaxSerialPositionsPerBatch = 400

// StockBatchExists проверяет существование батча
func (r *Repository) StockBatchExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM stock_batches WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check stock batch existence: %w", err)
	}
	return exists, nil
}

func (r *Repository) GetStockBatchByID(ctx context.Context, id string) (*StockBatch, error) {
	query := `
		SELECT id, direction, status, comment, metadata, created_at, updated_at
		FROM stock_batches
		WHERE id = $1
	`

	var batch StockBatch
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&batch.ID,
		&batch.Direction,
		&batch.Status,
		&batch.Comment,
		&batch.Metadata,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock batch: %w", err)
	}

	positions, err := r.GetPositionsByIncomeBatch(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch positions: %w", err)
	}
	batch.Positions = positions

	return &batch, nil
}

// GetStockBatches возвращает список батчей с пагинацией
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

	if req.Status != nil {
		conditions = append(conditions, "status = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		conditions = append(conditions, "comment ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	fromClause := `FROM stock_batches sb`
	if req.UnitID != nil && *req.UnitID != "" {
		fromClause += ` INNER JOIN stock_positions sp ON sp.income_batch_id = sb.id`
		conditions = append(conditions, "sp.unit_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.UnitID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(DISTINCT sb.id) " + fromClause + " " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count stock batches: %w", err)
	}

	query := `
		SELECT DISTINCT sb.id, sb.direction, sb.status, sb.comment, sb.metadata, sb.created_at, sb.updated_at
	` + fromClause + " " + whereClause + `
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
			&batch.Status,
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

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating batches: %w", err)
	}

	// подтягиваем позиции для каждого батча
	for i := range batches {
		positions, err := r.GetPositionsByIncomeBatch(ctx, batches[i].ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get positions for batch %s: %w", batches[i].ID, err)
		}
		batches[i].Positions = positions
	}

	return batches, total, nil
}

func (r *Repository) CreateStockBatchOnly(ctx context.Context, req CreateStockBatchOnlyRequest) (*StockBatch, error) {
	id := uuid.New().String()

	query := `
		INSERT INTO stock_batches (id, direction, status, comment, metadata, created_at, updated_at)
		VALUES ($1, $2, 'draft', $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, direction, status, comment, metadata, created_at, updated_at
	`

	var batch StockBatch
	err := r.pool.QueryRow(ctx, query, id, req.Direction, req.Comment, req.Metadata).Scan(
		&batch.ID,
		&batch.Direction,
		&batch.Status,
		&batch.Comment,
		&batch.Metadata,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock batch: %w", err)
	}
	batch.Positions = []PositionWithUnitResponse{}
	return &batch, nil
}

// UpdateStockBatchStatus обновляет статус батча
func (r *Repository) UpdateStockBatchStatus(ctx context.Context, id string, status StockBatchStatus) (*StockBatch, error) {
	query := `
		UPDATE stock_batches
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, direction, status, comment, metadata, created_at, updated_at
	`

	var batch StockBatch
	err := r.pool.QueryRow(ctx, query, status, id).Scan(
		&batch.ID,
		&batch.Direction,
		&batch.Status,
		&batch.Comment,
		&batch.Metadata,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update stock batch status: %w", err)
	}

	positions, err := r.GetPositionsByIncomeBatch(ctx, id)
	if err == nil {
		batch.Positions = positions
	} else {
		batch.Positions = []PositionWithUnitResponse{}
	}

	return &batch, nil
}

// CreateStockBatchWithPositions создаёт батч с позициями атомарно
func (r *Repository) CreateStockBatchWithPositions(ctx context.Context, req CreateStockBatchRequest) (*StockBatch, error) {
	if len(req.Positions) == 0 {
		return nil, fmt.Errorf("at least one position is required")
	}

	// Считаем serial-позиции
	serialCount := 0
	for _, pos := range req.Positions {
		trackingDetail, err := r.getUnitTrackingDetail(ctx, pos.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tracking detail for unit %s: %w", pos.UnitID, err)
		}
		if trackingDetail != nil && *trackingDetail == TrackingDetailSerial {
			serialCount += int(pos.Quantity)
		}
	}
	if serialCount > MaxSerialPositionsPerBatch {
		return nil, fmt.Errorf("too many serial positions in one batch: %d (max %d)", serialCount, MaxSerialPositionsPerBatch)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Батч
	batchID := uuid.New().String()
	var batch StockBatch
	err = tx.QueryRow(ctx, `
		INSERT INTO stock_batches (id, direction, status, comment, metadata, created_at, updated_at)
		VALUES ($1, $2, 'draft', $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, direction, status, comment, metadata, created_at, updated_at
	`, batchID, req.Direction, req.Comment, req.Metadata).Scan(
		&batch.ID,
		&batch.Direction,
		&batch.Status,
		&batch.Comment,
		&batch.Metadata,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock batch: %w", err)
	}

	var positions []PositionWithUnitResponse

	for _, posReq := range req.Positions {
		unitExists, err := r.CatalogUnitExists(ctx, posReq.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to check unit existence: %w", err)
		}
		if !unitExists {
			return nil, fmt.Errorf("unit with id '%s' not found", posReq.UnitID)
		}

		unit, err := r.GetCatalogUnitByID(ctx, posReq.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get unit details: %w", err)
		}

		trackingDetail, err := r.getUnitTrackingDetail(ctx, posReq.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tracking detail: %w", err)
		}

		posType := StockPositionTypeBatch
		if trackingDetail != nil && *trackingDetail == TrackingDetailSerial {
			posType = StockPositionTypeSerial
		}

		// barcode: get-or-create
		var barcodeID *string
		if posReq.Barcode != nil && *posReq.Barcode != "" {
			bc, err := r.GetOrCreateBarcode(ctx, *posReq.Barcode, posReq.Maker, &posReq.UnitID)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve barcode: %w", err)
			}
			barcodeID = &bc.ID
		}

		// serial: N записей
		if posType == StockPositionTypeSerial {
			if posReq.Quantity != float64(int(posReq.Quantity)) {
				return nil, fmt.Errorf("serial position quantity must be integer")
			}

			for i := 0; i < int(posReq.Quantity); i++ {
				posID := uuid.New().String()
				var stockPos StockPosition
				err = tx.QueryRow(ctx, `
					INSERT INTO stock_positions (
						id, type, income_batch_id, unit_id, quantity, unit_price,
						maker, barcode_id, created_at, updated_at
					)
					VALUES ($1, $2, $3, $4, 1, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
					RETURNING id, type, income_batch_id, unit_id, quantity, unit_price, maker, barcode_id, created_at, updated_at
				`, posID, posType, batch.ID, posReq.UnitID, posReq.UnitPrice, posReq.Maker, barcodeID).Scan(
					&stockPos.ID,
					&stockPos.Type,
					&stockPos.IncomeBatchID,
					&stockPos.UnitID,
					&stockPos.Quantity,
					&stockPos.UnitPrice,
					&stockPos.Maker,
					&stockPos.BarcodeID,
					&stockPos.CreatedAt,
					&stockPos.UpdatedAt,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to create serial position: %w", err)
				}

				positions = append(positions, PositionWithUnitResponse{
					ID:            stockPos.ID,
					Type:          stockPos.Type,
					IncomeBatchID: stockPos.IncomeBatchID,
					UnitID:        stockPos.UnitID,
					Quantity:      stockPos.Quantity,
					UnitPrice:     stockPos.UnitPrice,
					Maker:         stockPos.Maker,
					BarcodeID:     stockPos.BarcodeID,
					Remaining:     1,
					CreatedAt:     stockPos.CreatedAt,
					UpdatedAt:     stockPos.UpdatedAt,
					Unit:          *unit,
				})
			}
		} else {
			// batch: одна запись
			posID := uuid.New().String()
			var stockPos StockPosition
			err = tx.QueryRow(ctx, `
				INSERT INTO stock_positions (
					id, type, income_batch_id, unit_id, quantity, unit_price,
					maker, barcode_id, created_at, updated_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				RETURNING id, type, income_batch_id, unit_id, quantity, unit_price, maker, barcode_id, created_at, updated_at
			`, posID, posType, batch.ID, posReq.UnitID, posReq.Quantity, posReq.UnitPrice, posReq.Maker, barcodeID).Scan(
				&stockPos.ID,
				&stockPos.Type,
				&stockPos.IncomeBatchID,
				&stockPos.UnitID,
				&stockPos.Quantity,
				&stockPos.UnitPrice,
				&stockPos.Maker,
				&stockPos.BarcodeID,
				&stockPos.CreatedAt,
				&stockPos.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create batch position: %w", err)
			}

			positions = append(positions, PositionWithUnitResponse{
				ID:            stockPos.ID,
				Type:          stockPos.Type,
				IncomeBatchID: stockPos.IncomeBatchID,
				UnitID:        stockPos.UnitID,
				Quantity:      stockPos.Quantity,
				UnitPrice:     stockPos.UnitPrice,
				Maker:         stockPos.Maker,
				BarcodeID:     stockPos.BarcodeID,
				Remaining:     stockPos.Quantity,
				CreatedAt:     stockPos.CreatedAt,
				UpdatedAt:     stockPos.UpdatedAt,
				Unit:          *unit,
			})
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &StockBatch{
		ID:        batch.ID,
		Direction: batch.Direction,
		Status:    batch.Status,
		Comment:   batch.Comment,
		Metadata:  batch.Metadata,
		CreatedAt: batch.CreatedAt,
		UpdatedAt: batch.UpdatedAt,
		Positions: positions,
	}, nil
}

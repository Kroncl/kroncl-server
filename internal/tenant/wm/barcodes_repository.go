package wm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// --------
// BARCODES
// --------

// BarcodeExists проверяет существование баркода по строке
func (r *Repository) BarcodeExists(ctx context.Context, barcode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM barcodes WHERE barcode = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, barcode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check barcode existence: %w", err)
	}
	return exists, nil
}

// GetBarcodeByID возвращает баркод по ID
func (r *Repository) GetBarcodeByID(ctx context.Context, id string) (*Barcode, error) {
	query := `
		SELECT id, barcode, maker, catalog_unit_id, metadata, created_at, updated_at
		FROM barcodes
		WHERE id = $1
	`

	var b Barcode
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.Barcode,
		&b.Maker,
		&b.CatalogUnitID,
		&b.Metadata,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode: %w", err)
	}
	return &b, nil
}

// GetBarcodeByValue возвращает баркод по строке
func (r *Repository) GetBarcodeByValue(ctx context.Context, barcode string) (*Barcode, error) {
	query := `
		SELECT id, barcode, maker, catalog_unit_id, metadata, created_at, updated_at
		FROM barcodes
		WHERE barcode = $1
	`

	var b Barcode
	err := r.pool.QueryRow(ctx, query, barcode).Scan(
		&b.ID,
		&b.Barcode,
		&b.Maker,
		&b.CatalogUnitID,
		&b.Metadata,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode by value: %w", err)
	}
	return &b, nil
}

// GetBarcodes возвращает список баркодов с пагинацией
func (r *Repository) GetBarcodes(ctx context.Context, req GetBarcodesParams) ([]Barcode, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	if req.CatalogUnitID != nil && *req.CatalogUnitID != "" {
		conditions = append(conditions, "catalog_unit_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.CatalogUnitID)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		searchConditions := []string{
			"barcode ILIKE $" + strconv.Itoa(argIndex),
			"maker ILIKE $" + strconv.Itoa(argIndex),
		}
		conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM barcodes " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count barcodes: %w", err)
	}

	query := `
		SELECT id, barcode, maker, catalog_unit_id, metadata, created_at, updated_at
		FROM barcodes
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query barcodes: %w", err)
	}
	defer rows.Close()

	var barcodes []Barcode
	for rows.Next() {
		var b Barcode
		err := rows.Scan(
			&b.ID,
			&b.Barcode,
			&b.Maker,
			&b.CatalogUnitID,
			&b.Metadata,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan barcode: %w", err)
		}
		barcodes = append(barcodes, b)
	}

	return barcodes, total, nil
}

// CreateBarcode создаёт баркод в словаре
func (r *Repository) CreateBarcode(ctx context.Context, req CreateBarcodeRequest) (*Barcode, error) {
	id := uuid.New().String()

	query := `
		INSERT INTO barcodes (id, barcode, maker, catalog_unit_id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, barcode, maker, catalog_unit_id, metadata, created_at, updated_at
	`

	var b Barcode
	err := r.pool.QueryRow(ctx, query,
		id,
		strings.TrimSpace(req.Barcode),
		req.Maker,
		req.CatalogUnitID,
		req.Metadata,
	).Scan(
		&b.ID,
		&b.Barcode,
		&b.Maker,
		&b.CatalogUnitID,
		&b.Metadata,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create barcode: %w", err)
	}

	return &b, nil
}

// GetOrCreateBarcode возвращает существующий баркод или создаёт новый (upsert)
func (r *Repository) GetOrCreateBarcode(ctx context.Context, barcode string, maker *string, catalogUnitID *string) (*Barcode, error) {
	// пытаемся найти
	existing, err := r.GetBarcodeByValue(ctx, barcode)
	if err == nil {
		return existing, nil
	}

	// создаём
	return r.CreateBarcode(ctx, CreateBarcodeRequest{
		Barcode:       barcode,
		Maker:         maker,
		CatalogUnitID: catalogUnitID,
	})
}

// UpdateBarcode обновляет баркод
func (r *Repository) UpdateBarcode(ctx context.Context, id string, req UpdateBarcodeRequest) (*Barcode, error) {
	updater := core.NewUpdater("barcodes")

	if req.Barcode != nil {
		barcode := strings.TrimSpace(*req.Barcode)
		if barcode == "" {
			return nil, fmt.Errorf("barcode cannot be empty")
		}
		updater.SetString("barcode", barcode)
	}

	if req.Maker != nil {
		maker := strings.TrimSpace(*req.Maker)
		if maker == "" {
			updater.SetNull("maker")
		} else {
			updater.SetString("maker", maker)
		}
	}

	if req.CatalogUnitID != nil {
		if *req.CatalogUnitID == "" {
			updater.SetNull("catalog_unit_id")
		} else {
			updater.SetString("catalog_unit_id", *req.CatalogUnitID)
		}
	}

	if req.Metadata != nil {
		updater.SetJSONB("metadata", *req.Metadata)
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetBarcodeByID(ctx, id)
	}

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update barcode: %w", err)
	}

	return r.GetBarcodeByID(ctx, id)
}

// DeleteBarcode удаляет баркод
func (r *Repository) DeleteBarcode(ctx context.Context, id string) (bool, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM barcodes WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete barcode: %w", err)
	}
	return result.RowsAffected() > 0, nil
}

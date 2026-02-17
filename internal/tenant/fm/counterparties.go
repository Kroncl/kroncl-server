package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// GetCounterpartyByID возвращает контрагента по ID
func (r *Repository) GetCounterpartyByID(ctx context.Context, id string) (*Counterparty, error) {
	query := `
		SELECT 
			id, name, comment, type, status, metadata, created_at, updated_at
		FROM counterparties
		WHERE id = $1
	`

	var counterparty Counterparty
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&counterparty.ID,
		&counterparty.Name,
		&counterparty.Comment,
		&counterparty.Type,
		&counterparty.Status,
		&counterparty.Metadata,
		&counterparty.CreatedAt,
		&counterparty.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get counterparty: %w", err)
	}

	return &counterparty, nil
}

// GetCounterparties возвращает список контрагентов с пагинацией и фильтрацией
func (r *Repository) GetCounterparties(ctx context.Context, offset, limit int, filters GetCounterpartiesRequest) ([]Counterparty, int, error) {
	var whereClause string
	var args []interface{}
	var whereConditions []string
	argIndex := 1

	if filters.Type != nil {
		whereConditions = append(whereConditions, "type = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.Type)
		argIndex++
	}

	if filters.Status != nil {
		whereConditions = append(whereConditions, "status = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		searchConditions := []string{
			"name ILIKE $" + strconv.Itoa(argIndex),
			"comment ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM counterparties ` + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count counterparties: %w", err)
	}

	// Получаем контрагентов с пагинацией
	query := `
		SELECT 
			id, name, comment, type, status, metadata, created_at, updated_at
		FROM counterparties
		` + whereClause + `
		ORDER BY name ASC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query counterparties: %w", err)
	}
	defer rows.Close()

	var counterparties []Counterparty
	for rows.Next() {
		var counterparty Counterparty
		err := rows.Scan(
			&counterparty.ID,
			&counterparty.Name,
			&counterparty.Comment,
			&counterparty.Type,
			&counterparty.Status,
			&counterparty.Metadata,
			&counterparty.CreatedAt,
			&counterparty.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan counterparty: %w", err)
		}
		counterparties = append(counterparties, counterparty)
	}

	return counterparties, total, nil
}

// CreateCounterparty создает нового контрагента
func (r *Repository) CreateCounterparty(ctx context.Context, req CreateCounterpartyRequest) (*Counterparty, error) {
	// Валидация
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("counterparty name is required")
	}

	comment := strings.TrimSpace(req.Comment)
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	// Статус по умолчанию
	status := CounterpartyStatusActive
	if req.Status != "" {
		status = req.Status
	}

	id := uuid.New().String()

	query := `
		INSERT INTO counterparties (
			id, name, comment, type, status, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, name, comment, type, status, metadata, created_at, updated_at
	`

	var counterparty Counterparty
	err := r.pool.QueryRow(ctx, query,
		id,
		name,
		commentPtr,
		req.Type,
		status,
		req.Metadata,
	).Scan(
		&counterparty.ID,
		&counterparty.Name,
		&counterparty.Comment,
		&counterparty.Type,
		&counterparty.Status,
		&counterparty.Metadata,
		&counterparty.CreatedAt,
		&counterparty.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create counterparty: %w", err)
	}

	return &counterparty, nil
}

// UpdateCounterparty обновляет контрагента (без статуса)
func (r *Repository) UpdateCounterparty(ctx context.Context, id string, req UpdateCounterpartyRequest) (*Counterparty, error) {
	// Проверяем существование
	_, err := r.GetCounterpartyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("counterparty not found: %w", err)
	}

	updater := core.NewUpdater("counterparties")

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name != "" {
			updater.SetString("name", name)
		}
	}

	if req.Comment != nil {
		comment := strings.TrimSpace(*req.Comment)
		if comment == "" {
			updater.SetNull("comment")
		} else {
			updater.SetString("comment", comment)
		}
	}

	if req.Type != nil {
		updater.SetString("type", string(*req.Type))
	}

	if req.Metadata != nil {
		updater.Set("metadata", *req.Metadata)
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetCounterpartyByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update counterparty: %w", err)
	}

	return r.GetCounterpartyByID(ctx, id)
}

// ActivateCounterparty активирует контрагента
func (r *Repository) ActivateCounterparty(ctx context.Context, id string) (*Counterparty, error) {
	// Проверяем существование
	_, err := r.GetCounterpartyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("counterparty not found: %w", err)
	}

	query := `UPDATE counterparties SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = r.pool.Exec(ctx, query, CounterpartyStatusActive, id)
	if err != nil {
		return nil, fmt.Errorf("failed to activate counterparty: %w", err)
	}

	return r.GetCounterpartyByID(ctx, id)
}

// DeactivateCounterparty деактивирует контрагента
func (r *Repository) DeactivateCounterparty(ctx context.Context, id string) (*Counterparty, error) {
	// Проверяем существование
	_, err := r.GetCounterpartyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("counterparty not found: %w", err)
	}

	query := `UPDATE counterparties SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = r.pool.Exec(ctx, query, CounterpartyStatusInactive, id)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate counterparty: %w", err)
	}

	return r.GetCounterpartyByID(ctx, id)
}

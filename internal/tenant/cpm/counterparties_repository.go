package cpm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ----------
// COUNTERPARTIES
// ----------

func (r *Repository) GetCounterpartyByID(ctx context.Context, id string) (*Counterparty, error) {
	query := `
		SELECT id, name, comment, type, status, inn, ogrn, kpp, address, default_currency, metadata, created_at, updated_at
		FROM counterparties
		WHERE id = $1
	`

	var c Counterparty
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Comment, &c.Type, &c.Status,
		&c.INN, &c.OGRN, &c.KPP, &c.Address, &c.DefaultCurrency,
		&c.Metadata, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get counterparty: %w", err)
	}

	return &c, nil
}

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
			"inn ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM counterparties ` + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count counterparties: %w", err)
	}

	query := `
		SELECT id, name, comment, type, status, inn, ogrn, kpp, address, default_currency, metadata, created_at, updated_at
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
		var c Counterparty
		err := rows.Scan(
			&c.ID, &c.Name, &c.Comment, &c.Type, &c.Status,
			&c.INN, &c.OGRN, &c.KPP, &c.Address, &c.DefaultCurrency,
			&c.Metadata, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan counterparty: %w", err)
		}
		counterparties = append(counterparties, c)
	}

	return counterparties, total, nil
}

func (r *Repository) CreateCounterparty(ctx context.Context, req CreateCounterpartyRequest) (*Counterparty, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("counterparty name is required")
	}

	comment := strings.TrimSpace(req.Comment)
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	status := CounterpartyStatusActive
	if req.Status != "" {
		status = req.Status
	}

	// Валидация default_currency
	var defaultCurrencyPtr *string
	if req.DefaultCurrency != "" {
		if _, err := r.currencyService.GetByID(ctx, req.DefaultCurrency); err != nil {
			return nil, fmt.Errorf("invalid default_currency: %s", req.DefaultCurrency)
		}
		defaultCurrencyPtr = &req.DefaultCurrency
	}

	// CreateCounterparty
	var innPtr, ogrnPtr, kppPtr, addressPtr *string

	if req.INN = strings.TrimSpace(req.INN); req.INN != "" {
		innPtr = &req.INN
	}
	if req.OGRN = strings.TrimSpace(req.OGRN); req.OGRN != "" {
		ogrnPtr = &req.OGRN
	}
	if req.KPP = strings.TrimSpace(req.KPP); req.KPP != "" {
		kppPtr = &req.KPP
	}
	if req.Address = strings.TrimSpace(req.Address); req.Address != "" {
		addressPtr = &req.Address
	}

	id := uuid.New().String()

	query := `
		INSERT INTO counterparties (
			id, name, comment, type, status, inn, ogrn, kpp, address, default_currency, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING id, name, comment, type, status, inn, ogrn, kpp, address, default_currency, metadata, created_at, updated_at
	`

	var c Counterparty
	err := r.pool.QueryRow(ctx, query,
		id, name, commentPtr, req.Type, status,
		innPtr, ogrnPtr, kppPtr, addressPtr, defaultCurrencyPtr,
		req.Metadata,
	).Scan(
		&c.ID, &c.Name, &c.Comment, &c.Type, &c.Status,
		&c.INN, &c.OGRN, &c.KPP, &c.Address, &c.DefaultCurrency,
		&c.Metadata, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create counterparty: %w", err)
	}

	return &c, nil
}

func (r *Repository) UpdateCounterparty(ctx context.Context, id string, req UpdateCounterpartyRequest) (*Counterparty, error) {
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

	if req.INN != nil {
		inn := strings.TrimSpace(*req.INN)
		if inn == "" {
			updater.SetNull("inn")
		} else {
			updater.SetString("inn", inn)
		}
	}

	if req.OGRN != nil {
		ogrn := strings.TrimSpace(*req.OGRN)
		if ogrn == "" {
			updater.SetNull("ogrn")
		} else {
			updater.SetString("ogrn", ogrn)
		}
	}

	if req.KPP != nil {
		kpp := strings.TrimSpace(*req.KPP)
		if kpp == "" {
			updater.SetNull("kpp")
		} else {
			updater.SetString("kpp", kpp)
		}
	}

	if req.Address != nil {
		addr := strings.TrimSpace(*req.Address)
		if addr == "" {
			updater.SetNull("address")
		} else {
			updater.SetString("address", addr)
		}
	}

	if req.DefaultCurrency != nil {
		dc := strings.TrimSpace(*req.DefaultCurrency)
		if dc == "" {
			updater.SetNull("default_currency")
		} else {
			if _, err := r.currencyService.GetByID(ctx, dc); err != nil {
				return nil, fmt.Errorf("invalid default_currency: %s", dc)
			}
			updater.SetString("default_currency", dc)
		}
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

func (r *Repository) ActivateCounterparty(ctx context.Context, id string) (*Counterparty, error) {
	_, err := r.GetCounterpartyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("counterparty not found: %w", err)
	}

	_, err = r.pool.Exec(ctx, `UPDATE counterparties SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, CounterpartyStatusActive, id)
	if err != nil {
		return nil, fmt.Errorf("failed to activate counterparty: %w", err)
	}

	return r.GetCounterpartyByID(ctx, id)
}

func (r *Repository) DeactivateCounterparty(ctx context.Context, id string) (*Counterparty, error) {
	_, err := r.GetCounterpartyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("counterparty not found: %w", err)
	}

	_, err = r.pool.Exec(ctx, `UPDATE counterparties SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, CounterpartyStatusInactive, id)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate counterparty: %w", err)
	}

	return r.GetCounterpartyByID(ctx, id)
}

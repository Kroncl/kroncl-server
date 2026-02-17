package fm

import (
	"context"
	"encoding/json"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GetCreditByID возвращает кредит по ID с данными контрагента
func (r *Repository) GetCreditByID(ctx context.Context, id string) (*CreditDetail, error) {
	query := `
		SELECT 
			c.id,
			c.name,
			c.comment,
			c.type,
			c.status,
			c.total_amount,
			c.currency,
			c.interest_rate,
			c.start_date,
			c.end_date,
			c.metadata,
			c.created_at,
			c.updated_at,
			-- данные контрагента из связи
			cp.id as counterparty_id,
			cp.name as counterparty_name,
			cp.comment as counterparty_comment,
			cp.type as counterparty_type,
			cp.status as counterparty_status,
			cp.metadata as counterparty_metadata,
			cp.created_at as counterparty_created_at,
			cp.updated_at as counterparty_updated_at
		FROM credits c
		LEFT JOIN credit_counterparty cc ON c.id = cc.credit_id
		LEFT JOIN counterparties cp ON cc.counterparty_id = cp.id
		WHERE c.id = $1
	`

	var credit CreditDetail
	var counterparty Counterparty
	var cpID, cpName, cpComment, cpType, cpStatus *string
	var cpMetadata []byte // вместо *map[string]interface{} используем []byte для JSONB
	var cpCreatedAt, cpUpdatedAt *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&credit.ID,
		&credit.Name,
		&credit.Comment,
		&credit.Type,
		&credit.Status,
		&credit.TotalAmount,
		&credit.Currency,
		&credit.InterestRate,
		&credit.StartDate,
		&credit.EndDate,
		&credit.Metadata,
		&credit.CreatedAt,
		&credit.UpdatedAt,
		// counterparty
		&cpID,
		&cpName,
		&cpComment,
		&cpType,
		&cpStatus,
		&cpMetadata, // теперь сканируем как []byte
		&cpCreatedAt,
		&cpUpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get credit: %w", err)
	}

	// Заполняем данные контрагента если они есть
	if cpID != nil {
		counterparty.ID = *cpID
		counterparty.Name = *cpName
		counterparty.Comment = cpComment
		counterparty.Type = CounterpartyType(*cpType)
		counterparty.Status = CounterpartyStatus(*cpStatus)
		counterparty.CreatedAt = *cpCreatedAt
		counterparty.UpdatedAt = *cpUpdatedAt

		// Парсим metadata если оно есть
		if cpMetadata != nil {
			var metadata map[string]interface{}
			if err := json.Unmarshal(cpMetadata, &metadata); err == nil {
				counterparty.Metadata = metadata
			}
		}

		credit.Counterparty = &counterparty
	}

	return &credit, nil
}

// GetCredits возвращает список кредитов с пагинацией, фильтрацией и данными контрагентов
func (r *Repository) GetCredits(ctx context.Context, offset, limit int, filters GetCreditsRequest) ([]CreditDetail, int, error) {
	var whereClause string
	var args []interface{}
	var whereConditions []string
	argIndex := 1

	if filters.Type != nil {
		whereConditions = append(whereConditions, "c.type = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.Type)
		argIndex++
	}

	if filters.Status != nil {
		whereConditions = append(whereConditions, "c.status = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		searchConditions := []string{
			"c.name ILIKE $" + strconv.Itoa(argIndex),
			"c.comment ILIKE $" + strconv.Itoa(argIndex),
			"cp.name ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Базовый запрос с JOIN
	queryBase := `
		FROM credits c
		LEFT JOIN credit_counterparty cc ON c.id = cc.credit_id
		LEFT JOIN counterparties cp ON cc.counterparty_id = cp.id
	`

	// Получаем общее количество
	countQuery := "SELECT COUNT(DISTINCT c.id) " + queryBase + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count credits: %w", err)
	}

	// Получаем кредиты с пагинацией
	query := `
		SELECT DISTINCT
			c.id,
			c.name,
			c.comment,
			c.type,
			c.status,
			c.total_amount,
			c.currency,
			c.interest_rate,
			c.start_date,
			c.end_date,
			c.metadata,
			c.created_at,
			c.updated_at,
			cp.id as counterparty_id,
			cp.name as counterparty_name,
			cp.comment as counterparty_comment,
			cp.type as counterparty_type,
			cp.status as counterparty_status,
			cp.metadata as counterparty_metadata,
			cp.created_at as counterparty_created_at,
			cp.updated_at as counterparty_updated_at
	` + queryBase + whereClause + `
		ORDER BY c.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query credits: %w", err)
	}
	defer rows.Close()

	var credits []CreditDetail
	for rows.Next() {
		var credit CreditDetail
		var counterparty Counterparty
		var cpID, cpName, cpComment, cpType, cpStatus *string
		var cpMetadata []byte
		var cpCreatedAt, cpUpdatedAt *time.Time

		err := rows.Scan(
			&credit.ID,
			&credit.Name,
			&credit.Comment,
			&credit.Type,
			&credit.Status,
			&credit.TotalAmount,
			&credit.Currency,
			&credit.InterestRate,
			&credit.StartDate,
			&credit.EndDate,
			&credit.Metadata,
			&credit.CreatedAt,
			&credit.UpdatedAt,
			// counterparty
			&cpID,
			&cpName,
			&cpComment,
			&cpType,
			&cpStatus,
			&cpMetadata,
			&cpCreatedAt,
			&cpUpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan credit: %w", err)
		}

		// Заполняем данные контрагента если они есть
		if cpID != nil {
			counterparty.ID = *cpID
			counterparty.Name = *cpName
			counterparty.Comment = cpComment
			counterparty.Type = CounterpartyType(*cpType)
			counterparty.Status = CounterpartyStatus(*cpStatus)
			counterparty.CreatedAt = *cpCreatedAt
			counterparty.UpdatedAt = *cpUpdatedAt

			// Парсим metadata если оно есть
			if cpMetadata != nil {
				var metadata map[string]interface{}
				if err := json.Unmarshal(cpMetadata, &metadata); err == nil {
					counterparty.Metadata = metadata
				}
			}

			credit.Counterparty = &counterparty
		}

		credits = append(credits, credit)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating credits: %w", err)
	}

	return credits, total, nil
}

// CreateCredit создает новый кредит и связь с контрагентом в одной транзакции
func (r *Repository) CreateCredit(ctx context.Context, req CreateCreditRequest) (*CreditDetail, error) {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Проверяем существование активного контрагента
	var counterpartyExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM counterparties WHERE id = $1 AND status = 'active')`
	err = tx.QueryRow(ctx, checkQuery, req.CounterpartyID).Scan(&counterpartyExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check counterparty: %w", err)
	}
	if !counterpartyExists {
		return nil, fmt.Errorf("counterparty not found: %s", req.CounterpartyID)
	}

	// Валидация
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("credit name is required")
	}

	comment := strings.TrimSpace(req.Comment)
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	id := uuid.New().String()

	// 1. Создаем кредит (без counterparty_id)
	creditQuery := `
		INSERT INTO credits (
			id, name, comment, type, status, total_amount, currency,
			interest_rate, start_date, end_date, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, name, comment, type, status, total_amount, currency,
			interest_rate, start_date, end_date, metadata,
			created_at, updated_at
	`

	var credit Credit
	err = tx.QueryRow(ctx, creditQuery,
		id,
		name,
		commentPtr,
		req.Type,
		CreditStatusActive,
		req.TotalAmount,
		req.Currency,
		req.InterestRate,
		req.StartDate,
		req.EndDate,
		req.Metadata,
	).Scan(
		&credit.ID,
		&credit.Name,
		&credit.Comment,
		&credit.Type,
		&credit.Status,
		&credit.TotalAmount,
		&credit.Currency,
		&credit.InterestRate,
		&credit.StartDate,
		&credit.EndDate,
		&credit.Metadata,
		&credit.CreatedAt,
		&credit.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create credit: %w", err)
	}

	// 2. Создаем связь в credit_counterparty
	linkQuery := `
		INSERT INTO credit_counterparty (
			id, credit_id, counterparty_id, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP
		)
	`
	_, err = tx.Exec(ctx, linkQuery, credit.ID, req.CounterpartyID)
	if err != nil {
		return nil, fmt.Errorf("failed to link credit to counterparty: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Возвращаем полную информацию о кредите с контрагентом
	return r.GetCreditByID(ctx, credit.ID)
}

// UpdateCredit обновляет кредит (без статуса)
func (r *Repository) UpdateCredit(ctx context.Context, id string, req UpdateCreditRequest) (*CreditDetail, error) {
	// Проверяем существование
	_, err := r.GetCreditByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("credit not found: %w", err)
	}

	updater := core.NewUpdater("credits")

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

	if req.TotalAmount != nil {
		updater.SetInt("total_amount", int(*req.TotalAmount))
	}

	if req.Currency != nil {
		updater.SetString("currency", string(*req.Currency))
	}

	if req.InterestRate != nil {
		updater.SetFloat("interest_rate", *req.InterestRate)
	}

	if req.StartDate != nil {
		updater.SetTime("start_date", *req.StartDate)
	}

	if req.EndDate != nil {
		updater.SetTime("end_date", *req.EndDate)
	}

	if req.Metadata != nil {
		updater.Set("metadata", *req.Metadata)
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetCreditByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update credit: %w", err)
	}

	// Обновление связи с контрагентом если нужно (отдельным методом)
	if req.CounterpartyID != nil {
		// Проверяем существование нового контрагента
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM counterparties WHERE id = $1)`
		err = r.pool.QueryRow(ctx, checkQuery, *req.CounterpartyID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("failed to check counterparty: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("counterparty not found: %s", *req.CounterpartyID)
		}

		// Обновляем связь в credit_counterparty
		updateLinkQuery := `
			UPDATE credit_counterparty 
			SET counterparty_id = $1, updated_at = CURRENT_TIMESTAMP 
			WHERE credit_id = $2
		`
		_, err = r.pool.Exec(ctx, updateLinkQuery, *req.CounterpartyID, id)
		if err != nil {
			return nil, fmt.Errorf("failed to update credit counterparty: %w", err)
		}
	}

	return r.GetCreditByID(ctx, id)
}

// ActivateCredit активирует кредит
func (r *Repository) ActivateCredit(ctx context.Context, id string) (*CreditDetail, error) {
	// Проверяем существование
	_, err := r.GetCreditByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("credit not found: %w", err)
	}

	query := `UPDATE credits SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = r.pool.Exec(ctx, query, CreditStatusActive, id)
	if err != nil {
		return nil, fmt.Errorf("failed to activate credit: %w", err)
	}

	return r.GetCreditByID(ctx, id)
}

// DeactivateCredit (закрывает) кредит
func (r *Repository) DeactivateCredit(ctx context.Context, id string) (*CreditDetail, error) {
	// Проверяем существование
	_, err := r.GetCreditByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("credit not found: %w", err)
	}

	query := `UPDATE credits SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = r.pool.Exec(ctx, query, CreditStatusClosed, id)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate credit: %w", err)
	}

	return r.GetCreditByID(ctx, id)
}

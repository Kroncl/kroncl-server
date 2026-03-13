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

// ----------
// CREDITS
// ----------

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

// ---------
// PAYMENTS
// ---------

// PayCredit creates a transaction and links it to a credit
func (r *Repository) PayCredit(ctx context.Context, req PayCreditRequest) (*TransactionDetail, error) {
	// Начинаем транзакцию БД
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Получаем кредит и считаем остаток
	var totalAmount int64
	var paidAmount int64
	var creditType CreditType
	var creditStatus CreditStatus

	creditQuery := `
        SELECT 
            c.total_amount,
            c.type,
            c.status,
            COALESCE(SUM(t.base_amount), 0) as paid
        FROM credits c
        LEFT JOIN credit_transactions ct ON c.id = ct.credit_id
        LEFT JOIN transactions t ON ct.transaction_id = t.id AND t.reverse_to IS NULL
        WHERE c.id = $1
        GROUP BY c.id, c.total_amount, c.type, c.status
    `

	err = tx.QueryRow(ctx, creditQuery, req.CreditID).Scan(
		&totalAmount,
		&creditType,
		&creditStatus,
		&paidAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit: %w", err)
	}

	// Проверяем, что кредит активен
	if creditStatus != CreditStatusActive {
		return nil, fmt.Errorf("credit is not active")
	}

	// Проверяем, что платеж не превышает остаток
	remaining := totalAmount - paidAmount
	if req.Amount > remaining {
		return nil, fmt.Errorf("payment amount (%d) exceeds remaining debt (%d)", req.Amount, remaining)
	}

	// 2. Определяем направление транзакции на основе типа кредита
	direction := TransactionDirectionExpense
	if creditType == CreditTypeCredit {
		direction = TransactionDirectionIncome
	}

	// 3. Получаем ID категории для кредитов/займов
	var categoryID string
	if creditType == CreditTypeCredit {
		categoryID, err = r.GetCategoryIDBySlug(ctx, "credit")
	} else {
		categoryID, err = r.GetCategoryIDBySlug(ctx, "dept")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// 4. Создаем транзакцию
	transactionQuery := `
        INSERT INTO transactions (
            id, base_amount, currency, direction, status, comment,
            created_at, metadata
        ) VALUES (
            gen_random_uuid(), $1, $2, $3, $4, $5,
            $6, $7
        )
        RETURNING id
    `

	var transactionID string
	err = tx.QueryRow(ctx, transactionQuery,
		req.Amount,
		CurrencyRUB,
		direction,
		TransactionStatusCompleted,
		req.Comment,
		req.PaidAt,
		nil, // metadata
	).Scan(&transactionID)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// 5. Создаем связь с сотрудником
	linkEmployeeQuery := `
        INSERT INTO transaction_employee (
            id, employee_id, transaction_id, created_at
        ) VALUES (
            gen_random_uuid(), $1, $2, $3
        )
    `
	_, err = tx.Exec(ctx, linkEmployeeQuery, req.EmployeeID, transactionID, req.PaidAt)
	if err != nil {
		return nil, fmt.Errorf("failed to link employee: %w", err)
	}

	// 6. Создаем связь с категорией
	linkCategoryQuery := `
        INSERT INTO transaction_category (
            id, transaction_id, category_id, created_at
        ) VALUES (
            gen_random_uuid(), $1, $2, $3
        )
    `
	_, err = tx.Exec(ctx, linkCategoryQuery, transactionID, categoryID, req.PaidAt)
	if err != nil {
		return nil, fmt.Errorf("failed to link category: %w", err)
	}

	// 7. Создаем связь с кредитом
	linkCreditQuery := `
        INSERT INTO credit_transactions (
            id, credit_id, transaction_id, created_at
        ) VALUES (
            gen_random_uuid(), $1, $2, $3
        )
    `
	_, err = tx.Exec(ctx, linkCreditQuery, req.CreditID, transactionID, req.PaidAt)
	if err != nil {
		return nil, fmt.Errorf("failed to link credit: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Возвращаем созданную транзакцию
	return r.GetTransactionByID(ctx, transactionID)
}

// GetCreditTransactions returns list of transactions for a credit
func (r *Repository) GetCreditTransactions(ctx context.Context, creditID string, offset, limit int) ([]TransactionDetail, int64, error) {
	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM credit_transactions WHERE credit_id = $1`
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, creditID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count credit transactions: %w", err)
	}

	// Получаем транзакции с пагинацией
	query := `
        SELECT 
            t.id,
            t.base_amount,
            t.currency,
            t.direction,
            t.status,
            t.comment,
            t.reverse_to,
            t.created_at,
            t.metadata,
            te.employee_id,
            e.first_name,
            e.last_name,
            tc.category_id,
            c.name as category_name,
            c.description as category_description,
            c.direction as category_direction,
            c.created_at as category_created_at,
            c.updated_at as category_updated_at,
            c.slug as category_slug
        FROM credit_transactions ct
        JOIN transactions t ON ct.transaction_id = t.id
        LEFT JOIN transaction_employee te ON t.id = te.transaction_id
        LEFT JOIN employees e ON te.employee_id = e.id
        LEFT JOIN transaction_category tc ON t.id = tc.transaction_id
        LEFT JOIN transaction_categories c ON tc.category_id = c.id
        WHERE ct.credit_id = $1
        ORDER BY t.created_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.pool.Query(ctx, query, creditID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query credit transactions: %w", err)
	}
	defer rows.Close()

	var transactions []TransactionDetail
	for rows.Next() {
		var detail TransactionDetail
		var employeeID, employeeFirstName, employeeLastName *string
		var categoryID, categoryName, categoryDescription *string
		var reverseTo *string
		var categoryDirection *TransactionCategoryDirection
		var categoryCreatedAt, categoryUpdatedAt *time.Time
		var categorySlug *string

		err := rows.Scan(
			&detail.ID,
			&detail.BaseAmount,
			&detail.Currency,
			&detail.Direction,
			&detail.Status,
			&detail.Comment,
			&reverseTo,
			&detail.CreatedAt,
			&detail.Metadata,
			&employeeID,
			&employeeFirstName,
			&employeeLastName,
			&categoryID,
			&categoryName,
			&categoryDescription,
			&categoryDirection,
			&categoryCreatedAt,
			&categoryUpdatedAt,
			&categorySlug,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}

		detail.ReverseTo = reverseTo
		detail.EmployeeID = employeeID
		detail.EmployeeFirstName = employeeFirstName
		detail.EmployeeLastName = employeeLastName
		detail.CategoryID = categoryID
		detail.CategoryName = categoryName

		transactions = append(transactions, detail)
	}

	return transactions, total, nil
}

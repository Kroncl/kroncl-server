package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/tenant/hrm"
	"strconv"
	"strings"
	"time"
)

// CreateDealTransaction создает транзакцию с привязкой к сделке и автоматическим определением категории
func (r *Repository) CreateDealTransaction(ctx context.Context, dealID string, req CreateTransactionRequest) (*TransactionDetail, error) {
	if req.BaseAmount <= 0 {
		return nil, fmt.Errorf("base_amount must be greater than 0")
	}

	if req.Direction != TransactionDirectionIncome && req.Direction != TransactionDirectionExpense {
		return nil, fmt.Errorf("invalid transaction direction: %s", req.Direction)
	}

	switch req.Currency {
	case CurrencyRUB:
	default:
		return nil, fmt.Errorf("invalid currency: %s", req.Currency)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var dealExists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM deals WHERE id = $1)`, dealID).Scan(&dealExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check deal: %w", err)
	}
	if !dealExists {
		return nil, fmt.Errorf("deal not found: %s", dealID)
	}

	if req.EmployeeID != "" {
		employee, err := r.employeesRepository.GetEmployeeByID(ctx, req.EmployeeID)
		if err != nil {
			return nil, fmt.Errorf("invalid employee_id: %w", err)
		}
		_ = employee
	}

	comment := strings.TrimSpace(req.Comment)
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	status := TransactionStatusCompleted
	if req.Status != "" {
		status = TransactionStatus(req.Status)
	}

	transactionQuery := `
		INSERT INTO transactions (
			id, base_amount, currency, direction, status, comment,
			created_at, metadata
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5,
			CURRENT_TIMESTAMP, $6
		)
		RETURNING id
	`

	var transactionID string
	err = tx.QueryRow(ctx, transactionQuery,
		req.BaseAmount,
		req.Currency,
		req.Direction,
		status,
		commentPtr,
		req.Metadata,
	).Scan(&transactionID)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	linkDealQuery := `
		INSERT INTO deals_transactions (
			id, deal_id, transaction_id, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
	`
	_, err = tx.Exec(ctx, linkDealQuery, dealID, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to link deal to transaction: %w", err)
	}

	if req.EmployeeID != "" {
		linkEmployeeQuery := `
			INSERT INTO transaction_employee (
				id, employee_id, transaction_id, created_at
			) VALUES (
				gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP
			)
		`
		_, err = tx.Exec(ctx, linkEmployeeQuery, req.EmployeeID, transactionID)
		if err != nil {
			return nil, fmt.Errorf("failed to link employee to transaction: %w", err)
		}
	}

	categorySlug := DEAL_TRANSACTION_CATEGORY_EXPENSE_SLUG
	if req.Direction == TransactionDirectionIncome {
		categorySlug = DEAL_TRANSACTION_CATEGORY_INCOME_SLUG
	}

	var categoryID string
	err = tx.QueryRow(ctx, `SELECT id FROM transaction_categories WHERE slug = $1`, categorySlug).Scan(&categoryID)
	if err != nil {
		return nil, fmt.Errorf("deal category not found by slug: %s", categorySlug)
	}

	linkCategoryQuery := `
		INSERT INTO transaction_category (
			id, transaction_id, category_id, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP
		)
	`
	_, err = tx.Exec(ctx, linkCategoryQuery, transactionID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to link category to transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetTransactionByID(ctx, transactionID)
}

// GetDealTransactions возвращает список транзакций, привязанных к сделке, с пагинацией и фильтрацией
func (r *Repository) GetDealTransactions(ctx context.Context, dealID string, offset, limit int, filters GetTransactionsRequest) ([]TransactionDetail, int64, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	args = append(args, dealID)
	whereConditions = append(whereConditions, "dt.deal_id = $1")
	argIndex++

	queryBase := `
		FROM transactions t
		INNER JOIN deals_transactions dt ON t.id = dt.transaction_id
		LEFT JOIN transaction_employee te ON t.id = te.transaction_id
		LEFT JOIN employees e ON te.employee_id = e.id
		LEFT JOIN transaction_category tc ON t.id = tc.transaction_id
		LEFT JOIN transaction_categories c ON tc.category_id = c.id
	`

	if filters.StartDate != nil {
		whereConditions = append(whereConditions, "t.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		whereConditions = append(whereConditions, "t.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	if filters.Direction != nil {
		whereConditions = append(whereConditions, "t.direction = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.Direction)
		argIndex++
	}

	if filters.Status != nil {
		whereConditions = append(whereConditions, "t.status = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.CategoryID != nil {
		whereConditions = append(whereConditions, "c.id = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.CategoryID)
		argIndex++
	}

	if filters.EmployeeID != nil {
		whereConditions = append(whereConditions, "e.id = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.EmployeeID)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		searchConditions := []string{
			"t.comment ILIKE $" + strconv.Itoa(argIndex),
			"e.first_name ILIKE $" + strconv.Itoa(argIndex),
			"e.last_name ILIKE $" + strconv.Itoa(argIndex),
			"c.name ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := "SELECT COUNT(DISTINCT t.id) " + queryBase + whereClause
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deal transactions: %w", err)
	}

	query := `
		SELECT DISTINCT
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
	` + queryBase + whereClause + `
		ORDER BY t.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query deal transactions: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan deal transaction: %w", err)
		}

		detail.ReverseTo = reverseTo
		detail.EmployeeID = employeeID
		detail.EmployeeFirstName = employeeFirstName
		detail.EmployeeLastName = employeeLastName
		detail.CategoryID = categoryID
		detail.CategoryName = categoryName

		if categoryID != nil && categoryName != nil && categoryDirection != nil {
			detail.Category = &TransactionCategory{
				ID:          *categoryID,
				Name:        *categoryName,
				Description: categoryDescription,
				Direction:   *categoryDirection,
				CreatedAt:   *categoryCreatedAt,
				UpdatedAt:   *categoryUpdatedAt,
				Slug:        *categorySlug,
			}
		}

		if employeeID != nil && employeeFirstName != nil {
			fullEmployee, err := r.employeesRepository.GetEmployeeByID(ctx, *employeeID)
			if err == nil {
				detail.Employee = fullEmployee
			} else {
				detail.Employee = &hrm.EmployeeDetail{
					EmployeeListItem: hrm.EmployeeListItem{
						Employee: hrm.Employee{
							ID:        *employeeID,
							FirstName: *employeeFirstName,
							LastName:  employeeLastName,
							Status:    hrm.EmployeeStatusActive,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
				}
			}
		}

		transactions = append(transactions, detail)
	}

	return transactions, total, nil
}

// GetDealTransactionsSummary возвращает сводку по транзакциям сделки
func (r *Repository) GetDealTransactionsSummary(ctx context.Context, dealID string, filters GetTransactionsRequest) (*DealTransactionsSummary, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	args = append(args, dealID)
	whereConditions = append(whereConditions, "dt.deal_id = $1")
	argIndex++

	if filters.StartDate != nil {
		whereConditions = append(whereConditions, "t.created_at >= $"+strconv.Itoa(argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		whereConditions = append(whereConditions, "t.created_at <= $"+strconv.Itoa(argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	if filters.Status != nil {
		whereConditions = append(whereConditions, "t.status = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.CategoryID != nil {
		whereConditions = append(whereConditions, "c.id = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.CategoryID)
		argIndex++
	}

	if filters.EmployeeID != nil {
		whereConditions = append(whereConditions, "e.id = $"+strconv.Itoa(argIndex))
		args = append(args, *filters.EmployeeID)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		searchConditions := []string{
			"t.comment ILIKE $" + strconv.Itoa(argIndex),
			"e.first_name ILIKE $" + strconv.Itoa(argIndex),
			"e.last_name ILIKE $" + strconv.Itoa(argIndex),
			"c.name ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE -t.base_amount END), 0) as total_amount,
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE 0 END), 0) as income_amount,
			COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) as expense_amount,
			COALESCE(COUNT(CASE WHEN t.direction = 'income' THEN 1 END), 0) as income_count,
			COALESCE(COUNT(CASE WHEN t.direction = 'expense' THEN 1 END), 0) as expense_count,
			COUNT(t.id) as total_count
		FROM transactions t
		INNER JOIN deals_transactions dt ON t.id = dt.transaction_id
		LEFT JOIN transaction_employee te ON t.id = te.transaction_id
		LEFT JOIN employees e ON te.employee_id = e.id
		LEFT JOIN transaction_category tc ON t.id = tc.transaction_id
		LEFT JOIN transaction_categories c ON tc.category_id = c.id
	` + whereClause

	var summary DealTransactionsSummary
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&summary.TotalAmount,
		&summary.IncomeAmount,
		&summary.ExpenseAmount,
		&summary.IncomeCount,
		&summary.ExpenseCount,
		&summary.TotalCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get deal transactions summary: %w", err)
	}

	return &summary, nil
}

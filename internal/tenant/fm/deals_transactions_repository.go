package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/currency"
	"kroncl-server/internal/tenant/hrm"
	"strconv"
	"strings"
	"time"
)

func (r *Repository) CreateDealTransaction(ctx context.Context, dealID string, req CreateTransactionRequest) (*TransactionDetail, error) {
	// Проверяем существование сделки
	var dealExists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM deals WHERE id = $1)`, dealID).Scan(&dealExists)
	if err != nil || !dealExists {
		return nil, fmt.Errorf("deal not found: %s", dealID)
	}

	// Если категория не передана — подставляем авто-категорию сделки
	if req.CategoryID == "" {
		categorySlug := DEAL_TRANSACTION_CATEGORY_EXPENSE_SLUG
		if req.Direction == TransactionDirectionIncome {
			categorySlug = DEAL_TRANSACTION_CATEGORY_INCOME_SLUG
		}
		var categoryID string
		err := r.pool.QueryRow(ctx, `SELECT id FROM transaction_categories WHERE slug = $1`, categorySlug).Scan(&categoryID)
		if err != nil {
			return nil, fmt.Errorf("deal category not found by slug: %s", categorySlug)
		}
		req.CategoryID = categoryID
	}

	// Создаём транзакцию через общий метод
	detail, err := r.CreateTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	// Привязываем к сделке
	_, err = r.pool.Exec(ctx, `
		INSERT INTO deals_transactions (id, deal_id, transaction_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, dealID, detail.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to link deal: %w", err)
	}

	return detail, nil
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

func (r *Repository) GetDealTransactionsSummary(ctx context.Context, dealID string, filters GetTransactionsRequest, targetCurrency string) (*DealTransactionsSummary, error) {
	if targetCurrency == "" {
		targetCurrency = "RUB"
	}

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
			t.currency,
			DATE_TRUNC('hour', t.created_at) as date_hour,
			COALESCE(SUM(CASE WHEN t.direction = 'income' THEN t.base_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN t.direction = 'expense' THEN t.base_amount ELSE 0 END), 0) as expense,
			COALESCE(COUNT(CASE WHEN t.direction = 'income' THEN 1 END), 0) as income_count,
			COALESCE(COUNT(CASE WHEN t.direction = 'expense' THEN 1 END), 0) as expense_count,
			COUNT(t.id) as total_count
		FROM transactions t
		INNER JOIN deals_transactions dt ON t.id = dt.transaction_id
		LEFT JOIN transaction_employee te ON t.id = te.transaction_id
		LEFT JOIN employees e ON te.employee_id = e.id
		LEFT JOIN transaction_category tc ON t.id = tc.transaction_id
		LEFT JOIN transaction_categories c ON tc.category_id = c.id
	` + whereClause + `
		GROUP BY t.currency, DATE_TRUNC('hour', t.created_at)
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deal transactions: %w", err)
	}
	defer rows.Close()

	rawStats := make(map[string]map[time.Time]*currency.RawStat)
	var totalIncomeCount, totalExpenseCount, totalCount int64

	for rows.Next() {
		var currencyID string
		var dateHour time.Time
		var income, expense float64
		var incomeCount, expenseCount, count int64

		if err := rows.Scan(&currencyID, &dateHour, &income, &expense, &incomeCount, &expenseCount, &count); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}

		if rawStats[currencyID] == nil {
			rawStats[currencyID] = make(map[time.Time]*currency.RawStat)
		}
		rawStats[currencyID][dateHour] = &currency.RawStat{
			Income:  income,
			Expense: expense,
			Count:   count,
		}
		totalIncomeCount += incomeCount
		totalExpenseCount += expenseCount
		totalCount += count
	}

	converted, err := r.currencyService.ConvertSummary(ctx, rawStats, targetCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert summary: %w", err)
	}

	return &DealTransactionsSummary{
		TotalAmount:   converted.NetBalance,
		IncomeAmount:  converted.TotalIncome,
		ExpenseAmount: converted.TotalExpense,
		IncomeCount:   totalIncomeCount,
		ExpenseCount:  totalExpenseCount,
		TotalCount:    totalCount,
		Currency:      converted.Currency,
	}, nil
}

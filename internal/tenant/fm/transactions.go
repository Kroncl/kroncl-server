package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/tenant/hrm"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool                *pgxpool.Pool
	employeesRepository *hrm.Repository
}

func NewRepository(pool *pgxpool.Pool, employeesRepository *hrm.Repository) *Repository {
	return &Repository{pool: pool, employeesRepository: employeesRepository}
}

// инит транзакции
func (r *Repository) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*TransactionDetail, error) {
	// Валидация
	if req.BaseAmount <= 0 {
		return nil, fmt.Errorf("base_amount must be greater than 0")
	}

	if req.EmployeeID == "" {
		return nil, fmt.Errorf("employee_id is required")
	}

	// Проверяем соответствие знака направлению
	if req.Direction != TransactionDirectionIncome && req.Direction != TransactionDirectionExpense {
		return nil, fmt.Errorf("invalid transaction direction: %s", req.Direction)
	}

	// Валидация валюты
	switch req.Currency {
	case CurrencyRUB, CurrencyUSD, CurrencyEUR, CurrencyKZT:
		// OK
	default:
		return nil, fmt.Errorf("invalid currency: %s", req.Currency)
	}

	// Начинаем транзакцию БД
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// проверка сотрудника в той же транзакции
	employee, err := r.employeesRepository.GetEmployeeByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}
	_ = employee // используем если нужно

	// Подготавливаем данные
	comment := strings.TrimSpace(req.Comment)
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	// Статус по умолчанию - completed
	status := TransactionStatusCompleted
	if req.Status != "" {
		status = TransactionStatus(req.Status)
	}

	// 1. Создаем транзакцию
	transactionQuery := `
		INSERT INTO transactions (
			id, base_amount, currency, direction, status, comment,
			created_at, metadata
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5,
			CURRENT_TIMESTAMP, $6
		)
		RETURNING 
			id
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

	// 2. Создаем связь с сотрудником
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

	// 3. Если указана категория - создаем связь с категорией
	if req.CategoryID != "" {
		// Проверяем существование категории
		var exists bool
		checkCategoryQuery := `SELECT EXISTS(SELECT 1 FROM transaction_categories WHERE id = $1)`
		err = tx.QueryRow(ctx, checkCategoryQuery, req.CategoryID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("failed to check category: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("category not found: %s", req.CategoryID)
		}

		// Создаем связь с категорией
		linkCategoryQuery := `
			INSERT INTO transaction_category (
				id, transaction_id, category_id, created_at
			) VALUES (
				gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP
			)
		`
		_, err = tx.Exec(ctx, linkCategoryQuery, transactionID, req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to link category to transaction: %w", err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Возвращаем полную информацию о транзакции через GetTransactionByID
	return r.GetTransactionByID(ctx, transactionID)
}

// получение транзакции по ID
func (r *Repository) GetTransactionByID(ctx context.Context, id string) (*TransactionDetail, error) {
	query := `
		SELECT 
			t.id,
			t.base_amount,
			t.currency,
			t.direction,
			t.status,
			t.comment,
			t.created_at,
			t.metadata,
			-- Employee data
			te.employee_id,
			e.first_name,
			e.last_name,
			-- Category data
			tc.category_id,
			c.name as category_name,
			c.description as category_description,
			c.direction as category_direction,
			c.created_at as category_created_at,
			c.updated_at as category_updated_at,
			c.slug as category_slug
		FROM transactions t
		LEFT JOIN transaction_employee te ON t.id = te.transaction_id
		LEFT JOIN employees e ON te.employee_id = e.id
		LEFT JOIN transaction_category tc ON t.id = tc.transaction_id
		LEFT JOIN transaction_categories c ON tc.category_id = c.id
		WHERE t.id = $1
	`

	var detail TransactionDetail
	var employeeID, employeeFirstName, employeeLastName *string
	var categoryID, categoryName, categoryDescription *string
	var categoryDirection *TransactionCategoryDirection
	var categoryCreatedAt, categoryUpdatedAt *time.Time
	var categorySlug *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&detail.ID,
		&detail.BaseAmount,
		&detail.Currency,
		&detail.Direction,
		&detail.Status,
		&detail.Comment,
		&detail.CreatedAt,
		&detail.Metadata,
		// Employee
		&employeeID,
		&employeeFirstName,
		&employeeLastName,
		// Category
		&categoryID,
		&categoryName,
		&categoryDescription,
		&categoryDirection,
		&categoryCreatedAt,
		&categoryUpdatedAt,
		&categorySlug,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Заполняем Employee данные
	detail.EmployeeID = employeeID
	detail.EmployeeFirstName = employeeFirstName
	detail.EmployeeLastName = employeeLastName

	if employeeID != nil && employeeFirstName != nil {
		// Получаем полные данные сотрудника через репозиторий
		fullEmployee, err := r.employeesRepository.GetEmployeeByID(ctx, *employeeID)
		if err == nil {
			detail.Employee = fullEmployee
		} else {
			// fallback на то, что есть
			detail.Employee = &hrm.EmployeeDetail{
				EmployeeListItem: hrm.EmployeeListItem{
					Employee: hrm.Employee{
						ID:        *employeeID,
						FirstName: *employeeFirstName,
						LastName:  employeeLastName,
						Status:    hrm.EmployeeStatusActive,
						CreatedAt: time.Now(), // или дефолт
						UpdatedAt: time.Now(),
					},
				},
			}
		}
	}

	// Заполняем Category данные
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

	return &detail, nil
}

// получение списка транзакций с фильтрацией
func (r *Repository) GetTransactions(ctx context.Context, offset, limit int, filters GetTransactionsRequest) ([]TransactionDetail, int64, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Базовый запрос
	queryBase := `
		FROM transactions t
		LEFT JOIN transaction_employee te ON t.id = te.transaction_id
		LEFT JOIN employees e ON te.employee_id = e.id
		LEFT JOIN transaction_category tc ON t.id = tc.transaction_id
		LEFT JOIN transaction_categories c ON tc.category_id = c.id
	`

	// Фильтры
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

	// WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(DISTINCT t.id) " + queryBase + whereClause
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Основной запрос с пагинацией
	query := `
		SELECT DISTINCT
			t.id,
			t.base_amount,
			t.currency,
			t.direction,
			t.status,
			t.comment,
			t.created_at,
			t.metadata,
			te.employee_id,
			e.first_name,
			e.last_name,
			tc.category_id,
			c.name,
			c.description,
			c.direction,
			c.created_at,
			c.updated_at
	` + queryBase + whereClause + `
		ORDER BY t.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []TransactionDetail
	for rows.Next() {
		var detail TransactionDetail
		var employeeID, employeeFirstName, employeeLastName *string
		var categoryID, categoryName, categoryDescription *string
		var categoryDirection *TransactionCategoryDirection
		var categoryCreatedAt, categoryUpdatedAt *time.Time

		err := rows.Scan(
			&detail.ID,
			&detail.BaseAmount,
			&detail.Currency,
			&detail.Direction,
			&detail.Status,
			&detail.Comment,
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
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}

		detail.EmployeeID = employeeID
		detail.EmployeeFirstName = employeeFirstName
		detail.EmployeeLastName = employeeLastName
		detail.CategoryID = categoryID
		detail.CategoryName = categoryName

		transactions = append(transactions, detail)
	}

	return transactions, total, nil
}

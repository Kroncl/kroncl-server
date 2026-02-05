package hrm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetEmployeeByID(ctx context.Context, id string) (*EmployeeWithAccount, error) {
	query := `
		SELECT 
			e.id,
			e.first_name,
			e.last_name,
			e.email,
			e.phone,
			e.status,
			e.created_at,
			e.updated_at,
			ea.account_id,
			ea.created_at as linked_at
		FROM employees e
		LEFT JOIN employee_account ea ON e.id = ea.employee_id
		WHERE e.id = $1
	`

	var employee EmployeeWithAccount
	var accountID *string
	var linkedAt *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&employee.ID,
		&employee.FirstName,
		&employee.LastName,
		&employee.Email,
		&employee.Phone,
		&employee.Status,
		&employee.CreatedAt,
		&employee.UpdatedAt,
		&accountID,
		&linkedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	employee.AccountID = accountID
	employee.LinkedAt = linkedAt
	employee.IsAccountLinked = accountID != nil

	return &employee, nil
}

func (r *Repository) GetEmployees(ctx context.Context, offset, limit int) ([]EmployeeWithAccount, int, error) {
	// Сначала получаем общее количество
	countQuery := `SELECT COUNT(*) FROM employees`
	var total int
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	// Получаем сотрудников с пагинацией
	query := `
		SELECT 
			e.id,
			e.first_name,
			e.last_name,
			e.email,
			e.phone,
			e.status,
			e.created_at,
			e.updated_at,
			ea.account_id,
			ea.created_at as linked_at
		FROM employees e
		LEFT JOIN employee_account ea ON e.id = ea.employee_id
		ORDER BY e.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query employees: %w", err)
	}
	defer rows.Close()

	var employees []EmployeeWithAccount
	for rows.Next() {
		var employee EmployeeWithAccount
		var accountID *string
		var linkedAt *time.Time

		err := rows.Scan(
			&employee.ID,
			&employee.FirstName,
			&employee.LastName,
			&employee.Email,
			&employee.Phone,
			&employee.Status,
			&employee.CreatedAt,
			&employee.UpdatedAt,
			&accountID,
			&linkedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee: %w", err)
		}

		employee.AccountID = accountID
		employee.LinkedAt = linkedAt
		employee.IsAccountLinked = accountID != nil
		employees = append(employees, employee)
	}

	return employees, total, nil
}

func (r *Repository) UpdateEmployee(ctx context.Context, id string, req UpdateEmployeeRequest) (*EmployeeWithAccount, error) {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Собираем поля для обновления
	var updates []string
	var params []interface{}
	paramCounter := 1

	// Тримим и проверяем поля
	if req.FirstName != "" {
		firstName := strings.TrimSpace(req.FirstName)
		if firstName != "" && len(firstName) >= 2 {
			updates = append(updates, fmt.Sprintf("first_name = $%d", paramCounter))
			params = append(params, firstName)
			paramCounter++
		}
	}

	if req.LastName != "" {
		lastName := strings.TrimSpace(req.LastName)
		updates = append(updates, fmt.Sprintf("last_name = $%d", paramCounter))
		params = append(params, lastName)
		paramCounter++
	}

	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		if email != "" && strings.Contains(email, "@") {
			updates = append(updates, fmt.Sprintf("email = $%d", paramCounter))
			params = append(params, email)
			paramCounter++
		}
	}

	if req.Phone != "" {
		phone := strings.TrimSpace(req.Phone)
		updates = append(updates, fmt.Sprintf("phone = $%d", paramCounter))
		params = append(params, phone)
		paramCounter++
	}

	if req.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", paramCounter))
		params = append(params, req.Status)
		paramCounter++
	}

	// Если нет полей для обновления
	if len(updates) == 0 {
		return r.GetEmployeeByID(ctx, id)
	}

	// Добавляем updated_at
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// Добавляем ID в параметры
	params = append(params, id)

	// Обновляем сотрудника
	updateQuery := fmt.Sprintf(`
		UPDATE employees 
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), paramCounter)

	_, err = tx.Exec(ctx, updateQuery, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Получаем обновленного сотрудника
	return r.GetEmployeeByID(ctx, id)
}

func (r *Repository) getEmployeeByIDTx(ctx context.Context, tx pgx.Tx, id string) (*Employee, error) {
	query := `
		SELECT id, first_name, last_name, email, phone,
			status, created_at, updated_at
		FROM employees 
		WHERE id = $1
	`

	var employee Employee
	err := tx.QueryRow(ctx, query, id).Scan(
		&employee.ID,
		&employee.FirstName,
		&employee.LastName,
		&employee.Email,
		&employee.Phone,
		&employee.Status,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get employee in transaction: %w", err)
	}

	return &employee, nil
}

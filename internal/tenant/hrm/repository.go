package hrm

import (
	"context"
	"fmt"
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/core"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool             *pgxpool.Pool
	accountsService  *accounts.Service
	companiesService *companies.Service
}

func NewRepository(pool *pgxpool.Pool, accountsService *accounts.Service, companiesService *companies.Service) *Repository {
	return &Repository{pool: pool, accountsService: accountsService, companiesService: companiesService}
}

// GetEmployeeByID возвращает детальную информацию
func (r *Repository) GetEmployeeByID(ctx context.Context, id string) (*EmployeeDetail, error) {
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

	var employee EmployeeDetail
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

func (r *Repository) GetEmployees(ctx context.Context, offset, limit int, search string) ([]EmployeeListItem, int, error) {
	// Собираем условия WHERE
	var whereClause string
	var args []interface{}
	argIndex := 1

	if search != "" {
		whereClause = `WHERE 
			e.first_name ILIKE $` + strconv.Itoa(argIndex) + ` OR 
			e.last_name ILIKE $` + strconv.Itoa(argIndex) + ` OR 
			e.email ILIKE $` + strconv.Itoa(argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Сначала получаем общее количество
	countQuery := `SELECT COUNT(*) FROM employees e ` + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
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
		` + whereClause + `
		ORDER BY e.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query employees: %w", err)
	}
	defer rows.Close()

	var employees []EmployeeListItem
	for rows.Next() {
		var employee EmployeeListItem
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

func (r *Repository) UpdateEmployee(ctx context.Context, id string, req UpdateEmployeeRequest) (*EmployeeDetail, error) {
	updater := core.NewUpdater("employees")

	// Тримминг и валидация отдельно
	firstName := strings.TrimSpace(req.FirstName)
	if firstName != "" && len(firstName) >= 2 {
		updater.SetString("first_name", firstName)
	}

	lastName := strings.TrimSpace(req.LastName)
	if lastName == "" {
		updater.SetNull("last_name")
	} else {
		updater.SetString("last_name", lastName)
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		updater.SetNull("email")
	} else if strings.Contains(email, "@") {
		updater.SetString("email", email)
	}

	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		updater.SetNull("phone")
	} else {
		updater.SetString("phone", phone)
	}

	if req.Status != "" {
		updater.SetString("status", string(req.Status))
	}

	// Если нет изменений - возвращаем текущие данные
	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetEmployeeByID(ctx, id)
	}

	// Выполняем обновление
	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	return r.GetEmployeeByID(ctx, id)
}

func (r *Repository) DeleteEmployee(ctx context.Context, id string) (bool, error) {
	query := `DELETE FROM employees WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)

	if err != nil {
		return false, fmt.Errorf("failed to delete employee: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return rowsAffected > 0, nil
}

func (r *Repository) CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (*Employee, error) {
	// Валидация минимальных требований
	firstName := strings.TrimSpace(req.FirstName)
	if firstName == "" || len(firstName) < 2 {
		return nil, fmt.Errorf("first name is required and must be at least 2 characters")
	}

	// Генерация ID
	id := uuid.New().String()

	// Подготавливаем данные
	lastName := strings.TrimSpace(req.LastName)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	// Если email пустой, устанавливаем NULL
	var emailPtr *string
	if email != "" && strings.Contains(email, "@") {
		emailPtr = &email
	}

	// Если phone пустой, устанавливаем NULL
	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}

	query := `
		INSERT INTO employees (
			id, first_name, last_name, email, phone, status,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, first_name, last_name, email, phone,
			status, created_at, updated_at
	`

	var employee Employee
	err := r.pool.QueryRow(ctx, query,
		id,
		firstName,
		core.NullIfEmpty(lastName),
		emailPtr,
		phonePtr,
		EmployeeStatusActive,
	).Scan(
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
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	return &employee, nil
}

// LinkAccount связывает аккаунт с сотрудником
func (r *Repository) LinkAccount(ctx context.Context, employeeId, accountId string) (*EmployeeDetail, error) {
	// Проверяем, существует ли сотрудник
	_, err := r.GetEmployeeByID(ctx, employeeId)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// Вставляем или обновляем связь
	query := `
		INSERT INTO employee_account (employee_id, account_id, created_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (employee_id) 
		DO UPDATE SET 
			account_id = EXCLUDED.account_id,
			created_at = CURRENT_TIMESTAMP
		RETURNING created_at
	`

	var linkedAt time.Time
	err = r.pool.QueryRow(ctx, query, employeeId, accountId).Scan(&linkedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to link account: %w", err)
	}

	// Возвращаем обновленные данные сотрудника
	return r.GetEmployeeByID(ctx, employeeId)
}

// UnlinkAccount отвязывает аккаунт от сотрудника
func (r *Repository) UnlinkAccount(ctx context.Context, employeeId string) (*EmployeeDetail, error) {
	// Проверяем, существует ли сотрудник
	_, err := r.GetEmployeeByID(ctx, employeeId)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// Удаляем связь
	query := `DELETE FROM employee_account WHERE employee_id = $1`
	_, err = r.pool.Exec(ctx, query, employeeId)
	if err != nil {
		return nil, fmt.Errorf("failed to unlink account: %w", err)
	}

	// Возвращаем обновленные данные сотрудника
	return r.GetEmployeeByID(ctx, employeeId)
}

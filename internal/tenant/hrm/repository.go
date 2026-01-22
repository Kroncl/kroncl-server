package hrm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	if req.FirstName == "" || len(req.FirstName) < 2 {
		return nil, fmt.Errorf("first_name is required and must be at least 2 characters")
	}

	if req.Email != "" && !strings.Contains(req.Email, "@") {
		return nil, fmt.Errorf("invalid email format")
	}

	// Генерируем ID
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	// Чистим данные
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	query := `
		INSERT INTO employees (
			id, first_name, last_name, email, phone, 
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, first_name, last_name, email, phone,
			status, created_at, updated_at
	`

	var employee Employee
	err = r.pool.QueryRow(
		ctx,
		query,
		id,
		firstName,
		lastName,
		email,
		phone,
		EmployeeStatusActive,
		time.Now(),
		time.Now(),
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
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &employee, nil
}

func (r *Repository) GetEmployeeByID(ctx context.Context, id string) (*Employee, error) {
	query := `
		SELECT id, first_name, last_name, email, phone,
			status, created_at, updated_at
		FROM employees 
		WHERE id = $1
	`

	var employee Employee
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	return &employee, nil
}

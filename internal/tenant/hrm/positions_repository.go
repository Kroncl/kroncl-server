package hrm

import (
	"context"
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreatePosition создаёт новую должность
func (r *Repository) CreatePosition(ctx context.Context, req CreatePositionRequest) (*Position, error) {
	// Убираем дубликаты разрешений
	uniquePerms := config.UniquePermissions(req.Permissions)

	// Проверяем разрешения
	if invalid := config.ValidatePermissions(uniquePerms); len(invalid) > 0 {
		return nil, fmt.Errorf("invalid permissions: %v", invalid)
	}

	id := uuid.New().String()

	permissionsJSON, err := json.Marshal(uniquePerms)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		INSERT INTO employees_positions (id, name, description, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, permissions, created_at, updated_at
	`

	var pos Position
	var description *string
	var returnedPermissionsJSON []byte

	err = r.pool.QueryRow(ctx, query,
		id,
		req.Name,
		req.Description,
		permissionsJSON,
	).Scan(
		&pos.ID,
		&pos.Name,
		&description,
		&returnedPermissionsJSON,
		&pos.CreatedAt,
		&pos.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	pos.Description = description

	// Парсим permissions из JSONB
	if len(returnedPermissionsJSON) > 0 {
		if err := json.Unmarshal(returnedPermissionsJSON, &pos.Permissions); err != nil {
			return nil, fmt.Errorf("failed to parse permissions: %w", err)
		}
	} else {
		pos.Permissions = []string{}
	}

	return &pos, nil
}

// GetPositionByID возвращает должность по ID
func (r *Repository) GetPositionByID(ctx context.Context, id string) (*Position, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM employees_positions
		WHERE id = $1
	`

	var pos Position
	var description *string
	var permissionsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pos.ID,
		&pos.Name,
		&description,
		&permissionsJSON,
		&pos.CreatedAt,
		&pos.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("position not found")
		}
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	pos.Description = description

	// Парсим permissions из JSONB
	if len(permissionsJSON) > 0 {
		if err := json.Unmarshal(permissionsJSON, &pos.Permissions); err != nil {
			return nil, fmt.Errorf("failed to parse permissions: %w", err)
		}
	} else {
		pos.Permissions = []string{}
	}

	return &pos, nil
}

// GetPositions возвращает список должностей с пагинацией
func (r *Repository) GetPositions(ctx context.Context, page, limit int, search string) ([]Position, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var args []interface{}
	var conditions []string
	argIndex := 1

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := "SELECT COUNT(*) FROM employees_positions " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count positions: %w", err)
	}

	// Data
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM employees_positions
	` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var positions []Position
	for rows.Next() {
		var pos Position
		var description *string
		var permissionsJSON []byte

		err := rows.Scan(
			&pos.ID,
			&pos.Name,
			&description,
			&permissionsJSON,
			&pos.CreatedAt,
			&pos.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan position: %w", err)
		}

		pos.Description = description

		if len(permissionsJSON) > 0 {
			if err := json.Unmarshal(permissionsJSON, &pos.Permissions); err != nil {
				return nil, 0, fmt.Errorf("failed to parse permissions: %w", err)
			}
		} else {
			pos.Permissions = []string{}
		}

		positions = append(positions, pos)
	}

	return positions, total, nil
}

// UpdatePosition обновляет должность
func (r *Repository) UpdatePosition(ctx context.Context, id string, req UpdatePositionRequest) (*Position, error) {
	// Проверяем существование
	_, err := r.GetPositionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updater := core.NewUpdater("employees_positions")

	if req.Name != nil && *req.Name != "" {
		updater.SetString("name", *req.Name)
	}

	if req.Description != nil {
		if *req.Description == "" {
			updater.SetNull("description")
		} else {
			updater.SetString("description", *req.Description)
		}
	}

	if req.Permissions != nil {
		uniquePerms := config.UniquePermissions(req.Permissions)
		if invalid := config.ValidatePermissions(uniquePerms); len(invalid) > 0 {
			return nil, fmt.Errorf("invalid permissions: %v", invalid)
		}

		permissionsJSON, err := json.Marshal(uniquePerms)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal permissions: %w", err)
		}
		updater.Set("permissions", permissionsJSON)
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetPositionByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	return r.GetPositionByID(ctx, id)
}

// DeletePosition удаляет должность
func (r *Repository) DeletePosition(ctx context.Context, id string) error {
	// Проверяем, есть ли сотрудники, привязанные к этой должности
	var count int
	checkQuery := `SELECT COUNT(*) FROM employee_position WHERE position_id = $1`
	err := r.pool.QueryRow(ctx, checkQuery, id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check employee_position: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete position: %d employee(s) assigned to it", count)
	}

	query := `DELETE FROM employees_positions WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("position not found")
	}

	return nil
}

// CheckPositionExists проверяет существование должности по ID
func (r *Repository) CheckPositionExists(ctx context.Context, positionID string) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM employees_positions WHERE id = $1)`
	err := r.pool.QueryRow(ctx, query, positionID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check position existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("position not found")
	}
	return nil
}

// getPositionPermissions возвращает разрешения из должностей сотрудника
func (r *Repository) getPositionPermissions(ctx context.Context, accountID string) ([]string, error) {
	// 1. Находим employee по account_id
	var employeeID string
	employeeQuery := `
		SELECT employee_id
		FROM employee_account
		WHERE account_id = $1
	`
	err := r.pool.QueryRow(ctx, employeeQuery, accountID).Scan(&employeeID)
	if err != nil {
		// Если сотрудник не найден - возвращаем пустой массив
		return []string{}, nil
	}

	// 2. Получаем все должности сотрудника
	positionsQuery := `
		SELECT p.permissions
		FROM employee_position ep
		INNER JOIN employees_positions p ON ep.position_id = p.id
		WHERE ep.employee_id = $1
	`

	rows, err := r.pool.Query(ctx, positionsQuery, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var allPermissions []string
	for rows.Next() {
		var permissionsJSON []byte
		if err := rows.Scan(&permissionsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan permissions: %w", err)
		}

		var perms []string
		if len(permissionsJSON) > 0 {
			if err := json.Unmarshal(permissionsJSON, &perms); err != nil {
				return nil, fmt.Errorf("failed to parse permissions: %w", err)
			}
		}
		allPermissions = append(allPermissions, perms...)
	}

	return allPermissions, nil
}

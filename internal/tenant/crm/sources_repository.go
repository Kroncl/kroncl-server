package crm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ClientSourceExists проверяет существование источника по ID
func (r *Repository) ClientSourceExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM client_sources WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check client source existence: %w", err)
	}

	return exists, nil
}

// GetClientSourceByID возвращает источник по ID
func (r *Repository) GetClientSourceByID(ctx context.Context, id string) (*ClientSource, error) {
	query := `
		SELECT 
			id, name, url, type, comment, system, status, metadata, created_at, updated_at
		FROM client_sources
		WHERE id = $1
	`

	var source ClientSource
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Type,
		&source.Comment,
		&source.System,
		&source.Status,
		&source.Metadata,
		&source.CreatedAt,
		&source.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get client source: %w", err)
	}

	return &source, nil
}

// GetClientSourceByName возвращает источник по имени (для проверки уникальности)
func (r *Repository) GetClientSourceByName(ctx context.Context, name string) (*ClientSource, error) {
	query := `
		SELECT 
			id, name, url, type, comment, system, status, metadata, created_at, updated_at
		FROM client_sources
		WHERE name = $1
	`

	var source ClientSource
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Type,
		&source.Comment,
		&source.System,
		&source.Status,
		&source.Metadata,
		&source.CreatedAt,
		&source.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get client source by name: %w", err)
	}

	return &source, nil
}

// GetClientSources возвращает список источников с пагинацией и фильтрацией
func (r *Repository) GetClientSources(ctx context.Context, req GetSourcesRequest) ([]ClientSource, int, error) {
	var whereClause string
	var args []interface{}
	var whereConditions []string
	argIndex := 1

	// Вычисляем offset
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	if req.Type != nil {
		whereConditions = append(whereConditions, "type = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Type)
		argIndex++
	}

	if req.Status != nil {
		whereConditions = append(whereConditions, "status = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.System != nil {
		whereConditions = append(whereConditions, "system = $"+strconv.Itoa(argIndex))
		args = append(args, *req.System)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		searchConditions := []string{
			"name ILIKE $" + strconv.Itoa(argIndex),
			"comment ILIKE $" + strconv.Itoa(argIndex),
			"url ILIKE $" + strconv.Itoa(argIndex),
		}
		whereConditions = append(whereConditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Получаем общее количество
	countQuery := `SELECT COUNT(*) FROM client_sources ` + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count client sources: %w", err)
	}

	// Получаем источники с пагинацией
	query := `
		SELECT 
			id, name, url, type, comment, system, status, metadata, created_at, updated_at
		FROM client_sources
		` + whereClause + `
		ORDER BY 
			system DESC, -- системные сверху
			name ASC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query client sources: %w", err)
	}
	defer rows.Close()

	var sources []ClientSource
	for rows.Next() {
		var source ClientSource
		err := rows.Scan(
			&source.ID,
			&source.Name,
			&source.URL,
			&source.Type,
			&source.Comment,
			&source.System,
			&source.Status,
			&source.Metadata,
			&source.CreatedAt,
			&source.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan client source: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, total, nil
}

// CreateClientSource создает новый источник
func (r *Repository) CreateClientSource(ctx context.Context, req CreateSourceRequest) (*ClientSource, error) {
	// Валидация
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("source name is required")
	}

	// Проверяем уникальность имени
	existing, _ := r.GetClientSourceByName(ctx, name)
	if existing != nil {
		return nil, fmt.Errorf("source with name '%s' already exists", name)
	}

	var urlPtr *string
	if req.URL != nil {
		trimmedURL := strings.TrimSpace(*req.URL)
		if trimmedURL != "" {
			urlPtr = &trimmedURL
		}
	}

	var commentPtr *string
	if req.Comment != nil {
		trimmedComment := strings.TrimSpace(*req.Comment)
		if trimmedComment != "" {
			commentPtr = &trimmedComment
		}
	}

	// Устанавливаем статус по умолчанию, если не указан
	status := req.Status
	if status == "" {
		status = SourceStatusActive
	}

	id := uuid.New().String()

	query := `
		INSERT INTO client_sources (
			id, name, url, type, comment, system, status, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, name, url, type, comment, system, status, metadata, created_at, updated_at
	`

	var source ClientSource
	err := r.pool.QueryRow(ctx, query,
		id,
		name,
		urlPtr,
		req.Type,
		commentPtr,
		req.System,
		status,
		req.Metadata,
	).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Type,
		&source.Comment,
		&source.System,
		&source.Status,
		&source.Metadata,
		&source.CreatedAt,
		&source.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create client source: %w", err)
	}

	return &source, nil
}

// UpdateClientSource обновляет источник
func (r *Repository) UpdateClientSource(ctx context.Context, id string, req UpdateSourceRequest) (*ClientSource, error) {
	// Проверяем существование и системный флаг
	existing, err := r.GetClientSourceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("source not found: %w", err)
	}

	// Запрещаем изменение системных источников полностью
	if existing.System {
		return nil, fmt.Errorf("cannot update system source")
	}

	updater := core.NewUpdater("client_sources")

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("source name cannot be empty")
		}

		// Проверяем уникальность имени, если оно меняется
		if name != existing.Name {
			existingWithName, _ := r.GetClientSourceByName(ctx, name)
			if existingWithName != nil && existingWithName.ID != id {
				return nil, fmt.Errorf("source with name '%s' already exists", name)
			}
		}

		updater.SetString("name", name)
	}

	if req.URL != nil {
		if *req.URL == "" {
			updater.SetNull("url")
		} else {
			url := strings.TrimSpace(*req.URL)
			updater.SetString("url", url)
		}
	}

	if req.Type != nil {
		updater.SetString("type", string(*req.Type))
	}

	if req.Comment != nil {
		if *req.Comment == "" {
			updater.SetNull("comment")
		} else {
			comment := strings.TrimSpace(*req.Comment)
			updater.SetString("comment", comment)
		}
	}

	if req.Status != nil {
		updater.SetString("status", string(*req.Status))
	}

	if req.Metadata != nil {
		updater.SetJSONB("metadata", *req.Metadata)
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetClientSourceByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update client source: %w", err)
	}

	return r.GetClientSourceByID(ctx, id)
}

// ActivateClientSource активирует источник
func (r *Repository) ActivateClientSource(ctx context.Context, id string) (*ClientSource, error) {
	// Проверяем существование
	exists, err := r.ClientSourceExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check source existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("source not found")
	}

	query := `
		UPDATE client_sources 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
		RETURNING 
			id, name, url, type, comment, system, status, metadata, created_at, updated_at
	`

	var source ClientSource
	err = r.pool.QueryRow(ctx, query, SourceStatusActive, id).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Type,
		&source.Comment,
		&source.System,
		&source.Status,
		&source.Metadata,
		&source.CreatedAt,
		&source.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to activate source: %w", err)
	}

	return &source, nil
}

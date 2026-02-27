package crm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ClientExists проверяет существование клиента по ID
func (r *Repository) ClientExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check client existence: %w", err)
	}

	return exists, nil
}

// GetClientByID возвращает клиента по ID с его источником
func (r *Repository) GetClientByID(ctx context.Context, id string) (*ClientDetail, error) {
	query := `
		SELECT 
			c.id, c.first_name, c.last_name, c.patronymic, c.phone, c.email, c.comment, 
			c.type, c.status, c.metadata, c.created_at, c.updated_at,
			cs.id, cs.name, cs.url, cs.type, cs.comment, cs.system, cs.status, cs.metadata, cs.created_at, cs.updated_at
		FROM clients c
		INNER JOIN client_source csl ON c.id = csl.client_id
		INNER JOIN client_sources cs ON csl.source_id = cs.id
		WHERE c.id = $1
	`

	var client Client
	var source ClientSource

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&client.ID,
		&client.FirstName,
		&client.LastName,
		&client.Patronymic,
		&client.Phone,
		&client.Email,
		&client.Comment,
		&client.Type,
		&client.Status,
		&client.Metadata,
		&client.CreatedAt,
		&client.UpdatedAt,
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
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &ClientDetail{
		Client: client,
		Source: source, // теперь Source, не Sources
	}, nil
}

// GetClients возвращает список клиентов с пагинацией, фильтрацией и их источниками
func (r *Repository) GetClients(ctx context.Context, req GetClientsRequest) ([]ClientDetail, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	// Вычисляем offset
	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	// Добавляем фильтр по источнику если есть (должен быть ПЕРВЫМ)
	if req.SourceID != nil && *req.SourceID != "" {
		conditions = append(conditions, "csl.source_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SourceID)
		argIndex++
	}

	// Условия для clients
	if req.Type != nil {
		conditions = append(conditions, "c.type = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Type)
		argIndex++
	}

	if req.Status != nil {
		conditions = append(conditions, "c.status = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		// Только текстовые поля, исключаем UUID
		searchConditions := []string{
			"c.first_name ILIKE $" + strconv.Itoa(argIndex),
			"c.last_name ILIKE $" + strconv.Itoa(argIndex),
			"c.patronymic ILIKE $" + strconv.Itoa(argIndex),
			"c.phone ILIKE $" + strconv.Itoa(argIndex),
			"c.email ILIKE $" + strconv.Itoa(argIndex),
			"c.comment ILIKE $" + strconv.Itoa(argIndex),
		}
		conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	// Базовый запрос с JOIN
	baseQuery := `
		FROM clients c
		INNER JOIN client_source csl ON c.id = csl.client_id
		INNER JOIN client_sources cs ON csl.source_id = cs.id
	`

	// Формируем WHERE
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(*) " + baseQuery + " " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Получаем клиентов с пагинацией
	query := `
		SELECT 
			c.id, c.first_name, c.last_name, c.patronymic, c.phone, c.email, c.comment, 
			c.type, c.status, c.metadata, c.created_at, c.updated_at,
			cs.id, cs.name, cs.url, cs.type, cs.comment, cs.system, cs.status, cs.metadata, cs.created_at, cs.updated_at
	` + baseQuery + " " + whereClause + `
		ORDER BY c.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clientDetails []ClientDetail
	for rows.Next() {
		var client Client
		var source ClientSource

		err := rows.Scan(
			&client.ID,
			&client.FirstName,
			&client.LastName,
			&client.Patronymic,
			&client.Phone,
			&client.Email,
			&client.Comment,
			&client.Type,
			&client.Status,
			&client.Metadata,
			&client.CreatedAt,
			&client.UpdatedAt,
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
			return nil, 0, fmt.Errorf("failed to scan client: %w", err)
		}

		clientDetails = append(clientDetails, ClientDetail{
			Client: client,
			Source: source,
		})
	}

	return clientDetails, total, nil
}

// CreateClient создает нового клиента с обязательным источником
func (r *Repository) CreateClient(ctx context.Context, req CreateClientRequest, sourceID string) (*ClientDetail, error) {
	// Валидация
	if strings.TrimSpace(req.FirstName) == "" {
		return nil, fmt.Errorf("first name is required")
	}

	// Проверяем существование источника
	sourceExists, err := r.ClientSourceExists(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check source existence: %w", err)
	}
	if !sourceExists {
		return nil, fmt.Errorf("source with id '%s' not found", sourceID)
	}

	// Устанавливаем статус по умолчанию
	status := req.Status
	if status == "" {
		status = ClientStatusActive
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	clientID := uuid.New().String()

	// Вставляем клиента
	clientQuery := `
		INSERT INTO clients (
			id, first_name, last_name, patronymic, phone, email, comment, type, status, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, first_name, last_name, patronymic, phone, email, comment, type, status, metadata, created_at, updated_at
	`

	var client Client
	err = tx.QueryRow(ctx, clientQuery,
		clientID,
		req.FirstName,
		req.LastName,
		req.Patronymic,
		req.Phone,
		req.Email,
		req.Comment,
		req.Type,
		status,
		req.Metadata,
	).Scan(
		&client.ID,
		&client.FirstName,
		&client.LastName,
		&client.Patronymic,
		&client.Phone,
		&client.Email,
		&client.Comment,
		&client.Type,
		&client.Status,
		&client.Metadata,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Создаем связь с источником
	linkQuery := `
		INSERT INTO client_source (id, client_id, source_id, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`

	linkID := uuid.New().String()
	_, err = tx.Exec(ctx, linkQuery, linkID, clientID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to create client-source link: %w", err)
	}

	// Получаем информацию об источнике
	sourceQuery := `
		SELECT 
			id, name, url, type, comment, system, status, metadata, created_at, updated_at
		FROM client_sources
		WHERE id = $1
	`

	var source ClientSource
	err = tx.QueryRow(ctx, sourceQuery, sourceID).Scan(
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
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	// Коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &ClientDetail{
		Client: client,
		Source: source, // теперь Source, не Sources
	}, nil
}

// UpdateClient обновляет клиента и опционально его источник
func (r *Repository) UpdateClient(ctx context.Context, id string, req UpdateClientRequest) (*ClientDetail, error) {
	// Проверяем существование клиента
	exists, err := r.ClientExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check client existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем данные клиента
	updater := core.NewUpdater("clients")

	if req.FirstName != nil {
		firstName := strings.TrimSpace(*req.FirstName)
		if firstName == "" {
			return nil, fmt.Errorf("first name cannot be empty")
		}
		updater.SetString("first_name", firstName)
	}

	if req.LastName != nil {
		if *req.LastName == "" {
			updater.SetNull("last_name")
		} else {
			lastName := strings.TrimSpace(*req.LastName)
			updater.SetString("last_name", lastName)
		}
	}

	if req.Patronymic != nil {
		if *req.Patronymic == "" {
			updater.SetNull("patronymic")
		} else {
			patronymic := strings.TrimSpace(*req.Patronymic)
			updater.SetString("patronymic", patronymic)
		}
	}

	if req.Phone != nil {
		if *req.Phone == "" {
			updater.SetNull("phone")
		} else {
			phone := strings.TrimSpace(*req.Phone)
			updater.SetString("phone", phone)
		}
	}

	if req.Email != nil {
		if *req.Email == "" {
			updater.SetNull("email")
		} else {
			email := strings.TrimSpace(*req.Email)
			updater.SetString("email", email)
		}
	}

	if req.Comment != nil {
		if *req.Comment == "" {
			updater.SetNull("comment")
		} else {
			comment := strings.TrimSpace(*req.Comment)
			updater.SetString("comment", comment)
		}
	}

	if req.Type != nil {
		updater.SetString("type", string(*req.Type))
	}

	if req.Status != nil {
		updater.SetString("status", string(*req.Status))
	}

	if req.Metadata != nil {
		updater.SetJSONB("metadata", *req.Metadata)
	}

	// Применяем обновления клиента, если они есть
	if updater.HasChanges() {
		clientQuery, clientArgs := updater.Where("id = $1", id).Build()
		_, err = tx.Exec(ctx, clientQuery, clientArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to update client: %w", err)
		}
	}

	// Обновляем источник, если указан
	if req.SourceID != nil {
		// Проверяем существование нового источника
		sourceExists, err := r.ClientSourceExists(ctx, *req.SourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to check source existence: %w", err)
		}
		if !sourceExists {
			return nil, fmt.Errorf("source with id '%s' not found", *req.SourceID)
		}

		// Проверяем, есть ли уже связь
		var linkID string
		checkQuery := `SELECT id FROM client_source WHERE client_id = $1`
		err = tx.QueryRow(ctx, checkQuery, id).Scan(&linkID)

		if err == nil {
			// Связь есть - обновляем
			updateLinkQuery := `UPDATE client_source SET source_id = $1 WHERE client_id = $2`
			_, err = tx.Exec(ctx, updateLinkQuery, *req.SourceID, id)
			if err != nil {
				return nil, fmt.Errorf("failed to update client source: %w", err)
			}
		} else {
			// Связи нет - создаем новую
			insertLinkQuery := `
				INSERT INTO client_source (id, client_id, source_id, created_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
			`
			newLinkID := uuid.New().String()
			_, err = tx.Exec(ctx, insertLinkQuery, newLinkID, id, *req.SourceID)
			if err != nil {
				return nil, fmt.Errorf("failed to create client source: %w", err)
			}
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetClientByID(ctx, id)
}

// ActivateClient активирует клиента
func (r *Repository) ActivateClient(ctx context.Context, id string) (*ClientDetail, error) {
	// Проверяем существование
	exists, err := r.ClientExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check client existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	query := `
		UPDATE clients 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`

	_, err = r.pool.Exec(ctx, query, ClientStatusActive, id)
	if err != nil {
		return nil, fmt.Errorf("failed to activate client: %w", err)
	}

	return r.GetClientByID(ctx, id)
}

// DeactivateClient деактивирует клиента
func (r *Repository) DeactivateClient(ctx context.Context, id string) (*ClientDetail, error) {
	// Проверяем существование
	exists, err := r.ClientExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check client existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	query := `
		UPDATE clients 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`

	_, err = r.pool.Exec(ctx, query, ClientStatusInactive, id)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate client: %w", err)
	}

	return r.GetClientByID(ctx, id)
}

package dm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/wm"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

// ---------
// DEAL TYPES
// ---------

// ReorderDealStatuses изменяет порядок статусов сделок
func (r *Repository) ReorderDealStatuses(ctx context.Context, statusIDs []string) error {
	if len(statusIDs) == 0 {
		return nil
	}

	// 1. Проверяем существование всех статусов одним запросом
	placeholders := make([]string, len(statusIDs))
	args := make([]interface{}, len(statusIDs))
	for i, id := range statusIDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) = $%d as all_exist
		FROM deal_statuses
		WHERE id IN (%s)
	`, len(statusIDs)+1, strings.Join(placeholders, ", "))

	args = append(args, len(statusIDs))

	var allExist bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&allExist)
	if err != nil {
		return fmt.Errorf("failed to check deal statuses existence: %w", err)
	}

	if !allExist {
		return fmt.Errorf("one or more deal statuses not found")
	}

	// 2. Массовое обновление через CASE с параметризацией
	caseWhen := make([]string, len(statusIDs))
	whenArgs := make([]interface{}, len(statusIDs)*2)

	for i, id := range statusIDs {
		caseWhen[i] = fmt.Sprintf("WHEN $%d THEN $%d", i*2+1, i*2+2)
		whenArgs[i*2] = id
		whenArgs[i*2+1] = i + 1
	}

	updateQuery := fmt.Sprintf(`
		UPDATE deal_statuses 
		SET sort_order = CASE id %s END,
			updated_at = CURRENT_TIMESTAMP 
		WHERE id IN (%s)`,
		strings.Join(caseWhen, " "),
		placeholders[0], // используем первый плейсхолдер для IN
	)

	// Добавляем аргументы для IN
	inArgs := make([]interface{}, len(statusIDs))
	for i, id := range statusIDs {
		inArgs[i] = id
	}

	// Объединяем все аргументы
	allArgs := append(whenArgs, inArgs...)

	_, err = r.pool.Exec(ctx, updateQuery, allArgs...)
	if err != nil {
		return fmt.Errorf("failed to update deal statuses order: %w", err)
	}

	return nil
}

func (r *Repository) GetDealTypeByID(ctx context.Context, id string) (*DealType, error) {
	query := `
		SELECT id, name, comment, created_at, updated_at
		FROM deal_types
		WHERE id = $1
	`

	var dealType DealType
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&dealType.ID,
		&dealType.Name,
		&dealType.Comment,
		&dealType.CreatedAt,
		&dealType.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get deal type: %w", err)
	}

	return &dealType, nil
}

func (r *Repository) GetDealTypes(ctx context.Context, page, limit int, search string) ([]DealType, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	if search != "" {
		searchConditions := []string{
			"name ILIKE $" + strconv.Itoa(argIndex),
			"comment ILIKE $" + strconv.Itoa(argIndex),
		}
		conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(*) FROM deal_types " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deal types: %w", err)
	}

	// Получаем типы с пагинацией
	query := `
		SELECT id, name, comment, created_at, updated_at
		FROM deal_types
	` + whereClause + `
		ORDER BY name ASC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query deal types: %w", err)
	}
	defer rows.Close()

	var dealTypes []DealType
	for rows.Next() {
		var dealType DealType
		err := rows.Scan(
			&dealType.ID,
			&dealType.Name,
			&dealType.Comment,
			&dealType.CreatedAt,
			&dealType.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan deal type: %w", err)
		}
		dealTypes = append(dealTypes, dealType)
	}

	return dealTypes, total, nil
}

func (r *Repository) CreateDealType(ctx context.Context, req CreateDealTypeRequest) (*DealType, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	id := uuid.New().String()

	query := `
		INSERT INTO deal_types (id, name, comment, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, comment, created_at, updated_at
	`

	var dealType DealType
	err := r.pool.QueryRow(ctx, query,
		id,
		name,
		core.NullIfEmptyPtr(req.Comment),
	).Scan(
		&dealType.ID,
		&dealType.Name,
		&dealType.Comment,
		&dealType.CreatedAt,
		&dealType.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create deal type: %w", err)
	}

	return &dealType, nil
}

func (r *Repository) UpdateDealType(ctx context.Context, id string, req UpdateDealTypeRequest) (*DealType, error) {
	// Проверяем существование
	_, err := r.GetDealTypeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("deal type not found: %w", err)
	}

	updater := core.NewUpdater("deal_types")

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name != "" {
			updater.SetString("name", name)
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

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetDealTypeByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update deal type: %w", err)
	}

	return r.GetDealTypeByID(ctx, id)
}

// DeleteDealType удаляет тип сделки
func (r *Repository) DeleteDealType(ctx context.Context, id string) error {
	// Проверяем существование
	_, err := r.GetDealTypeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("deal type not found: %w", err)
	}

	// Проверяем, используется ли тип в сделках
	checkQuery := `SELECT EXISTS(SELECT 1 FROM deals WHERE type_id = $1)`
	var isUsed bool
	err = r.pool.QueryRow(ctx, checkQuery, id).Scan(&isUsed)
	if err != nil {
		return fmt.Errorf("failed to check if deal type is used: %w", err)
	}

	if isUsed {
		return fmt.Errorf("cannot delete deal type that is used in deals")
	}

	// Удаляем тип
	query := `DELETE FROM deal_types WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deal type: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("deal type not found")
	}

	return nil
}

// GetDealTypesByIDs возвращает типы сделок по их ID
func (r *Repository) GetDealTypesByIDs(ctx context.Context, ids []string) ([]DealType, error) {
	if len(ids) == 0 {
		return []DealType{}, nil
	}

	// Создаем плейсхолдеры для IN запроса
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, name, comment, created_at, updated_at
		FROM deal_types
		WHERE id IN (%s)
		ORDER BY name ASC
	`, strings.Join(placeholders, ", "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deal types by ids: %w", err)
	}
	defer rows.Close()

	var dealTypes []DealType
	for rows.Next() {
		var dealType DealType
		err := rows.Scan(
			&dealType.ID,
			&dealType.Name,
			&dealType.Comment,
			&dealType.CreatedAt,
			&dealType.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deal type: %w", err)
		}
		dealTypes = append(dealTypes, dealType)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deal types: %w", err)
	}

	return dealTypes, nil
}

// ---------
// DEAL STATUSES
// ---------

// GetDealStatusByID возвращает статус по ID
func (r *Repository) GetDealStatusByID(ctx context.Context, id string) (*DealStatus, error) {
	query := `
        SELECT id, name, comment, sort_order, color, is_default, created_at, updated_at
        FROM deal_statuses
        WHERE id = $1
    `

	var status DealStatus
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&status.ID,
		&status.Name,
		&status.Comment,
		&status.SortOrder,
		&status.Color,
		&status.IsDefault,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get deal status: %w", err)
	}

	return &status, nil
}

// GetDealStatuses возвращает список статусов с пагинацией
func (r *Repository) GetDealStatuses(ctx context.Context, page, limit int, search string) ([]DealStatus, int, error) {
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	if search != "" {
		searchConditions := []string{
			"name ILIKE $" + strconv.Itoa(argIndex),
			"comment ILIKE $" + strconv.Itoa(argIndex),
		}
		conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(*) FROM deal_statuses " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deal statuses: %w", err)
	}

	// Получаем статусы с пагинацией
	query := `
        SELECT id, name, comment, sort_order, color, is_default, created_at, updated_at
        FROM deal_statuses
    ` + whereClause + `
        ORDER BY sort_order ASC, name ASC
        LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query deal statuses: %w", err)
	}
	defer rows.Close()

	var statuses []DealStatus
	for rows.Next() {
		var status DealStatus
		err := rows.Scan(
			&status.ID,
			&status.Name,
			&status.Comment,
			&status.SortOrder,
			&status.Color,
			&status.IsDefault,
			&status.CreatedAt,
			&status.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan deal status: %w", err)
		}
		statuses = append(statuses, status)
	}

	return statuses, total, nil
}

// CreateDealStatus создает новый статус
func (r *Repository) CreateDealStatus(ctx context.Context, req CreateDealStatusRequest) (*DealStatus, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	id := uuid.New().String()

	// Определяем значение is_default
	isDefault := false
	if req.IsDefault != nil && *req.IsDefault {
		// Если пытаются создать дефолтный статус, нужно сбросить дефолтность у текущего
		// Начинаем транзакцию
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		// Сбрасываем is_default у всех статусов
		_, err = tx.Exec(ctx, `UPDATE deal_statuses SET is_default = false, updated_at = CURRENT_TIMESTAMP`)
		if err != nil {
			return nil, fmt.Errorf("failed to reset default statuses: %w", err)
		}

		isDefault = true

		// Создаем статус
		query := `
            INSERT INTO deal_statuses (id, name, comment, sort_order, color, is_default, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
            RETURNING id, name, comment, sort_order, color, is_default, created_at, updated_at
        `

		var status DealStatus
		err = tx.QueryRow(ctx, query,
			id,
			name,
			core.NullIfEmptyPtr(req.Comment),
			req.SortOrder,
			core.NullIfEmptyPtr(req.Color),
			isDefault,
		).Scan(
			&status.ID,
			&status.Name,
			&status.Comment,
			&status.SortOrder,
			&status.Color,
			&status.IsDefault,
			&status.CreatedAt,
			&status.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create deal status: %w", err)
		}

		// Коммитим транзакцию
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return &status, nil
	} else {
		// Обычное создание (не дефолтный статус)
		query := `
            INSERT INTO deal_statuses (id, name, comment, sort_order, color, is_default, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
            RETURNING id, name, comment, sort_order, color, is_default, created_at, updated_at
        `

		var status DealStatus
		err := r.pool.QueryRow(ctx, query,
			id,
			name,
			core.NullIfEmptyPtr(req.Comment),
			req.SortOrder,
			core.NullIfEmptyPtr(req.Color),
			isDefault,
		).Scan(
			&status.ID,
			&status.Name,
			&status.Comment,
			&status.SortOrder,
			&status.Color,
			&status.IsDefault,
			&status.CreatedAt,
			&status.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create deal status: %w", err)
		}

		return &status, nil
	}
}

// UpdateDealStatus обновляет статус
func (r *Repository) UpdateDealStatus(ctx context.Context, id string, req UpdateDealStatusRequest) (*DealStatus, error) {
	// Проверяем существование статуса
	existing, err := r.GetDealStatusByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("deal status not found: %w", err)
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Определяем, будет ли статус дефолтным после обновления
	newIsDefault := existing.IsDefault
	if req.IsDefault != nil {
		newIsDefault = *req.IsDefault
	}

	// Если снимаем дефолтность
	if existing.IsDefault && !newIsDefault {
		// Проверяем, есть ли другие статусы, которые станут/останутся дефолтными
		var otherDefaultCount int
		err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM deal_statuses WHERE id != $1 AND is_default = true`, id).Scan(&otherDefaultCount)
		if err != nil {
			return nil, fmt.Errorf("failed to check other default statuses: %w", err)
		}

		// Если нет других дефолтных статусов, и мы пытаемся снять флаг с единственного - ошибка
		if otherDefaultCount == 0 {
			// Проверяем, есть ли вообще другие статусы (чтобы не остаться без статусов)
			var otherCount int
			err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM deal_statuses WHERE id != $1`, id).Scan(&otherCount)
			if err != nil {
				return nil, fmt.Errorf("failed to check other statuses: %w", err)
			}

			if otherCount == 0 {
				return nil, fmt.Errorf("cannot remove default flag from the only status")
			}

			// Есть другие статусы, но ни один не дефолтный - значит нужно сделать первый попавшийся дефолтным
			var newDefaultID string
			err = tx.QueryRow(ctx, `
				SELECT id FROM deal_statuses 
				WHERE id != $1 
				ORDER BY sort_order ASC, created_at ASC 
				LIMIT 1
			`, id).Scan(&newDefaultID)
			if err != nil {
				return nil, fmt.Errorf("failed to find status to set as new default: %w", err)
			}

			// Назначаем новый дефолтный статус
			_, err = tx.Exec(ctx, `UPDATE deal_statuses SET is_default = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, newDefaultID)
			if err != nil {
				return nil, fmt.Errorf("failed to set new default status: %w", err)
			}
		}
	}

	// Если делаем этот статус дефолтным (и он еще не дефолтный)
	if !existing.IsDefault && newIsDefault {
		// Сбрасываем is_default у всех статусов
		_, err = tx.Exec(ctx, `UPDATE deal_statuses SET is_default = false, updated_at = CURRENT_TIMESTAMP`)
		if err != nil {
			return nil, fmt.Errorf("failed to reset default statuses: %w", err)
		}
		// Этот статус станет дефолтным в апдейте ниже
	}

	// Строим апдейт
	updater := core.NewUpdater("deal_statuses")

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name != "" {
			updater.SetString("name", name)
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

	if req.SortOrder != nil {
		updater.SetInt("sort_order", *req.SortOrder)
	}

	if req.Color != nil {
		if *req.Color == "" {
			updater.SetNull("color")
		} else {
			color := strings.TrimSpace(*req.Color)
			updater.SetString("color", color)
		}
	}

	if req.IsDefault != nil {
		updater.SetBool("is_default", *req.IsDefault)
	}

	query, args := updater.Where("id = $1", id).Build()
	if query != "" {
		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update deal status: %w", err)
		}
	}

	// Финальная проверка: убеждаемся что есть хотя бы один дефолтный статус
	var defaultCount int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM deal_statuses WHERE is_default = true`).Scan(&defaultCount)
	if err != nil {
		return nil, fmt.Errorf("failed to verify default status existence: %w", err)
	}

	if defaultCount == 0 {
		// Если вдруг не осталось дефолтных статусов (баг) - назначаем первый
		var fallbackID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM deal_statuses 
			ORDER BY sort_order ASC, created_at ASC 
			LIMIT 1
		`).Scan(&fallbackID)
		if err != nil {
			return nil, fmt.Errorf("critical: no statuses found to set as default")
		}

		_, err = tx.Exec(ctx, `UPDATE deal_statuses SET is_default = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, fallbackID)
		if err != nil {
			return nil, fmt.Errorf("failed to set fallback default status: %w", err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetDealStatusByID(ctx, id)
}

// DeleteDealStatus удаляет статус сделки
func (r *Repository) DeleteDealStatus(ctx context.Context, id string) error {
	// Проверяем существование
	status, err := r.GetDealStatusByID(ctx, id)
	if err != nil {
		return fmt.Errorf("deal status not found: %w", err)
	}

	// Проверяем, используется ли статус в сделках
	checkQuery := `SELECT EXISTS(SELECT 1 FROM deal_status WHERE status_id = $1)`
	var isUsed bool
	err = r.pool.QueryRow(ctx, checkQuery, id).Scan(&isUsed)
	if err != nil {
		return fmt.Errorf("failed to check if deal status is used: %w", err)
	}

	if isUsed {
		return fmt.Errorf("cannot delete deal status that is used in deals")
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Проверяем, не является ли этот статус единственным дефолтным
	if status.IsDefault {
		// Проверяем, есть ли другие статусы
		var otherCount int
		err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM deal_statuses WHERE id != $1`, id).Scan(&otherCount)
		if err != nil {
			return fmt.Errorf("failed to check other statuses: %w", err)
		}

		if otherCount == 0 {
			return fmt.Errorf("cannot delete the only deal status")
		}

		// Если есть другие статусы, но все они не дефолтные - тоже нельзя удалять
		var otherDefaultCount int
		err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM deal_statuses WHERE id != $1 AND is_default = true`, id).Scan(&otherDefaultCount)
		if err != nil {
			return fmt.Errorf("failed to check other default statuses: %w", err)
		}

		if otherDefaultCount == 0 {
			return fmt.Errorf("cannot delete the only default status. Set another status as default first")
		}
	}

	// Удаляем статус
	query := `DELETE FROM deal_statuses WHERE id = $1`
	result, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deal status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("deal status not found")
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDealStatusesByIDs возвращает статусы сделок по их ID
func (r *Repository) GetDealStatusesByIDs(ctx context.Context, ids []string) ([]DealStatus, error) {
	if len(ids) == 0 {
		return []DealStatus{}, nil
	}

	// Создаем плейсхолдеры для IN запроса
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
        SELECT id, name, comment, sort_order, color, is_default, created_at, updated_at
        FROM deal_statuses
        WHERE id IN (%s)
        ORDER BY sort_order ASC, name ASC
    `, strings.Join(placeholders, ", "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deal statuses by ids: %w", err)
	}
	defer rows.Close()

	var statuses []DealStatus
	for rows.Next() {
		var status DealStatus
		err := rows.Scan(
			&status.ID,
			&status.Name,
			&status.Comment,
			&status.SortOrder,
			&status.Color,
			&status.IsDefault,
			&status.CreatedAt,
			&status.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deal status: %w", err)
		}
		statuses = append(statuses, status)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deal statuses: %w", err)
	}

	return statuses, nil
}

// ---------
// DEALS
// ---------

// CreateDeal создает новую сделку
func (r *Repository) CreateDeal(ctx context.Context, req CreateDealRequest) (*DealWithPositions, error) {
	id := uuid.New().String()

	// Получаем дефолтный статус
	var defaultStatusID *string
	defaultQuery := `SELECT id FROM deal_statuses WHERE is_default = true LIMIT 1`
	err := r.pool.QueryRow(ctx, defaultQuery).Scan(&defaultStatusID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default deal status: %w", err)
	}

	if defaultStatusID == nil {
		return nil, fmt.Errorf("no default deal status found")
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Создаем сделку
	dealQuery := `
        INSERT INTO deals (id, comment, type_id, created_at, updated_at)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id, comment, type_id, created_at, updated_at
    `

	var deal Deal
	err = tx.QueryRow(ctx, dealQuery,
		id,
		req.Comment,
		req.TypeID,
	).Scan(
		&deal.ID,
		&deal.Comment,
		&deal.TypeID,
		&deal.CreatedAt,
		&deal.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create deal: %w", err)
	}

	// 2. Присваиваем дефолтный статус
	statusQuery := `
        INSERT INTO deal_status (deal_id, status_id, created_at)
        VALUES ($1, $2, CURRENT_TIMESTAMP)
    `
	_, err = tx.Exec(ctx, statusQuery, deal.ID, *defaultStatusID)
	if err != nil {
		return nil, fmt.Errorf("failed to assign default status to deal: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Получаем полную информацию о сделке
	detailedDeal, err := r.GetDealWithDetails(ctx, deal.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load deal details: %w", err)
	}

	return detailedDeal, nil
}

// UpdateDeal обновляет сделку и все связанные сущности
func (r *Repository) UpdateDeal(ctx context.Context, id string, req UpdateDealRequest) error {
	// Проверяем существование сделки
	exists, err := r.DealExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check deal existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("deal not found")
	}

	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Обновляем основные поля сделки
	if err := r.updateDealBase(ctx, tx, id, req); err != nil {
		return err
	}

	// 2. Обновляем клиента если указан
	if err := r.updateDealClient(ctx, tx, id, req.ClientID); err != nil {
		return err
	}

	// 3. Обновляем статус если указан
	if req.StatusID != nil {
		if err := r.updateDealStatus(ctx, tx, id, *req.StatusID); err != nil {
			return err
		}
	}

	// 4. Обновляем сотрудников если указаны (полная замена)
	if req.Employees != nil {
		if err := r.updateDealEmployees(ctx, tx, id, req.Employees); err != nil {
			return err
		}
	}

	// 5. Обновляем позиции если указаны
	if req.Positions != nil {
		if err := r.updateDealPositions(ctx, tx, id, req.Positions); err != nil {
			return err
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteDeal удаляет сделку
func (r *Repository) DeleteDeal(ctx context.Context, id string) error {
	// Проверяем существование
	exists, err := r.DealExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check deal existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("deal not found")
	}

	query := `DELETE FROM deals WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deal: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("deal not found")
	}

	return nil
}

// DealExists проверяет существование сделки
func (r *Repository) DealExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM deals WHERE id = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check deal existence: %w", err)
	}
	return exists, nil
}

// ---------
// DEALS - GET
// ---------

// GetDealWithDetails возвращает сделку со всеми связанными сущностями
func (r *Repository) GetDealWithDetails(ctx context.Context, id string) (*DealWithPositions, error) {
	// 1. Получаем базовую информацию о сделке с типом и статусом через JOIN
	dealQuery := `
		SELECT 
			d.id, d.comment, d.type_id, d.created_at, d.updated_at,
			dt.id, dt.name, dt.comment, dt.created_at, dt.updated_at,
			ds.id, ds.name, ds.comment, ds.sort_order, ds.color, ds.created_at, ds.updated_at
		FROM deals d
		LEFT JOIN deal_types dt ON d.type_id = dt.id
		LEFT JOIN deal_status ds_active ON d.id = ds_active.deal_id
		LEFT JOIN deal_statuses ds ON ds_active.status_id = ds.id
		WHERE d.id = $1
	`

	var deal DealWithPositions
	// Инициализируем все слайсы как пустые, а не nil
	deal.Employees = []hrm.EmployeeListItem{}
	deal.Positions = []DealPosition{}

	var typeID *string
	var typeName *string
	var typeComment *string
	var typeCreatedAt *time.Time
	var typeUpdatedAt *time.Time

	var statusID *string
	var statusName *string
	var statusComment *string
	var statusSortOrder *int
	var statusColor *string
	var statusCreatedAt *time.Time
	var statusUpdatedAt *time.Time

	err := r.pool.QueryRow(ctx, dealQuery, id).Scan(
		&deal.ID,
		&deal.Comment,
		&deal.TypeID,
		&deal.CreatedAt,
		&deal.UpdatedAt,
		&typeID,
		&typeName,
		&typeComment,
		&typeCreatedAt,
		&typeUpdatedAt,
		&statusID,
		&statusName,
		&statusComment,
		&statusSortOrder,
		&statusColor,
		&statusCreatedAt,
		&statusUpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Заполняем тип если есть
	if typeID != nil {
		deal.Type = &DealType{
			ID:        *typeID,
			Name:      *typeName,
			Comment:   typeComment,
			CreatedAt: *typeCreatedAt,
			UpdatedAt: *typeUpdatedAt,
		}
	}

	// Заполняем статус если есть
	if statusID != nil {
		deal.Status = &DealStatus{
			ID:        *statusID,
			Name:      *statusName,
			Comment:   statusComment,
			SortOrder: *statusSortOrder,
			Color:     statusColor,
			CreatedAt: *statusCreatedAt,
			UpdatedAt: *statusUpdatedAt,
		}
	}

	// 2. Получаем ID клиента
	clientQuery := `SELECT client_id FROM deal_client WHERE deal_id = $1`
	var clientID *string
	err = r.pool.QueryRow(ctx, clientQuery, id).Scan(&clientID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get client link: %w", err)
	}

	if err == pgx.ErrNoRows || clientID == nil {
		deal.Client = nil
		deal.ClientID = nil
	} else {
		deal.ClientID = clientID
	}

	// 3. Получаем ID сотрудников
	empQuery := `SELECT employee_id FROM deal_employees WHERE deal_id = $1 ORDER BY created_at ASC`
	empRows, err := r.pool.Query(ctx, empQuery, id)
	if err == nil {
		defer empRows.Close()
		var employeeIDs []string
		for empRows.Next() {
			var empID string
			if err := empRows.Scan(&empID); err == nil {
				employeeIDs = append(employeeIDs, empID)
			}
		}

		// 4. Получаем позиции сделки
		positionsQuery := `
			SELECT 
				id, name, comment, price, quantity, unit,
				unit_id, position_id, created_at, updated_at
			FROM deal_positions
			WHERE deal_id = $1
			ORDER BY created_at ASC
		`

		posRows, err := r.pool.Query(ctx, positionsQuery, id)
		if err != nil {
			return nil, fmt.Errorf("failed to query deal positions: %w", err)
		}
		defer posRows.Close()

		var unitIDs []string
		var stockPositionIDs []string
		var positions []DealPosition

		for posRows.Next() {
			var pos DealPosition
			err := posRows.Scan(
				&pos.ID,
				&pos.Name,
				&pos.Comment,
				&pos.Price,
				&pos.Quantity,
				&pos.Unit,
				&pos.UnitID,
				&pos.PositionID,
				&pos.CreatedAt,
				&pos.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan deal position: %w", err)
			}
			positions = append(positions, pos)

			if pos.UnitID != nil {
				unitIDs = append(unitIDs, *pos.UnitID)
			}
			if pos.PositionID != nil {
				stockPositionIDs = append(stockPositionIDs, *pos.PositionID)
			}
		}
		deal.Positions = positions

		// 5. Массово получаем все связанные сущности через методы других репозиториев
		var wg errgroup.Group
		var clients []crm.ClientDetail
		var employees []hrm.EmployeeListItem
		var units []wm.CatalogUnit
		var stockPositions []wm.StockPosition

		// Получаем клиента
		if clientID != nil {
			wg.Go(func() error {
				var err error
				clients, err = r.crmRepository.GetClientsByIDs(ctx, []string{*clientID})
				return err
			})
		}

		// Получаем сотрудников
		if len(employeeIDs) > 0 {
			wg.Go(func() error {
				var err error
				employees, err = r.hrmRepository.GetEmployeesByIDs(ctx, employeeIDs)
				return err
			})
		}

		// Получаем единицы каталога
		if len(unitIDs) > 0 {
			wg.Go(func() error {
				var err error
				units, err = r.wmRepository.GetCatalogUnitsByIDs(ctx, unitIDs)
				return err
			})
		}

		// Получаем складские позиции
		if len(stockPositionIDs) > 0 {
			wg.Go(func() error {
				var err error
				stockPositions, err = r.wmRepository.GetStockPositionsByIDs(ctx, stockPositionIDs)
				return err
			})
		}

		if err := wg.Wait(); err != nil {
			return nil, fmt.Errorf("failed to load related entities: %w", err)
		}

		// Инжектим клиента
		if len(clients) > 0 {
			deal.Client = &clients[0]
		}

		// Инжектим сотрудников
		deal.Employees = employees

		// Создаем map для быстрого доступа к юнитам
		unitMap := make(map[string]wm.CatalogUnit)
		for _, u := range units {
			unitMap[u.ID] = u
		}

		// Создаем map для быстрого доступа к складским позициям
		stockPosMap := make(map[string]wm.StockPosition)
		for _, sp := range stockPositions {
			stockPosMap[sp.ID] = sp
		}

		// Инжектим в позиции
		for i := range deal.Positions {
			if deal.Positions[i].UnitID != nil {
				if unit, ok := unitMap[*deal.Positions[i].UnitID]; ok {
					deal.Positions[i].CatalogUnit = &unit
				}
			}
			if deal.Positions[i].PositionID != nil {
				if sp, ok := stockPosMap[*deal.Positions[i].PositionID]; ok {
					deal.Positions[i].CatalogPosition = &sp
				}
			}
		}
	}

	return &deal, nil
}

// GetDealsWithDetails возвращает список сделок с полной информацией
func (r *Repository) GetDealsWithDetails(ctx context.Context, req GetDealsParams) (interface{}, int, error) {
	// Если есть группировка по статусам
	if req.GroupBy != nil && *req.GroupBy == "status" {
		return r.getDealsGroupedByStatus(ctx, req)
	}

	// Иначе обычная пагинация
	return r.getDealsPaginated(ctx, req)
}

// getDealsGroupedByStatus возвращает сделки, сгруппированные по статусам
func (r *Repository) getDealsGroupedByStatus(ctx context.Context, req GetDealsParams) ([]DealGroup, int, error) {
	// 1. Получаем все статусы, отсортированные по sort_order
	statusesQuery := `
		SELECT id, name, color, sort_order
		FROM deal_statuses
		ORDER BY sort_order ASC
	`
	statusRows, err := r.pool.Query(ctx, statusesQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query statuses: %w", err)
	}
	defer statusRows.Close()

	type statusInfo struct {
		ID        string
		Name      string
		Color     *string
		SortOrder int
	}

	var statuses []statusInfo
	for statusRows.Next() {
		var s statusInfo
		if err := statusRows.Scan(&s.ID, &s.Name, &s.Color, &s.SortOrder); err != nil {
			return nil, 0, fmt.Errorf("failed to scan status: %w", err)
		}
		statuses = append(statuses, s)
	}

	// 2. Для каждого статуса получаем сделки
	var groups []DealGroup
	var allDealIDs []string
	dealGroupsMap := make(map[string][]string) // statusID -> []dealID

	for _, s := range statuses {
		// Получаем ID сделок для этого статуса
		dealIDsQuery := `
			SELECT d.id
			FROM deals d
			INNER JOIN deal_status ds ON d.id = ds.deal_id
			WHERE ds.status_id = $1
			ORDER BY d.created_at DESC
		`
		rows, err := r.pool.Query(ctx, dealIDsQuery, s.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query deal IDs for status %s: %w", s.ID, err)
		}
		defer rows.Close()

		var dealIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, 0, fmt.Errorf("failed to scan deal ID: %w", err)
			}
			dealIDs = append(dealIDs, id)
			allDealIDs = append(allDealIDs, id)
		}
		dealGroupsMap[s.ID] = dealIDs
	}

	// 3. Если нет сделок - возвращаем пустые группы
	if len(allDealIDs) == 0 {
		for _, s := range statuses {
			groups = append(groups, DealGroup{
				StatusID:    s.ID,
				StatusName:  s.Name,
				StatusColor: s.Color,
				SortOrder:   s.SortOrder,
				Deals:       []Deal{},
				Count:       0,
			})
		}
		return groups, 0, nil
	}

	// 4. Получаем детали всех сделок БЕЗ ПАГИНАЦИИ
	dealsReq := GetDealsParams{
		// Устанавливаем лимит больше чем может быть сделок
		Page:  1,
		Limit: 10000, // достаточно большой лимит
	}

	// Копируем фильтры из исходного запроса
	dealsReq.TypeID = req.TypeID
	dealsReq.StatusID = req.StatusID
	dealsReq.ClientID = req.ClientID
	dealsReq.EmployeeID = req.EmployeeID
	dealsReq.Search = req.Search

	// Получаем все сделки
	allDeals, total, err := r.getDealsPaginated(ctx, dealsReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all deals: %w", err)
	}

	// Создаем map для быстрого доступа к сделкам
	dealsMap := make(map[string]Deal)
	for _, deal := range allDeals {
		dealsMap[deal.ID] = deal
	}

	// 5. Формируем группы
	for _, s := range statuses {
		var groupDeals []Deal
		for _, dealID := range dealGroupsMap[s.ID] {
			if deal, ok := dealsMap[dealID]; ok {
				groupDeals = append(groupDeals, deal)
			}
		}
		groups = append(groups, DealGroup{
			StatusID:    s.ID,
			StatusName:  s.Name,
			StatusColor: s.Color,
			SortOrder:   s.SortOrder,
			Deals:       groupDeals,
			Count:       len(groupDeals),
		})
	}

	return groups, int(total), nil
}

// GetDealsWithDetails возвращает список сделок с полной информацией
func (r *Repository) getDealsPaginated(ctx context.Context, req GetDealsParams) ([]Deal, int, error) {
	// 1. Базовый запрос для получения сделок с пагинацией
	var args []interface{}
	var conditions []string
	argIndex := 1

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	// Добавляем фильтры
	if req.TypeID != nil && *req.TypeID != "" {
		conditions = append(conditions, "d.type_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.TypeID)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		conditions = append(conditions, "d.comment ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	// Добавляем JOIN для фильтрации по статусу
	if req.StatusID != nil && *req.StatusID != "" {
		conditions = append(conditions, "ds.status_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.StatusID)
		argIndex++
	}

	// Добавляем JOIN для фильтрации по клиенту
	if req.ClientID != nil && *req.ClientID != "" {
		conditions = append(conditions, "dc.client_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ClientID)
		argIndex++
	}

	// Добавляем JOIN для фильтрации по сотруднику
	if req.EmployeeID != nil && *req.EmployeeID != "" {
		conditions = append(conditions, "de.employee_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.EmployeeID)
		argIndex++
	}

	// Формируем базовый запрос с JOIN для фильтров
	fromClause := `FROM deals d`

	if req.StatusID != nil && *req.StatusID != "" {
		fromClause += ` INNER JOIN deal_status ds ON d.id = ds.deal_id`
	}
	if req.ClientID != nil && *req.ClientID != "" {
		fromClause += ` INNER JOIN deal_client dc ON d.id = dc.deal_id`
	}
	if req.EmployeeID != nil && *req.EmployeeID != "" {
		fromClause += ` INNER JOIN deal_employees de ON d.id = de.deal_id`
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Получаем общее количество
	countQuery := "SELECT COUNT(DISTINCT d.id) " + fromClause + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Получаем базовые сделки с пагинацией
	query := `
		SELECT DISTINCT
			d.id, d.comment, d.type_id, d.created_at, d.updated_at
	` + fromClause + whereClause + `
		ORDER BY d.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, req.Limit, offset)

	rows, err := r.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query deals: %w", err)
	}
	defer rows.Close()

	var deals []Deal
	var dealIDs []string

	for rows.Next() {
		var deal Deal
		// Инициализируем слайсы
		deal.Employees = []hrm.EmployeeListItem{}
		err := rows.Scan(
			&deal.ID,
			&deal.Comment,
			&deal.TypeID,
			&deal.CreatedAt,
			&deal.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan deal: %w", err)
		}
		deals = append(deals, deal)
		dealIDs = append(dealIDs, deal.ID)
	}

	if len(deals) == 0 {
		return deals, total, nil
	}

	// 2. Собираем все ID для массовых запросов
	var (
		allTypeIDs     []string
		allClientIDs   []string
		allStatusIDs   []string
		allEmployeeIDs []string
	)

	typeSet := make(map[string]bool)
	clientSet := make(map[string]bool)
	statusSet := make(map[string]bool)
	employeeSet := make(map[string]bool)

	// Временная структура для хранения связей
	type DealRelations struct {
		ClientID    *string
		StatusID    *string
		EmployeeIDs []string
	}

	relationsMap := make(map[string]*DealRelations)

	// Получаем все связи для каждой сделки
	for _, dealID := range dealIDs {
		relations := &DealRelations{
			EmployeeIDs: []string{},
		}

		// Получаем client_id
		var clientID *string
		err := r.pool.QueryRow(ctx, `SELECT client_id FROM deal_client WHERE deal_id = $1`, dealID).Scan(&clientID)
		if err == nil && clientID != nil {
			relations.ClientID = clientID
			if !clientSet[*clientID] {
				clientSet[*clientID] = true
				allClientIDs = append(allClientIDs, *clientID)
			}
		}

		// Получаем status_id
		var statusID *string
		err = r.pool.QueryRow(ctx, `SELECT status_id FROM deal_status WHERE deal_id = $1`, dealID).Scan(&statusID)
		if err == nil && statusID != nil {
			relations.StatusID = statusID
			if !statusSet[*statusID] {
				statusSet[*statusID] = true
				allStatusIDs = append(allStatusIDs, *statusID)
			}
		}

		// Получаем employee_ids
		empRows, err := r.pool.Query(ctx, `SELECT employee_id FROM deal_employees WHERE deal_id = $1`, dealID)
		if err == nil {
			for empRows.Next() {
				var empID string
				if err := empRows.Scan(&empID); err == nil {
					relations.EmployeeIDs = append(relations.EmployeeIDs, empID)
					if !employeeSet[empID] {
						employeeSet[empID] = true
						allEmployeeIDs = append(allEmployeeIDs, empID)
					}
				}
			}
			empRows.Close()
		}

		relationsMap[dealID] = relations
	}

	// Собираем типы
	for _, deal := range deals {
		if deal.TypeID != nil {
			if !typeSet[*deal.TypeID] {
				typeSet[*deal.TypeID] = true
				allTypeIDs = append(allTypeIDs, *deal.TypeID)
			}
		}
	}

	// 3. Массово получаем все сущности
	var wg errgroup.Group

	var types []DealType
	var clients []crm.ClientDetail
	var statuses []DealStatus
	var employees []hrm.EmployeeListItem

	if len(allTypeIDs) > 0 {
		wg.Go(func() error {
			var err error
			types, err = r.GetDealTypesByIDs(ctx, allTypeIDs)
			return err
		})
	}

	if len(allClientIDs) > 0 {
		wg.Go(func() error {
			var err error
			clients, err = r.crmRepository.GetClientsByIDs(ctx, allClientIDs)
			return err
		})
	}

	if len(allStatusIDs) > 0 {
		wg.Go(func() error {
			var err error
			statuses, err = r.GetDealStatusesByIDs(ctx, allStatusIDs)
			return err
		})
	}

	if len(allEmployeeIDs) > 0 {
		wg.Go(func() error {
			var err error
			employees, err = r.hrmRepository.GetEmployeesByIDs(ctx, allEmployeeIDs)
			return err
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, 0, fmt.Errorf("failed to load related entities: %w", err)
	}

	// 4. Создаем map для быстрого доступа
	typeMap := make(map[string]DealType)
	for _, t := range types {
		typeMap[t.ID] = t
	}

	clientMap := make(map[string]crm.ClientDetail)
	for _, c := range clients {
		clientMap[c.ID] = c
	}

	statusMap := make(map[string]DealStatus)
	for _, s := range statuses {
		statusMap[s.ID] = s
	}

	employeeMap := make(map[string]hrm.EmployeeListItem)
	for _, e := range employees {
		employeeMap[e.ID] = e
	}

	// 5. Инжектим сущности в сделки
	for i := range deals {
		deal := &deals[i]
		relations := relationsMap[deal.ID]

		// Тип
		if deal.TypeID != nil {
			if t, ok := typeMap[*deal.TypeID]; ok {
				deal.Type = &t
			}
		}

		// Клиент
		if relations != nil && relations.ClientID != nil {
			if c, ok := clientMap[*relations.ClientID]; ok {
				deal.Client = &c
			}
		}

		// Статус
		if relations != nil && relations.StatusID != nil {
			if s, ok := statusMap[*relations.StatusID]; ok {
				deal.Status = &s
			}
		}

		// Сотрудники (уже инициализированы пустым слайсом)
		if relations != nil && len(relations.EmployeeIDs) > 0 {
			var dealEmployees []hrm.EmployeeListItem
			for _, empID := range relations.EmployeeIDs {
				if emp, ok := employeeMap[empID]; ok {
					dealEmployees = append(dealEmployees, emp)
				}
			}
			deal.Employees = dealEmployees
		}
	}

	return deals, total, nil
}

// EnsureDefaultStatusExists проверяет что есть хотя бы один дефолтный статус
// (можно вызывать после операций, которые могут удалить/изменить статусы)
func (r *Repository) EnsureDefaultStatusExists(ctx context.Context) error {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deal_statuses WHERE is_default = true`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check default statuses: %w", err)
	}

	if count == 0 {
		// Если нет дефолтного статуса, назначаем первый по sort_order
		var id string
		err := r.pool.QueryRow(ctx, `
            SELECT id FROM deal_statuses 
            ORDER BY sort_order ASC, created_at ASC 
            LIMIT 1
        `).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to find status to set as default: %w", err)
		}

		_, err = r.pool.Exec(ctx, `UPDATE deal_statuses SET is_default = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)
		if err != nil {
			return fmt.Errorf("failed to set default status: %w", err)
		}
	}

	return nil
}

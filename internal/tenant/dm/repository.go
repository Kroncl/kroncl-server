package dm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/fm"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool         *pgxpool.Pool
	fmRepository *fm.Repository
}

func NewRepository(pool *pgxpool.Pool, fmRepository *fm.Repository) *Repository {
	return &Repository{pool: pool, fmRepository: fmRepository}
}

// ---------
// DEAL TYPES
// ---------

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

// ---------
// DEAL STATUSES
// ---------

func (r *Repository) GetDealStatusByID(ctx context.Context, id string) (*DealStatus, error) {
	query := `
		SELECT id, name, comment, sort_order, color, created_at, updated_at
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
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get deal status: %w", err)
	}

	return &status, nil
}

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
		SELECT id, name, comment, sort_order, color, created_at, updated_at
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

func (r *Repository) CreateDealStatus(ctx context.Context, req CreateDealStatusRequest) (*DealStatus, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	id := uuid.New().String()

	query := `
		INSERT INTO deal_statuses (id, name, comment, sort_order, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, comment, sort_order, color, created_at, updated_at
	`

	var status DealStatus
	err := r.pool.QueryRow(ctx, query,
		id,
		name,
		core.NullIfEmptyPtr(req.Comment),
		req.SortOrder,
		core.NullIfEmptyPtr(req.Color),
	).Scan(
		&status.ID,
		&status.Name,
		&status.Comment,
		&status.SortOrder,
		&status.Color,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create deal status: %w", err)
	}

	return &status, nil
}

func (r *Repository) UpdateDealStatus(ctx context.Context, id string, req UpdateDealStatusRequest) (*DealStatus, error) {
	// Проверяем существование
	_, err := r.GetDealStatusByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("deal status not found: %w", err)
	}

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

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return r.GetDealStatusByID(ctx, id)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update deal status: %w", err)
	}

	return r.GetDealStatusByID(ctx, id)
}

// DeleteDealStatus удаляет статус сделки
func (r *Repository) DeleteDealStatus(ctx context.Context, id string) error {
	// Проверяем существование
	_, err := r.GetDealStatusByID(ctx, id)
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

	// Удаляем статус
	query := `DELETE FROM deal_statuses WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deal status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("deal status not found")
	}

	return nil
}

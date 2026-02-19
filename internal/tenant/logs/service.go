package logs

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(tenantPool *pgxpool.Pool) *Service {
	return &Service{
		pool: tenantPool,
	}
}

// Log creates a log entry asynchronously
func (s *Service) Log(ctx context.Context, key, accountId string, opts ...LogOption) {
	// Запускаем в горутине, чтобы не блокировать основной поток
	go func() {
		// Создаем новый контекст с таймаутом, чтобы не зависнуть
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Базовые параметры
		log := Log{
			ID:          uuid.New().String(),
			Key:         key,
			Status:      LogStatusSuccess,
			Criticality: int(config.GetCriticality(key)),
			AccountID:   accountId,
			Metadata:    make(map[string]interface{}),
			CreatedAt:   time.Now(),
		}

		// Применяем опции
		for _, opt := range opts {
			opt(&log)
		}

		query := `
			INSERT INTO logs (
				id, key, status, criticality, account_id, request_id, 
				user_agent, ip, metadata, created_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			)
		`

		_, err := s.pool.Exec(ctx, query,
			log.ID,
			log.Key,
			log.Status,
			log.Criticality,
			log.AccountID,
			log.RequestID,
			log.UserAgent,
			log.IP,
			log.Metadata,
			log.CreatedAt,
		)

		if err != nil {
			// Только логируем ошибку, не возвращаем
			fmt.Printf("failed to create log: %v\n", err)
		}
	}()
}

// LogSync creates a log entry synchronously (для критических случаев)
func (s *Service) LogSync(ctx context.Context, key, accountId string, opts ...LogOption) error {
	log := Log{
		ID:          uuid.New().String(),
		Key:         key,
		Status:      LogStatusSuccess,
		Criticality: int(config.GetCriticality(key)),
		AccountID:   accountId,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}

	for _, opt := range opts {
		opt(&log)
	}

	query := `
		INSERT INTO logs (
			id, key, status, criticality, account_id, request_id, 
			user_agent, ip, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := s.pool.Exec(ctx, query,
		log.ID,
		log.Key,
		log.Status,
		log.Criticality,
		log.AccountID,
		log.RequestID,
		log.UserAgent,
		log.IP,
		log.Metadata,
		log.CreatedAt,
	)

	return err
}

// GetLogByID возвращает лог по ID
func (s *Service) GetLogByID(ctx context.Context, id string) (*Log, error) {
	query := `
		SELECT 
			id, key, status, criticality, account_id, request_id,
			user_agent, ip, metadata, created_at
		FROM logs
		WHERE id = $1
	`

	var log Log
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.Key,
		&log.Status,
		&log.Criticality,
		&log.AccountID,
		&log.RequestID,
		&log.UserAgent,
		&log.IP,
		&log.Metadata,
		&log.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	return &log, nil
}

// GetLogs возвращает список логов с пагинацией и фильтрацией
func (s *Service) GetLogs(ctx context.Context, req GetLogsRequest) ([]Log, int64, error) {
	// Базовый запрос
	queryBase := `FROM logs`

	// WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.AccountID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("account_id = $%d", argIndex))
		args = append(args, *req.AccountID)
		argIndex++
	}

	if req.Key != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("key = $%d", argIndex))
		args = append(args, *req.Key)
		argIndex++
	}

	if req.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.MinCriticality != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("criticality >= $%d", argIndex))
		args = append(args, *req.MinCriticality)
		argIndex++
	}

	if req.MaxCriticality != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("criticality <= $%d", argIndex))
		args = append(args, *req.MaxCriticality)
		argIndex++
	}

	if req.StartDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *req.StartDate)
		argIndex++
	}

	if req.EndDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *req.EndDate)
		argIndex++
	}

	if req.Search != nil && *req.Search != "" {
		// Поиск по metadata (например, в PostgreSQL можно использовать jsonb_path_exists)
		// Упрощённо: ищем вхождение строки в metadata::text
		whereConditions = append(whereConditions, fmt.Sprintf("metadata::text ILIKE $%d", argIndex))
		args = append(args, "%"+*req.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Общее количество
	countQuery := "SELECT COUNT(*) " + queryBase + whereClause
	var total int64
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Основной запрос с пагинацией
	limit := 20
	offset := 0
	if req.Limit > 0 {
		limit = req.Limit
	}
	if req.Page > 0 {
		offset = (req.Page - 1) * limit
	}

	query := `
		SELECT 
			id, key, status, criticality, account_id, request_id,
			user_agent, ip, metadata, created_at
	` + queryBase + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	allArgs := append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var log Log
		err := rows.Scan(
			&log.ID,
			&log.Key,
			&log.Status,
			&log.Criticality,
			&log.AccountID,
			&log.RequestID,
			&log.UserAgent,
			&log.IP,
			&log.Metadata,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

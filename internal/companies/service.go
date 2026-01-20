package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/auth"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleGuest  = "guest"
)

type Service struct {
	pool       *pgxpool.Pool
	jwtService *auth.JWTService
}

func NewService(pool *pgxpool.Pool, jwtService *auth.JWTService) *Service {
	return &Service{
		pool:       pool,
		jwtService: jwtService,
	}
}

func (s *Service) Create(ctx context.Context, ownerId string, slug string, name string, description string, avatarURL string, isPublic bool) (*Company, error) {
	// 1. Валидация
	if err := s.ValidateCompanyName(name); err != nil {
		return nil, err
	}

	// 2. Проверка slug (можно в транзакции, но проверяем до нее для раннего фейла)
	isUnique, err := s.checkSlugUnique(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("slug uniqueness check failed: %w", err)
	}
	if !isUnique {
		return nil, fmt.Errorf("company slug isn't unique")
	}

	// 3. Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	// ВАЖНО: откатываем если не закоммитили
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	currentTime := time.Now()

	// 4. Генерируем UUID для компании
	companyID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate company UUID: %w", err)
	}

	// 5. Создаем компанию
	companyQuery := `
		INSERT INTO companies (
			id, slug, name, description, avatar_url, 
			is_public, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, slug, name, description, avatar_url, 
		          is_public, created_at, updated_at
	`

	var company Company
	err = tx.QueryRow(
		ctx, companyQuery,
		companyID,
		slug,
		name,
		description,
		avatarURL,
		isPublic,
		currentTime,
		currentTime,
	).Scan(
		&company.ID,
		&company.Slug,
		&company.Name,
		&company.Description,
		&company.AvatarUrl,
		&company.IsPublic,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// 6. Получаем ID роли
	var ownerRoleID int
	err = tx.QueryRow(
		ctx,
		`SELECT id FROM roles WHERE code = $1`,
		RoleOwner,
	).Scan(&ownerRoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find owner role: %w", err)
	}

	// 7. Добавляем создателя как владельца в company_accounts
	memberQuery := `
		INSERT INTO company_accounts (
			company_id, account_id, role_id, permissions,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (company_id, account_id) DO NOTHING
	`

	_, err = tx.Exec(
		ctx, memberQuery,
		companyID,
		ownerId,
		ownerRoleID,
		`{}`,
		currentTime,
		currentTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add owner to company: %w", err)
	}

	// 8. Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Обнуляем tx, чтобы defer не откатил
	tx = nil

	return &company, nil
}

func (s *Service) checkSlugUnique(ctx context.Context, slug string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM companies WHERE slug = $1`

	err := s.pool.QueryRow(ctx, query, strings.ToLower(slug)).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

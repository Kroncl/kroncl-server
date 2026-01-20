package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/auth"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

func (s *Service) Create(slug string, name string, description string, avatar_url string, is_public bool) (*Company, error) {
	isUnique, err := s.checkSlugUnique(slug)
	if err != nil {
		return nil, fmt.Errorf("slug uniqueness check failed")
	}
	if !isUnique {
		return nil, fmt.Errorf("company slug isn't unique")
	}

	okValidationName := s.ValidateCompanyName(name)
	if okValidationName != nil {
		return nil, okValidationName
	}

	ctx := context.Background()
	uuid, err := s.generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate company UUID: %w", err)
	}

	currentTime := time.Now()

	query := `
		INSERT INTO companies (id, slug, name, description, avatar_url, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, slug, name, description, avatar_url, is_public, created_at, updated_at
	`

	var company Company
	err = s.pool.QueryRow(
		ctx, query,
		uuid,
		slug,
		name,
		description,
		avatar_url,
		is_public,
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
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &company, nil
}

func (s *Service) checkSlugUnique(slug string) (bool, error) {
	ctx := context.Background()
	var count int
	query := `SELECT COUNT(*) FROM companies WHERE slug = $1`

	err := s.pool.QueryRow(ctx, query, strings.ToLower(slug)).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (s *Service) generateUUID() (string, error) {
	ctx := context.Background()
	var uuid string
	err := s.pool.QueryRow(ctx, "SELECT gen_random_uuid()").Scan(&uuid)
	return uuid, err
}

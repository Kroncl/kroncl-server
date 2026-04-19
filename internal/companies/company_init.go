package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/pricing"
	"time"

	"github.com/google/uuid"
)

func (s *Service) Create(
	ctx context.Context,
	ownerId string,
	slug string,
	name string,
	description string,
	avatarURL string,
	isPublic bool,
	planCode string,
	region string,
) (*CreateCompanyResponse, error) {
	if planCode == "" {
		return nil, fmt.Errorf("plan_code is required")
	}

	if region == "" {
		region = RegionRu
	}
	if !IsValidRegion(region) {
		return nil, fmt.Errorf("invalid region: %s", region)
	}

	if err := s.ValidateCompanyName(name); err != nil {
		return nil, err
	}

	isUnique, err := s.checkSlugUnique(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("slug uniqueness check failed: %w", err)
	}
	if !isUnique {
		return nil, fmt.Errorf("company slug isn't unique")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	currentTime := time.Now()

	companyID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate company UUID: %w", err)
	}

	companyQuery := `
		INSERT INTO companies (
			id, slug, name, description, avatar_url, 
			is_public, region, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, slug, name, description, avatar_url, 
		          is_public, email, region, site, metadata, created_at, updated_at
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
		region,
		currentTime,
		currentTime,
	).Scan(
		&company.ID,
		&company.Slug,
		&company.Name,
		&company.Description,
		&company.AvatarUrl,
		&company.IsPublic,
		&company.Email,
		&company.Region,
		&company.Site,
		&company.Metadata,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	memberQuery := `
		INSERT INTO company_accounts (
			company_id, account_id, role_code, permissions,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (company_id, account_id) DO NOTHING
	`

	_, err = tx.Exec(
		ctx, memberQuery,
		companyID,
		ownerId,
		RoleOwner,
		`{}`,
		currentTime,
		currentTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add owner to company: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx = nil

	expiresAt := currentTime.Add(time.Duration(config.PRICING_TRIAL_PERIOD_DAYS) * 24 * time.Hour)

	stoicPlanCode := config.PRICING_PLAN_LVL_1

	trialTx, err := s.pricingService.CreateTransaction(
		ctx,
		company.ID,
		ownerId,
		nil,
		pricing.CurrencyRUB,
		&stoicPlanCode,
		&planCode,
		true,
		expiresAt,
	)
	if err != nil {
		s.deleteCompany(ctx, company.ID)
		return nil, fmt.Errorf("failed to create trial transaction: %w", err)
	}

	_, err = s.pricingService.UpdateTransactionStatus(ctx, trialTx.ID, pricing.TransactionStatusSuccess)
	if err != nil {
		s.deleteCompany(ctx, company.ID)
		return nil, fmt.Errorf("failed to update trial transaction status: %w", err)
	}

	storage, err := s.storage.InitStorage(ctx, company.ID)
	if err != nil || storage == nil {
		s.deleteCompany(ctx, company.ID)
		return nil, fmt.Errorf("error init company storage: %w", err)
	}

	companyWithStorage := CreateCompanyResponse{
		Company: company,
		Storage: storage,
	}

	return &companyWithStorage, nil
}

func (s *Service) deleteCompany(ctx context.Context, companyID string) {
	_, _ = s.pool.Exec(ctx, `DELETE FROM companies WHERE id = $1`, companyID)
}

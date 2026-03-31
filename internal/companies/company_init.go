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
) (*CreateCompanyResponse, error) {
	// 0. Проверка planCode
	if planCode == "" {
		return nil, fmt.Errorf("plan_code is required")
	}

	// 1. Валидация
	if err := s.ValidateCompanyName(name); err != nil {
		return nil, err
	}

	// 2. Проверка slug
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

	// 8. Коммитим транзакцию по созданию компании
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx = nil

	// 9. Создаем trial-транзакцию
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
		// Если транзакция не создалась — удаляем компанию
		s.deleteCompany(ctx, company.ID)
		return nil, fmt.Errorf("failed to create trial transaction: %w", err)
	}

	// 10. Обновляем статус транзакции на success
	_, err = s.pricingService.UpdateTransactionStatus(ctx, trialTx.ID, pricing.TransactionStatusSuccess)
	if err != nil {
		s.deleteCompany(ctx, company.ID)
		return nil, fmt.Errorf("failed to update trial transaction status: %w", err)
	}

	// 11. Запускаем процесс создания хранилища
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

// deleteCompany удаляет компанию при ошибке в не транзакционных операциях
func (s *Service) deleteCompany(ctx context.Context, companyID string) {
	_, _ = s.pool.Exec(ctx, `DELETE FROM companies WHERE id = $1`, companyID)
}

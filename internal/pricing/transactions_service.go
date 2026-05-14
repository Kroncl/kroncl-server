package pricing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GetLastSuccessfulTransaction возвращает последнюю успешную транзакцию для компании
func (s *Service) GetLastSuccessfulTransaction(ctx context.Context, companyID string) (*PricingTransaction, error) {
	query := `
		SELECT id, company_id, account_id, amount, currency, status, plan_code,
		       is_trial, next_plan_code, promocode_id, expires_at, created_at, updated_at
		FROM pricing_transactions
		WHERE company_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var tx PricingTransaction
	var amount *int
	var planCode *string
	var nextPlanCode *string
	var promocodeId *string

	err := s.pool.QueryRow(ctx, query, companyID, TransactionStatusSuccess).Scan(
		&tx.ID,
		&tx.CompanyID,
		&tx.AccountID,
		&amount,
		&tx.Currency,
		&tx.Status,
		&planCode,
		&tx.IsTrial,
		&nextPlanCode,
		&promocodeId,
		&tx.ExpiresAt,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get last successful transaction: %w", err)
	}

	tx.Amount = amount
	tx.PlanCode = planCode
	tx.NextPlanCode = nextPlanCode
	tx.PromocodeId = promocodeId

	return &tx, nil
}

// CreateTransaction создает новую транзакцию
func (s *Service) CreateTransaction(ctx context.Context, companyID, accountID string, amount *int, currency Currency, planCode, nextPlanCode *string, isTrial bool, expiresAt time.Time, promocodeCode *string) (*PricingTransaction, error) {
	var promocodeID *string
	var finalPlanCode *string
	var finalExpiresAt = expiresAt

	// Промокод обрабатываем только для trial транзакций
	if isTrial && promocodeCode != nil && *promocodeCode != "" {
		promocode, err := s.GetPromocodeByCode(ctx, *promocodeCode)
		if err == nil && promocode != nil {
			// промокод существует - применяем его
			promocodeID = &promocode.ID
			finalPlanCode = &promocode.PlanID
			finalExpiresAt = time.Now().AddDate(0, 0, promocode.TrialPeriodDays)
		}
		// если ошибка - игнорируем, оставляем исходные параметры
	}

	// Если промокод не применился, используем переданные параметры
	if finalPlanCode == nil {
		finalPlanCode = planCode
	}

	// Проверяем, была ли уже успешная trial-транзакция для этой компании
	if isTrial {
		var trialExists bool
		checkQuery := `
			SELECT EXISTS(
				SELECT 1 FROM pricing_transactions
				WHERE company_id = $1 AND is_trial = true AND status = $2
			)
		`
		err := s.pool.QueryRow(ctx, checkQuery, companyID, TransactionStatusSuccess).Scan(&trialExists)
		if err != nil {
			return nil, fmt.Errorf("failed to check trial transaction: %w", err)
		}
		if trialExists {
			return nil, fmt.Errorf("trial period already used for this company")
		}
	}

	id := uuid.New().String()

	query := `
		INSERT INTO pricing_transactions (
			id, company_id, account_id, amount, currency, status,
			plan_code, is_trial, next_plan_code, promocode_id, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, company_id, account_id, amount, currency, status,
		          plan_code, is_trial, next_plan_code, promocode_id, expires_at, created_at, updated_at
	`

	var tx PricingTransaction
	var returnedAmount *int
	var returnedPlanCode *string
	var returnedNextPlanCode *string
	var returnedPromocodeId *string

	now := time.Now()

	err := s.pool.QueryRow(ctx, query,
		id,
		companyID,
		accountID,
		amount,
		currency,
		TransactionStatusPending,
		finalPlanCode,
		isTrial,
		nextPlanCode,
		promocodeID,
		finalExpiresAt,
		now,
		now,
	).Scan(
		&tx.ID,
		&tx.CompanyID,
		&tx.AccountID,
		&returnedAmount,
		&tx.Currency,
		&tx.Status,
		&returnedPlanCode,
		&tx.IsTrial,
		&returnedNextPlanCode,
		&returnedPromocodeId,
		&tx.ExpiresAt,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	tx.Amount = returnedAmount
	tx.PlanCode = returnedPlanCode
	tx.NextPlanCode = returnedNextPlanCode
	tx.PromocodeId = returnedPromocodeId

	return &tx, nil
}

// UpdateTransactionStatus обновляет статус транзакции
func (s *Service) UpdateTransactionStatus(ctx context.Context, transactionID string, status TransactionStatus) (*PricingTransaction, error) {
	query := `
		UPDATE pricing_transactions
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, company_id, account_id, amount, currency, status, plan_code,
		          is_trial, next_plan_code, promocode_id, expires_at, created_at, updated_at
	`

	var tx PricingTransaction
	var amount *int
	var planCode *string
	var nextPlanCode *string
	var promocodeId *string

	err := s.pool.QueryRow(ctx, query, transactionID, status).Scan(
		&tx.ID,
		&tx.CompanyID,
		&tx.AccountID,
		&amount,
		&tx.Currency,
		&tx.Status,
		&planCode,
		&tx.IsTrial,
		&nextPlanCode,
		&promocodeId,
		&tx.ExpiresAt,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	tx.Amount = amount
	tx.PlanCode = planCode
	tx.NextPlanCode = nextPlanCode
	tx.PromocodeId = promocodeId

	return &tx, nil
}

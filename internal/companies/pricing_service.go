package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/pricing"
	"strings"
	"time"
)

// GetCompanyPlan возвращает текущий план компании
func (s *Service) GetCompanyPlan(ctx context.Context, companyID string) (*CompanyPlanResponse, error) {
	// Получаем последнюю успешную транзакцию
	tx, err := s.pricingService.GetLastSuccessfulTransaction(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last transaction: %w", err)
	}

	// Получаем текущий план по plan_code
	if tx.PlanCode == nil {
		return nil, fmt.Errorf("no plan code in transaction")
	}
	currentPlan, err := s.pricingService.GetPlanByCode(ctx, *tx.PlanCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get current plan: %w", err)
	}

	// Вычисляем оставшиеся дни
	daysLeft := int(time.Until(tx.ExpiresAt).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	resp := &CompanyPlanResponse{
		IsTrial:     tx.IsTrial,
		ExpiresAt:   tx.ExpiresAt,
		DaysLeft:    daysLeft,
		CurrentPlan: *currentPlan,
	}

	// Если есть next_plan_code, получаем следующий план
	if tx.NextPlanCode != nil && *tx.NextPlanCode != "" {
		nextPlan, err := s.pricingService.GetPlanByCode(ctx, *tx.NextPlanCode)
		if err == nil {
			resp.NextPlan = nextPlan
		}
	}

	return resp, nil
}

// GetCompanyTransactions возвращает историю транзакций компании (исключая trial)
func (s *Service) GetCompanyTransactions(ctx context.Context, companyID string, page, limit int) ([]pricing.PricingTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var args []interface{}
	argIndex := 1
	offset := (page - 1) * limit

	// Базовый where: не trial
	whereClause := "WHERE company_id = $1 AND is_trial = false"
	args = append(args, companyID)
	argIndex += 1

	// Count
	countQuery := "SELECT COUNT(*) FROM pricing_transactions " + whereClause
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Data
	query := `
		SELECT id, company_id, account_id, amount, currency, status, plan_code,
		       is_trial, next_plan_code, expires_at, created_at, updated_at
		FROM pricing_transactions
	` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	allArgs := append(args, limit, offset)
	rows, err := s.pool.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []pricing.PricingTransaction
	for rows.Next() {
		var tx pricing.PricingTransaction
		var amount *int
		var planCode *string
		var nextPlanCode *string

		err := rows.Scan(
			&tx.ID,
			&tx.CompanyID,
			&tx.AccountID,
			&amount,
			&tx.Currency,
			&tx.Status,
			&planCode,
			&tx.IsTrial,
			&nextPlanCode,
			&tx.ExpiresAt,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}

		tx.Amount = amount
		tx.PlanCode = planCode
		tx.NextPlanCode = nextPlanCode
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// CreateNewTransaction создает новую транзакцию для смены плана
func (s *Service) CreateNewTransaction(ctx context.Context, companyID, accountID string, req *MigratePlanRequest) (*pricing.PricingTransaction, error) {
	// 0. Проверяем, нет ли уже зависшей (pending) транзакции
	var pendingExists bool
	checkPendingQuery := `
		SELECT EXISTS(
			SELECT 1 FROM pricing_transactions
			WHERE company_id = $1 AND status = $2 AND is_trial = false
		)
	`
	err := s.pool.QueryRow(ctx, checkPendingQuery, companyID, pricing.TransactionStatusPending).Scan(&pendingExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check pending transaction: %w", err)
	}
	if pendingExists {
		return nil, fmt.Errorf("a pending transaction already exists for this company. Please wait for it to be processed or contact support")
	}

	// 1. Валидируем период
	var months int
	switch strings.ToLower(req.Period) {
	case "month":
		months = 1
	case "year":
		months = 12
	default:
		return nil, fmt.Errorf("invalid period: must be 'month' or 'year'")
	}

	// 2. Получаем целевой план
	targetPlan, err := s.pricingService.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return nil, fmt.Errorf("invalid plan code: %w", err)
	}

	// 3. Определяем сумму
	var amount int
	if months == 1 {
		amount = targetPlan.PricePerMonth
	} else {
		amount = targetPlan.PricePerYear
	}

	// 4. expires_at от текущего момента
	expiresAt := time.Now().AddDate(0, months, 0)

	// 5. Получаем последнюю успешную транзакцию (для логирования, если нужно)
	_, err = s.pricingService.GetLastSuccessfulTransaction(ctx, companyID)
	if err != nil {
		// Не фатально, просто логируем
		// Но можно и вернуть ошибку, если хочешь строго
	}

	// 6. Определяем next_plan_code (для будущей миграции, пока nil)
	var nextPlanCode *string = nil

	// 7. Создаем транзакцию со статусом pending
	tx, err := s.pricingService.CreateTransaction(
		ctx,
		companyID,
		accountID,
		&amount,
		pricing.CurrencyRUB,
		&req.PlanCode,
		nextPlanCode,
		false,
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

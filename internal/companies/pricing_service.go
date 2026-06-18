package companies

import (
	"context"
	"fmt"
	"kroncl-server/internal/billing"
	"kroncl-server/internal/config"
	"kroncl-server/internal/pricing"
	"strings"
	"time"
)

func (s *Service) RevokeTransaction(ctx context.Context, companyID, transactionID string) error {
	var status pricing.TransactionStatus
	var txCompanyID string
	query := `SELECT status, company_id FROM pricing_transactions WHERE id = $1`
	err := s.pool.QueryRow(ctx, query, transactionID).Scan(&status, &txCompanyID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	if txCompanyID != companyID {
		return fmt.Errorf("transaction does not belong to this company")
	}

	if status != pricing.TransactionStatusPending {
		return fmt.Errorf("transaction status is %s, cannot revoke", status)
	}

	_, err = s.pricingService.UpdateTransactionStatus(ctx, transactionID, pricing.TransactionStatusRevoked)
	if err != nil {
		return fmt.Errorf("failed to revoke transaction: %w", err)
	}

	return nil
}

func (s *Service) GetCompanyPlan(ctx context.Context, companyID string) (*CompanyPlanResponse, error) {
	tx, err := s.pricingService.GetLastSuccessfulTransaction(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last transaction: %w", err)
	}

	if tx.PlanCode == nil {
		return nil, fmt.Errorf("no plan code in transaction")
	}
	currentPlan, err := s.pricingService.GetPlanByCode(ctx, *tx.PlanCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get current plan: %w", err)
	}

	var daysTotal int
	if tx.IsTrial {
		daysTotal = config.PRICING_TRIAL_PERIOD_DAYS
	} else {
		daysTotal = int(tx.ExpiresAt.Sub(tx.CreatedAt).Hours() / 24)
		if daysTotal <= 0 {
			// fallback: если не получилось, берем разницу с текущим
			daysTotal = int(time.Until(tx.ExpiresAt).Hours() / 24)
			if daysTotal < 0 {
				daysTotal = 0
			}
		}
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
		DaysTotal:   daysTotal,
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

func (s *Service) CreateNewTransaction(ctx context.Context, companyID, accountID string, req *MigratePlanRequest) (*pricing.PricingTransaction, error) {
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

	var months int
	switch strings.ToLower(req.Period) {
	case "month":
		months = 1
	case "year":
		months = 12
	default:
		return nil, fmt.Errorf("invalid period: must be 'month' or 'year'")
	}

	targetPlan, err := s.pricingService.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return nil, fmt.Errorf("invalid plan code: %w", err)
	}

	var amount int
	if months == 1 {
		amount = targetPlan.PricePerMonth
	} else {
		amount = targetPlan.PricePerYear
	}

	expiresAt := time.Now().AddDate(0, months, 0)

	_, err = s.pricingService.GetLastSuccessfulTransaction(ctx, companyID)
	if err != nil {
		// Не фатально, просто логируем
	}

	var nextPlanCode *string = nil

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
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

func (s *Service) InitPayment(
	ctx context.Context,
	companyID, accountID string,
	req *MigratePlanRequest,
	successURL string,
) (*InitPaymentResult, error) {
	company, err := s.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	plan, err := s.pricingService.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("plan not found")
	}

	tx, err := s.CreateNewTransaction(ctx, companyID, accountID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	amountKopecks := uint64(*tx.Amount * 100) // рубли -> копейки

	// Формируем описание с датой окончания
	expiresAt := tx.ExpiresAt.Format("02.01.2006")
	description := fmt.Sprintf("Оплата тарифа «%s» для компании «%s» до %s", plan.Name, company.Name, expiresAt)

	initResp, err := s.billingService.InitPayment(ctx, &billing.InitPaymentRequest{
		OrderID:     tx.ID,
		Amount:      amountKopecks,
		Description: description,
		CustomerKey: accountID,
		WebhookURL:  s.billingService.GetWebhookURL(),
		SuccessURL:  successURL,
		FailURL:     successURL,
	})
	if err != nil {
		s.pricingService.UpdateTransactionStatus(ctx, tx.ID, pricing.TransactionStatusUnsuccess)
		return nil, fmt.Errorf("failed to init payment: %w", err)
	}

	return &InitPaymentResult{
		Transaction:    tx,
		PaymentPageURL: initResp.PaymentPageURL,
		PaymentID:      initResp.PaymentID,
	}, nil
}

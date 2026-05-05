package pricing

import (
	"context"
	"fmt"
)

type PricingMetrics struct {
	TotalTransactions   int `json:"total_transactions"`
	SuccessTransactions int `json:"success_transactions"`
	PendingTransactions int `json:"pending_transactions"`
	TrialTransactions   int `json:"trial_transactions"`
}

func (s *Service) GetPricingMetrics(ctx context.Context) (*PricingMetrics, error) {
	query := `
        SELECT 
            COUNT(*) as total_transactions,
            COUNT(CASE WHEN status = 'success' THEN 1 END) as success_transactions,
            COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_transactions,
            COUNT(CASE WHEN is_trial = true THEN 1 END) as trial_transactions
        FROM pricing_transactions
    `

	var metrics PricingMetrics
	err := s.pool.QueryRow(ctx, query).Scan(
		&metrics.TotalTransactions,
		&metrics.SuccessTransactions,
		&metrics.PendingTransactions,
		&metrics.TrialTransactions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing metrics: %w", err)
	}

	return &metrics, nil
}

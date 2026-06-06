package fm

import (
	"context"
	"fmt"
	"math"
	"time"
)

// --------------
// TIMELINE
// --------------

func (r *Repository) GetDailyBalances(ctx context.Context, startDate, endDate time.Time) ([]float64, []string, error) {
	query := `
        WITH daily AS (
            SELECT 
                DATE(created_at) as day,
                COALESCE(SUM(CASE WHEN direction = 'income' THEN base_amount ELSE 0 END), 0) -
                COALESCE(SUM(CASE WHEN direction = 'expense' THEN base_amount ELSE 0 END), 0) as balance
            FROM transactions
            WHERE created_at >= $1 AND created_at <= $2
            GROUP BY DATE(created_at)
            ORDER BY day ASC
        )
        SELECT day, balance FROM daily
    `

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query daily balances: %w", err)
	}
	defer rows.Close()

	var balances []float64
	var dates []string

	for rows.Next() {
		var day time.Time
		var balance float64
		if err := rows.Scan(&day, &balance); err != nil {
			return nil, nil, fmt.Errorf("failed to scan daily balance: %w", err)
		}
		balances = append(balances, balance)
		dates = append(dates, day.Format("2006-01-02"))
	}

	return balances, dates, nil
}

func (r *Repository) ThetaForecast(ctx context.Context, req ForecastRequest) (*ForecastResponse, error) {
	now := time.Now().Truncate(24 * time.Hour)

	var startDate, endDate time.Time

	if req.StartDate != nil {
		startDate = req.StartDate.Truncate(24 * time.Hour)
	} else {
		earliest, err := r.getEarliestTransactionDate(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get earliest transaction date: %w", err)
		}
		startDate = earliest
	}

	if req.EndDate != nil {
		endDate = req.EndDate.Truncate(24 * time.Hour)
	} else {
		endDate = now
	}

	horizon := 30
	if req.Horizon > 0 {
		horizon = req.Horizon
	}

	balances, dates, err := r.GetDailyBalances(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(balances) < 2 {
		return nil, fmt.Errorf("need at least 2 data points for forecasting, got %d", len(balances))
	}

	// Theta method: разлагаем на две линии и усредняем
	theta := 2.0 // стандартный параметр

	// Линия 1: простое экспоненциальное сглаживание (SES) с дрифтом
	alpha := 0.3
	ses := make([]float64, len(balances))
	ses[0] = balances[0]
	for i := 1; i < len(balances); i++ {
		ses[i] = alpha*balances[i] + (1-alpha)*ses[i-1]
	}

	// Определяем тренд по последним точкам SES
	var trend float64
	if len(ses) >= 2 {
		trend = (ses[len(ses)-1] - ses[0]) / float64(len(ses))
	}

	// Линия 2: линейная регрессия
	n := float64(len(balances))
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range balances {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Прогноз Theta: усредняем SES с дрифтом и линейную регрессию
	forecast := make([]float64, horizon)
	for i := 0; i < horizon; i++ {
		sesForecast := ses[len(ses)-1] + trend*float64(i+1)
		regForecast := slope*(n+float64(i)) + intercept
		forecast[i] = (sesForecast + (theta-1)*regForecast) / theta
	}

	// Собираем ответ
	points := make([]ForecastDataPoint, 0, len(balances)+horizon)

	for i, b := range balances {
		points = append(points, ForecastDataPoint{
			Date:     dates[i],
			Balance:  math.Round(b*100) / 100,
			IsActual: true,
		})
	}

	for i, f := range forecast {
		forecastDate := endDate.AddDate(0, 0, i+1)
		points = append(points, ForecastDataPoint{
			Date:     forecastDate.Format("2006-01-02"),
			Balance:  math.Round(f*100) / 100,
			IsActual: false,
		})
	}

	// Оценка уверенности
	confidence := "low"
	if len(balances) >= 30 {
		confidence = "medium"
	}
	if len(balances) >= 90 {
		confidence = "high"
	}

	return &ForecastResponse{
		Method:     ForecastMethodTheta,
		Points:     points,
		Horizon:    horizon,
		DataPoints: len(balances),
		Confidence: confidence,
	}, nil
}

func (r *Repository) getEarliestTransactionDate(ctx context.Context) (time.Time, error) {
	query := `SELECT created_at FROM transactions ORDER BY created_at ASC LIMIT 1`

	var earliest time.Time
	err := r.pool.QueryRow(ctx, query).Scan(&earliest)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return time.Now().Truncate(24 * time.Hour), nil
		}
		return time.Time{}, fmt.Errorf("failed to get earliest transaction: %w", err)
	}

	return earliest.Truncate(24 * time.Hour), nil
}

// --------------
// SUMMARY
// --------------
// ThetaForecastSummary делает прогноз и агрегирует в сводку
func (r *Repository) ThetaForecastSummary(ctx context.Context, req ForecastSummaryRequest) (*ForecastSummaryResponse, error) {
	// Переиспользуем ThetaForecast для получения точек
	timelineReq := ForecastRequest{
		Method:    req.Method,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Horizon:   req.Horizon,
	}

	timelineResp, err := r.ThetaForecast(ctx, timelineReq)
	if err != nil {
		return nil, fmt.Errorf("theta forecast failed: %w", err)
	}

	// Определяем startDate/endDate для текущей статистики
	now := time.Now().Truncate(24 * time.Hour)
	var endDate time.Time
	if req.EndDate != nil {
		endDate = req.EndDate.Truncate(24 * time.Hour)
	} else {
		endDate = now
	}

	horizon := 30
	if req.Horizon > 0 {
		horizon = req.Horizon
	}

	// Получаем текущие показатели из БД
	currentStats, err := r.getCurrentStats(ctx, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get current stats: %w", err)
	}

	// Из точек прогноза берём только прогнозные (is_actual = false)
	var forecastBalances []float64
	for _, p := range timelineResp.Points {
		if !p.IsActual {
			forecastBalances = append(forecastBalances, p.Balance)
		}
	}

	if len(forecastBalances) == 0 {
		return nil, fmt.Errorf("no forecast points generated")
	}

	// Прогнозный баланс на конец периода = последняя точка прогноза
	predictedBalance := forecastBalances[len(forecastBalances)-1]

	// Прогноз net flow = изменение баланса за горизонт
	predictedNetFlow := predictedBalance - currentStats.Balance

	// Распределяем net flow на доходы и расходы пропорционально историческим
	var incomeRatio float64
	totalFlow := currentStats.Income + currentStats.Expense
	if totalFlow > 0 {
		incomeRatio = currentStats.Income / totalFlow
	} else {
		incomeRatio = 0.5
	}

	var predictedIncome, predictedExpense float64
	if predictedNetFlow >= 0 {
		// Баланс растёт — значит доходы превышают расходы
		predictedIncome = currentStats.Income + predictedNetFlow*incomeRatio
		predictedExpense = currentStats.Expense + predictedNetFlow*(1-incomeRatio)
	} else {
		// Баланс падает — расходы превышают доходы
		netLoss := -predictedNetFlow
		predictedIncome = currentStats.Income + predictedNetFlow*(1-incomeRatio)
		predictedExpense = currentStats.Expense + netLoss*incomeRatio
	}

	// Прогноз количества операций — пропорционально горизонту
	dailyAvgTx := float64(currentStats.TxCount) / float64(timelineResp.DataPoints)
	if timelineResp.DataPoints == 0 {
		dailyAvgTx = 0
	}
	predictedTxCount := int(math.Round(dailyAvgTx * float64(horizon)))

	return &ForecastSummaryResponse{
		Method:           ForecastMethodTheta,
		Horizon:          horizon,
		DataPoints:       timelineResp.DataPoints,
		Confidence:       timelineResp.Confidence,
		PredictedBalance: math.Round(predictedBalance*100) / 100,
		PredictedIncome:  math.Round(predictedIncome*100) / 100,
		PredictedExpense: math.Round(predictedExpense*100) / 100,
		PredictedNetFlow: math.Round(predictedNetFlow*100) / 100,
		PredictedTxCount: predictedTxCount,
		CurrentBalance:   math.Round(currentStats.Balance*100) / 100,
		CurrentIncome:    math.Round(currentStats.Income*100) / 100,
		CurrentExpense:   math.Round(currentStats.Expense*100) / 100,
		CurrentTxCount:   currentStats.TxCount,
	}, nil
}

// currentStats — внутренняя структура для текущих показателей
type currentStats struct {
	Balance float64
	Income  float64
	Expense float64
	TxCount int
}

// getCurrentStats получает текущие показатели на указанную дату
func (r *Repository) getCurrentStats(ctx context.Context, asOf time.Time) (currentStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN direction = 'income' THEN base_amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN direction = 'expense' THEN base_amount ELSE 0 END), 0) as balance,
			COALESCE(SUM(CASE WHEN direction = 'income' THEN base_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN direction = 'expense' THEN base_amount ELSE 0 END), 0) as expense,
			COUNT(*) as tx_count
		FROM transactions
		WHERE created_at <= $1
	`

	var stats currentStats
	err := r.pool.QueryRow(ctx, query, asOf).Scan(
		&stats.Balance,
		&stats.Income,
		&stats.Expense,
		&stats.TxCount,
	)
	if err != nil {
		return currentStats{}, fmt.Errorf("failed to get current stats: %w", err)
	}

	return stats, nil
}

type Row interface {
	Scan(dest ...interface{}) error
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close()
}

package fm

import "time"

type ForecastMethod string

const (
	ForecastMethodTheta ForecastMethod = "theta"
)

// ----------------
// TIMELINE
// ----------------
// ForecastRequest represents incoming forecast request
type ForecastRequest struct {
	Method    ForecastMethod `json:"method" validate:"omitempty,oneof=theta"`
	StartDate *time.Time     `json:"start_date,omitempty"`                       // опционально, дефолт = 30 дней назад
	EndDate   *time.Time     `json:"end_date,omitempty"`                         // опционально, дефолт = сейчас
	Horizon   int            `json:"horizon" validate:"omitempty,min=1,max=365"` // на сколько дней вперёд, дефолт 30
}

// ForecastDataPoint represents one point in the forecast
type ForecastDataPoint struct {
	Date     string  `json:"date"`
	Balance  float64 `json:"balance"`
	IsActual bool    `json:"is_actual"` // true для исторических, false для прогнозных
}

// ForecastResponse represents the full forecast response
type ForecastResponse struct {
	Method     ForecastMethod      `json:"method"`
	Points     []ForecastDataPoint `json:"points"`
	Horizon    int                 `json:"horizon"`
	DataPoints int                 `json:"data_points"` // сколько исторических точек использовано
	Confidence string              `json:"confidence"`  // "low", "medium", "high" в зависимости от объёма данных
}

// --------------
// SUMMARY
// --------------
// ForecastSummaryRequest — запрос на сводный прогноз
type ForecastSummaryRequest struct {
	Method    ForecastMethod `json:"method" validate:"omitempty,oneof=theta"`
	StartDate *time.Time     `json:"start_date,omitempty"`
	EndDate   *time.Time     `json:"end_date,omitempty"`
	Horizon   int            `json:"horizon" validate:"omitempty,min=1,max=365"`
}

// ForecastSummaryResponse — сводка прогноза
type ForecastSummaryResponse struct {
	Method     ForecastMethod `json:"method"`
	Horizon    int            `json:"horizon"`
	DataPoints int            `json:"data_points"`
	Confidence string         `json:"confidence"`

	// Итоговые показатели за горизонт прогнозирования
	PredictedBalance float64 `json:"predicted_balance"`  // баланс на конец периода
	PredictedIncome  float64 `json:"predicted_income"`   // суммарный доход за горизонт
	PredictedExpense float64 `json:"predicted_expense"`  // суммарный расход за горизонт
	PredictedNetFlow float64 `json:"predicted_net_flow"` // доход - расход за горизонт
	PredictedTxCount int     `json:"predicted_tx_count"` // прогноз количества операций

	// Текущие фактические показатели (для сравнения)
	CurrentBalance float64 `json:"current_balance"`
	CurrentIncome  float64 `json:"current_income"`
	CurrentExpense float64 `json:"current_expense"`
	CurrentTxCount int     `json:"current_tx_count"`
}

package pricing

import (
	"time"
)

type Currency string

const (
	CurrencyRUB Currency = "RUB"
)

type PricingPlan struct {
	Code              string    `json:"code"`
	Lvl               int       `json:"lvl"`
	PricePerMonth     int       `json:"price_per_month"`
	PricePerYear      int       `json:"price_per_year"`
	PriceCurrency     Currency  `json:"price_currency"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	LimitDbMB         int       `json:"limit_db_mb"`
	LimitObjectsMB    int       `json:"limit_objects_mb"`
	LimitObjectsCount int       `json:"limit_objects_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

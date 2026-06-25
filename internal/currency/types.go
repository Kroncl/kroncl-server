package currency

import "time"

type CurrencyType string

const (
	CurrencyTypeFiat   CurrencyType = "fiat"
	CurrencyTypeCrypto CurrencyType = "crypto"
)

type RateSource string

const (
	RateSourceCBR       RateSource = "cbr"
	RateSourceCoinGecko RateSource = "coingecko"
	RateSourceManual    RateSource = "manual"
)

type CurrencyRate struct {
	Rate      float64    `json:"rate"`
	Source    RateSource `json:"source"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Currency struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Type   CurrencyType `json:"type"`
	Symbol string       `json:"symbol"`
	Rate   CurrencyRate `json:"rate"`
}

type CurrencyDatePair struct {
	CurrencyID string
	Date       time.Time
}

// RawStat — сырые данные по валюте на дату
type RawStat struct {
	Income  float64
	Expense float64
	Count   int64
}

// ConvertedSummary — результат конвертации
type ConvertedSummary struct {
	TotalIncome      float64   `json:"total_income"`
	TotalExpense     float64   `json:"total_expense"`
	NetBalance       float64   `json:"net_balance"`
	TransactionCount int64     `json:"transaction_count"`
	AvgTransaction   float64   `json:"avg_transaction"`
	Currency         *Currency `json:"currency"`
}

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

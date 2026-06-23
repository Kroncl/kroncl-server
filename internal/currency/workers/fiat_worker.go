package currencyworkers

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"kroncl-server/internal/config"
	"kroncl-server/internal/currency"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Valutes []Valute `xml:"Valute"`
}

type Valute struct {
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  int    `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

type FiatWorker struct {
	pool       *pgxpool.Pool
	httpClient *http.Client
	cron       *cron.Cron
	interval   string
	cfg        *config.CurrencyConfig
	service    *currency.Service
}

func NewFiatWorker(
	pool *pgxpool.Pool,
	interval string,
	cfg *config.CurrencyConfig,
	service *currency.Service,
) *FiatWorker {
	return &FiatWorker{
		pool: pool,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cron:     cron.New(),
		interval: interval,
		cfg:      cfg,
		service:  service,
	}
}

func (w *FiatWorker) Start() error {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := w.collectAndSave(ctx); err != nil {
			log.Printf("❌ Initial fiat rates collection failed: %v", err)
		}
		log.Printf("✅ Initial fiat rates collected")
	}()

	_, err := w.cron.AddFunc(w.interval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := w.collectAndSave(ctx); err != nil {
			log.Printf("❌ Failed to collect fiat rates: %v", err)
			return
		}

		log.Printf("✅ Fiat rates updated")
	})

	if err != nil {
		return err
	}

	w.cron.Start()
	log.Printf("✅ Fiat currency worker started with interval: %s", w.interval)
	return nil
}

func (w *FiatWorker) collectAndSave(ctx context.Context) error {
	rows, err := w.pool.Query(ctx, `SELECT id FROM currencies WHERE type = 'fiat'`)
	if err != nil {
		return fmt.Errorf("failed to get fiat currencies: %w", err)
	}
	defer rows.Close()

	ourCurrencies := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return err
		}
		ourCurrencies[code] = true
	}

	apiURL, _ := url.JoinPath(w.cfg.CbrApiUrl, "daily_utf8.xml")
	resp, err := w.httpClient.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch CBR rates: %w", err)
	}
	defer resp.Body.Close()

	var valCurs ValCurs
	if err := xml.NewDecoder(resp.Body).Decode(&valCurs); err != nil {
		return fmt.Errorf("failed to decode XML: %w", err)
	}

	for _, v := range valCurs.Valutes {
		if !ourCurrencies[v.CharCode] {
			continue
		}

		valueStr := strings.Replace(v.Value, ",", ".", 1)
		rate, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			log.Printf("⚠️ Failed to parse rate for %s: %v", v.CharCode, err)
			continue
		}

		// Номинал по умолчанию 1, для некоторых валют — 10, 100, 1000
		if v.Nominal > 1 {
			rate = rate / float64(v.Nominal)
		}

		_, err = w.pool.Exec(ctx, `
			INSERT INTO currency_rates (currency_id, rate, source, updated_at)
			VALUES ($1, $2, $3, NOW())
		`, v.CharCode, rate, currency.RateSourceCBR)
		if err != nil {
			log.Printf("⚠️ Failed to save rate for %s: %v", v.CharCode, err)
		}
	}

	w.service.InvalidateCache()

	return nil
}

func (w *FiatWorker) Stop() {
	w.cron.Stop()
}

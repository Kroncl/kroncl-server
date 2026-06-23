package currencyworkers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"kroncl-server/internal/config"
	"kroncl-server/internal/currency"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type CryptoWorker struct {
	pool       *pgxpool.Pool
	httpClient *http.Client
	cron       *cron.Cron
	interval   string
	cfg        *config.CurrencyConfig
	service    *currency.Service
}

func NewCryptoWorker(
	pool *pgxpool.Pool,
	interval string,
	cfg *config.CurrencyConfig,
	service *currency.Service,
) *CryptoWorker {
	return &CryptoWorker{
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

func (w *CryptoWorker) Start() error {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := w.collectAndSave(ctx); err != nil {
			log.Printf("❌ Initial crypto rates collection failed: %v", err)
		}
		log.Printf("✅ Initial crypto rates collected")
	}()

	_, err := w.cron.AddFunc(w.interval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := w.collectAndSave(ctx); err != nil {
			log.Printf("❌ Failed to collect crypto rates: %v", err)
			return
		}

		log.Printf("✅ Crypto rates updated")
	})

	if err != nil {
		return err
	}

	w.cron.Start()
	log.Printf("✅ Crypto currency worker started with interval: %s", w.interval)
	return nil
}

func (w *CryptoWorker) collectAndSave(ctx context.Context) error {
	// Получаем crypto валюты с full_code из таблицы currencies
	rows, err := w.pool.Query(ctx, `
		SELECT id, full_code 
		FROM currencies 
		WHERE type = 'crypto' AND full_code IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to get crypto currencies: %w", err)
	}
	defer rows.Close()

	type cryptoInfo struct {
		ID       string
		FullCode string
	}

	var cryptos []cryptoInfo
	for rows.Next() {
		var c cryptoInfo
		if err := rows.Scan(&c.ID, &c.FullCode); err != nil {
			return err
		}
		cryptos = append(cryptos, c)
	}

	if len(cryptos) == 0 {
		log.Println("⚠️ No crypto currencies with full_code found")
		return nil
	}

	// Собираем список full_code для запроса
	fullCodes := make([]string, len(cryptos))
	codeToID := make(map[string]string)
	for i, c := range cryptos {
		fullCodes[i] = c.FullCode
		codeToID[c.FullCode] = c.ID
	}

	// Запрос к CoinGecko
	req, err := http.NewRequestWithContext(ctx, "GET", w.cfg.CoinGeckoApiUrl+"/simple/price", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("vs_currencies", "rub")
	q.Add("ids", strings.Join(fullCodes, ","))
	req.URL.RawQuery = q.Encode()

	if w.cfg.CoinGeckoToken != "" {
		req.Header.Set("x-cg-demo-api-key", w.cfg.CoinGeckoToken)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch CoinGecko rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CoinGecko returned status %d", resp.StatusCode)
	}

	// Парсим ответ: {"bitcoin":{"rub":4712189},...}
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Сохраняем только те, что есть у нас
	for fullCode, rates := range result {
		currencyID, ok := codeToID[fullCode]
		if !ok {
			continue
		}

		rate, ok := rates["rub"]
		if !ok {
			log.Printf("⚠️ No RUB rate for %s", fullCode)
			continue
		}

		_, err = w.pool.Exec(ctx, `
			INSERT INTO currency_rates (currency_id, rate, source, updated_at)
			VALUES ($1, $2, $3, NOW())
		`, currencyID, rate, currency.RateSourceCoinGecko)
		if err != nil {
			log.Printf("⚠️ Failed to save rate for %s: %v", currencyID, err)
		}
	}

	// ревокаем кэш курсов
	w.service.InvalidateCache()

	return nil
}

func (w *CryptoWorker) Stop() {
	w.cron.Stop()
}

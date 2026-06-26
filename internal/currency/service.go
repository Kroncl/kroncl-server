package currency

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool       *pgxpool.Pool
	cache      map[string]float64 // key: "USD_2026-06-23" = 73.44, "BTC_2026-06-23_15" = 4712189
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
	cacheTimer *time.Timer
}

func NewService(pool *pgxpool.Pool) *Service {
	s := &Service{
		pool:     pool,
		cache:    make(map[string]float64),
		cacheTTL: 5 * time.Minute,
	}
	s.resetCacheTimer()
	return s
}

func (s *Service) GetAll(ctx context.Context, codes []string) ([]Currency, error) {
	rub := s.rubCurrency()

	hasFilter := len(codes) > 0
	addRUB := !hasFilter

	// Фильтруем RUB и проверяем, нужен ли он
	filtered := make([]string, 0, len(codes))
	for _, code := range codes {
		if strings.ToUpper(code) == "RUB" {
			addRUB = true
		} else {
			filtered = append(filtered, code)
		}
	}

	// Если только RUB — возвращаем сразу
	if hasFilter && len(filtered) == 0 && addRUB {
		return []Currency{rub}, nil
	}

	query := `
		SELECT c.id, c.name, c.type, c.symbol, cr.rate, cr.source, cr.updated_at
		FROM currencies c
		JOIN LATERAL (
			SELECT rate, source, updated_at
			FROM currency_rates
			WHERE currency_id = c.id
			ORDER BY updated_at DESC
			LIMIT 1
		) cr ON true
	`

	var args []interface{}
	if len(filtered) > 0 {
		placeholders := make([]string, len(filtered))
		for i, code := range filtered {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args = append(args, code)
		}
		query += fmt.Sprintf(" WHERE c.id IN (%s)", strings.Join(placeholders, ", "))
	}

	query += " ORDER BY c.type, c.name"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []Currency
	for rows.Next() {
		var c Currency
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Symbol, &c.Rate.Rate, &c.Rate.Source, &c.Rate.UpdatedAt); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}

	if addRUB {
		currencies = append([]Currency{rub}, currencies...)
	}

	return currencies, rows.Err()
}

func (s *Service) GetByID(ctx context.Context, id string) (*Currency, error) {
	if strings.ToUpper(id) == "RUB" {
		rub := s.rubCurrency()
		return &rub, nil
	}

	var c Currency
	err := s.pool.QueryRow(ctx, `
		SELECT c.id, c.name, c.type, c.symbol, cr.rate, cr.source, cr.updated_at
		FROM currencies c
		JOIN LATERAL (
			SELECT rate, source, updated_at
			FROM currency_rates
			WHERE currency_id = c.id
			ORDER BY updated_at DESC
			LIMIT 1
		) cr ON true
		WHERE c.id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Type, &c.Symbol, &c.Rate.Rate, &c.Rate.Source, &c.Rate.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// utils
func (s *Service) rubCurrency() Currency {
	return Currency{
		ID:     "RUB",
		Name:   "Российский рубль",
		Type:   "fiat",
		Symbol: "₽",
		Rate: CurrencyRate{
			Rate:      1,
			Source:    "manual",
			UpdatedAt: time.Now(),
		},
	}
}

// -------------
// CONVERTING [CACHED]
// -------------

// cacheKey — единый для всех, точность до часа
func (s *Service) cacheKey(currencyID string, date time.Time) string {
	return fmt.Sprintf("%s_%s_%02d", currencyID, date.Format("2006-01-02"), date.Hour())
}

// GetRate возвращает курс валюты к RUB на указанное время (ближайший до)
func (s *Service) GetRate(ctx context.Context, currencyID string, date time.Time) (float64, error) {
	if strings.ToUpper(currencyID) == "RUB" {
		return 1, nil
	}

	cacheKey := s.cacheKey(currencyID, date)

	s.cacheMu.RLock()
	if rate, ok := s.cache[cacheKey]; ok {
		s.cacheMu.RUnlock()
		return rate, nil
	}
	s.cacheMu.RUnlock()

	var rate float64
	err := s.pool.QueryRow(ctx, `
		SELECT rate FROM currency_rates
		WHERE currency_id = $1
		ORDER BY ABS(EXTRACT(EPOCH FROM (updated_at - $2))) ASC
		LIMIT 1
	`, currencyID, date).Scan(&rate)
	if err != nil {
		return 0, fmt.Errorf("no rate for %s before %s: %w", currencyID, date.Format("2006-01-02 15:04"), err)
	}

	s.cacheMu.Lock()
	s.cache[cacheKey] = rate
	s.cacheMu.Unlock()

	return rate, nil
}

// GetRates массово получает курсы к RUB
func (s *Service) GetRates(ctx context.Context, pairs []CurrencyDatePair) (map[string]float64, error) {
	if len(pairs) == 0 {
		return map[string]float64{}, nil
	}

	result := make(map[string]float64)
	var missing []CurrencyDatePair
	missingKeys := make(map[string]string) // cacheKey -> currencyID

	s.cacheMu.RLock()
	for _, p := range pairs {
		if strings.ToUpper(p.CurrencyID) == "RUB" {
			result[s.cacheKey("RUB", p.Date)] = 1
			continue
		}
		cacheKey := s.cacheKey(p.CurrencyID, p.Date)
		if rate, ok := s.cache[cacheKey]; ok {
			result[cacheKey] = rate
		} else {
			missing = append(missing, p)
			missingKeys[cacheKey] = p.CurrencyID
		}
	}
	s.cacheMu.RUnlock()

	if len(missing) == 0 {
		return result, nil
	}

	// Убираем дубликаты
	unique := make(map[string]CurrencyDatePair)
	for _, p := range missing {
		key := s.cacheKey(p.CurrencyID, p.Date)
		if _, exists := unique[key]; !exists {
			unique[key] = p
		}
	}

	// Строим запрос, передаём cache_key
	query := `SELECT req.cache_key, cr.rate FROM (VALUES `
	var args []interface{}
	idx := 1
	var rows []string

	for cacheKey, p := range unique {
		rows = append(rows, fmt.Sprintf("($%d, $%d::timestamptz, $%d)", idx, idx+1, idx+2))
		args = append(args, p.CurrencyID, p.Date, cacheKey)
		idx += 3
	}
	query += strings.Join(rows, ", ")
	query += `) AS req(currency_id, date, cache_key)
		CROSS JOIN LATERAL (
			SELECT rate
			FROM currency_rates cr
			WHERE cr.currency_id = req.currency_id
			ORDER BY ABS(EXTRACT(EPOCH FROM (cr.updated_at - req.date))) ASC
			LIMIT 1
		) cr`

	queryRows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}
	defer queryRows.Close()

	s.cacheMu.Lock()
	for queryRows.Next() {
		var cacheKey string
		var rate float64
		if err := queryRows.Scan(&cacheKey, &rate); err != nil {
			s.cacheMu.Unlock()
			return nil, err
		}
		s.cache[cacheKey] = rate
		result[cacheKey] = rate
	}
	s.cacheMu.Unlock()

	return result, nil
}

// Convert конвертирует сумму из одной валюты в другую
func (s *Service) Convert(ctx context.Context, amount float64, fromCurrency, toCurrency string, date time.Time) (float64, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	// Конвертация через RUB: from -> RUB -> to
	pairs := []CurrencyDatePair{
		{CurrencyID: fromCurrency, Date: date},
		{CurrencyID: toCurrency, Date: date},
	}

	rates, err := s.GetRates(ctx, pairs)
	if err != nil {
		return 0, err
	}

	fromKey := s.cacheKey(fromCurrency, date)
	toKey := s.cacheKey(toCurrency, date)

	fromRate, ok := rates[fromKey]
	if !ok {
		return 0, fmt.Errorf("no rate for %s", fromCurrency)
	}
	toRate, ok := rates[toKey]
	if !ok {
		return 0, fmt.Errorf("no rate for %s", toCurrency)
	}

	// amount в RUB, затем в целевую валюту
	rubAmount := amount * fromRate
	return rubAmount / toRate, nil
}

func (s *Service) resetCacheTimer() {
	if s.cacheTimer != nil {
		s.cacheTimer.Stop()
	}
	s.cacheTimer = time.AfterFunc(s.cacheTTL, func() {
		s.cacheMu.Lock()
		s.cache = make(map[string]float64)
		s.cacheMu.Unlock()
	})
}

func (s *Service) InvalidateCache() {
	s.cacheMu.Lock()
	s.cache = make(map[string]float64)
	s.cacheMu.Unlock()
	s.resetCacheTimer()
}

// ---------------
// CONVERTATION
// ---------------

func (s *Service) ConvertSummary(
	ctx context.Context,
	rawStats map[string]map[time.Time]*RawStat,
	targetCurrency string,
) (*ConvertedSummary, error) {

	if targetCurrency == "" {
		targetCurrency = "RUB"
	}

	// Собираем пары для запроса курсов — все fromCurrency + targetCurrency на каждую дату
	var pairs []CurrencyDatePair
	datesMap := make(map[time.Time]bool)

	for currencyID, dates := range rawStats {
		for date := range dates {
			datesMap[date] = true
			pairs = append(pairs, CurrencyDatePair{
				CurrencyID: currencyID,
				Date:       date,
			})
		}
	}

	// Добавляем курсы для целевой валюты на все те же даты
	if targetCurrency != "RUB" {
		for date := range datesMap {
			pairs = append(pairs, CurrencyDatePair{
				CurrencyID: targetCurrency,
				Date:       date,
			})
		}
	}

	rates, err := s.GetRates(ctx, pairs)
	if err != nil {
		return nil, err
	}

	// Конвертируем каждую группу напрямую в целевую валюту
	var totalIncome, totalExpense float64
	var totalCount int64

	for currencyID, dates := range rawStats {
		for date, stat := range dates {
			// Получаем курс fromCurrency → RUB
			fromRate := 1.0
			if currencyID != "RUB" {
				key := s.cacheKey(currencyID, date)
				var ok bool
				fromRate, ok = rates[key]
				if !ok {
					continue
				}
			}

			// Получаем курс targetCurrency → RUB (или 1 если RUB)
			toRate := 1.0
			if targetCurrency != "RUB" {
				key := s.cacheKey(targetCurrency, date)
				var ok bool
				toRate, ok = rates[key]
				if !ok {
					continue
				}
			}

			// amount × (fromRate / toRate) — конвертация на дату транзакции
			rate := fromRate / toRate
			totalIncome += stat.Income * rate
			totalExpense += stat.Expense * rate
			totalCount += stat.Count
		}
	}

	netBalance := totalIncome - totalExpense

	avgTransaction := 0.0
	if totalCount > 0 {
		avgTransaction = totalIncome / float64(totalCount)
	}

	targetCur, _ := s.GetByID(ctx, targetCurrency)

	if targetCur != nil && targetCur.Type == "fiat" {
		totalIncome = math.Round(totalIncome*100) / 100
		totalExpense = math.Round(totalExpense*100) / 100
		netBalance = math.Round(netBalance*100) / 100
		if totalCount > 0 {
			avgTransaction = math.Round(avgTransaction*100) / 100
		}
	}

	return &ConvertedSummary{
		TotalIncome:      totalIncome,
		TotalExpense:     totalExpense,
		NetBalance:       netBalance,
		TransactionCount: totalCount,
		AvgTransaction:   avgTransaction,
		Currency:         targetCur,
	}, nil
}

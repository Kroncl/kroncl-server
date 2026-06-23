package currency

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
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

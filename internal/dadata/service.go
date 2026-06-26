package dadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kroncl-server/internal/config"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
	cfg  *config.DaDataConfig
}

func NewService(
	pool *pgxpool.Pool,
	cfg *config.DaDataConfig,
) *Service {
	s := &Service{
		pool: pool,
		cfg:  cfg,
	}
	return s
}

func (s *Service) FindPartyByINN(ctx context.Context, inn string) (*PartySuggestion, error) {
	req := FindPartyRequest{Query: inn, BranchType: "MAIN"}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.cfg.ApiUrl+"/findById/party", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Token "+s.cfg.ApiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call DaData: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DaData returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result FindPartyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Suggestions) == 0 {
		return nil, fmt.Errorf("company not found by INN: %s", inn)
	}

	return &result.Suggestions[0], nil
}
func (s *Service) BuildCounterpartyPreview(party *PartySuggestion) *CounterpartyPreview {
	var cpType string
	switch party.Data.Type {
	case "INDIVIDUAL":
		cpType = "person"
	case "LEGAL":
		cpType = "legal"
	default:
		cpType = "legal"
	}

	return &CounterpartyPreview{
		Name:    party.Data.Name.ShortWithOPF,
		INN:     party.Data.INN,
		KPP:     party.Data.KPP,
		OGRN:    party.Data.OGRN,
		Address: party.Data.Address.UnrestrictedValue,
		Type:    cpType,
	}
}

func (s *Service) SuggestParty(ctx context.Context, query string) ([]PartySuggestion, error) {
	req := FindPartyRequest{Query: query}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.cfg.ApiUrl+"/suggest/party", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Token "+s.cfg.ApiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call DaData: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DaData returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result FindPartyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Suggestions, nil
}

package docs

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
)

// GetSettings возвращает настройки документов для компании
func (s *Service) GetSettings(ctx context.Context) (*DocsSettings, error) {
	query := `
		SELECT 
			legal_name, legal_address, inn, ogrn, 
			bank_name, bank_bic, bank_account,
			director_name, accountant_name,
			warranty_terms, additional_terms,
			created_at, updated_at
		FROM docs_settings
		LIMIT 1
	`

	var settings DocsSettings
	err := s.pool.QueryRow(ctx, query).Scan(
		&settings.LegalName,
		&settings.LegalAddress,
		&settings.Inn,
		&settings.Ogrn,
		&settings.BankName,
		&settings.BankBic,
		&settings.BankAccount,
		&settings.DirectorName,
		&settings.AccountantName,
		&settings.WarrantyTerms,
		&settings.AdditionalTerms,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get docs settings: %w", err)
	}

	return &settings, nil
}

// UpdateSettings обновляет настройки документов для компании
func (s *Service) UpdateSettings(ctx context.Context, req UpdateDocsSettingsRequest) (*DocsSettings, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Проверяем, существует ли запись
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM docs_settings)`).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check existence: %w", err)
	}

	if !exists {
		// Создаём пустую запись
		_, err = tx.Exec(ctx, `
			INSERT INTO docs_settings (created_at, updated_at) 
			VALUES (NOW(), NOW())
		`)
		if err != nil {
			return nil, fmt.Errorf("failed to create settings record: %w", err)
		}
	}

	// Обновляем поля
	updater := core.NewUpdater("docs_settings")

	if req.LegalName != nil {
		if *req.LegalName == "" {
			updater.SetNull("legal_name")
		} else {
			updater.SetString("legal_name", *req.LegalName)
		}
	}
	if req.LegalAddress != nil {
		if *req.LegalAddress == "" {
			updater.SetNull("legal_address")
		} else {
			updater.SetString("legal_address", *req.LegalAddress)
		}
	}
	if req.Inn != nil {
		if *req.Inn == "" {
			updater.SetNull("inn")
		} else {
			updater.SetString("inn", *req.Inn)
		}
	}
	if req.Ogrn != nil {
		if *req.Ogrn == "" {
			updater.SetNull("ogrn")
		} else {
			updater.SetString("ogrn", *req.Ogrn)
		}
	}
	if req.BankName != nil {
		if *req.BankName == "" {
			updater.SetNull("bank_name")
		} else {
			updater.SetString("bank_name", *req.BankName)
		}
	}
	if req.BankBic != nil {
		if *req.BankBic == "" {
			updater.SetNull("bank_bic")
		} else {
			updater.SetString("bank_bic", *req.BankBic)
		}
	}
	if req.BankAccount != nil {
		if *req.BankAccount == "" {
			updater.SetNull("bank_account")
		} else {
			updater.SetString("bank_account", *req.BankAccount)
		}
	}
	if req.DirectorName != nil {
		if *req.DirectorName == "" {
			updater.SetNull("director_name")
		} else {
			updater.SetString("director_name", *req.DirectorName)
		}
	}
	if req.AccountantName != nil {
		if *req.AccountantName == "" {
			updater.SetNull("accountant_name")
		} else {
			updater.SetString("accountant_name", *req.AccountantName)
		}
	}
	if req.WarrantyTerms != nil {
		if *req.WarrantyTerms == "" {
			updater.SetNull("warranty_terms")
		} else {
			updater.SetString("warranty_terms", *req.WarrantyTerms)
		}
	}
	if req.AdditionalTerms != nil {
		if *req.AdditionalTerms == "" {
			updater.SetNull("additional_terms")
		} else {
			updater.SetString("additional_terms", *req.AdditionalTerms)
		}
	}

	query, args := updater.Build()
	if query != "" {
		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update docs settings: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetSettings(ctx)
}

package hrm

import (
	"context"
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
)

// validateNoConflict проверяет, что increase и reduce не противоречат друг другу
// Нельзя одновременно добавить и исключить одно и то же разрешение
func validateNoConflict(increase, reduce []string) error {
	increaseMap := make(map[string]bool)
	for _, p := range increase {
		increaseMap[p] = true
	}

	for _, p := range reduce {
		if increaseMap[p] {
			return fmt.Errorf("permission %s appears in both increase and reduce lists", p)
		}
	}

	return nil
}

// GetAccountSettings возвращает настройки аккаунта, если записи нет - создаёт
func (r *Repository) GetAccountSettings(ctx context.Context, accountID string) (*AccountSettings, error) {
	// Пробуем получить существующие настройки
	query := `
		SELECT account_id, increase_permissions, reduce_permissions, created_at, updated_at
		FROM accounts_settings
		WHERE account_id = $1
	`

	var settings AccountSettings
	var increaseJSON, reducingJSON []byte

	err := r.pool.QueryRow(ctx, query, accountID).Scan(
		&settings.AccountID,
		&increaseJSON,
		&reducingJSON,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	// Если записи нет - создаём
	if err != nil {
		// Создаём запись с пустыми разрешениями
		emptyJSON := []byte("[]")
		createQuery := `
			INSERT INTO accounts_settings (account_id, increase_permissions, reduce_permissions, created_at, updated_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING account_id, increase_permissions, reduce_permissions, created_at, updated_at
		`

		err = r.pool.QueryRow(ctx, createQuery, accountID, emptyJSON, emptyJSON).Scan(
			&settings.AccountID,
			&increaseJSON,
			&reducingJSON,
			&settings.CreatedAt,
			&settings.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create account settings: %w", err)
		}
	}

	// Парсим increase_permissions
	if len(increaseJSON) > 0 {
		if err := json.Unmarshal(increaseJSON, &settings.IncreasePermissions); err != nil {
			return nil, fmt.Errorf("failed to parse increase_permissions: %w", err)
		}
	} else {
		settings.IncreasePermissions = []string{}
	}

	// Парсим reduce_permissions
	if len(reducingJSON) > 0 {
		if err := json.Unmarshal(reducingJSON, &settings.ReducePermissions); err != nil {
			return nil, fmt.Errorf("failed to parse reduce_permissions: %w", err)
		}
	} else {
		settings.ReducePermissions = []string{}
	}

	settings.AccountID = accountID

	return &settings, nil
}

// UpsertAccountSettings создает или обновляет настройки аккаунта
func (r *Repository) UpsertAccountSettings(ctx context.Context, accountID string, req UpdateAccountSettingsRequest) (*AccountSettings, error) {
	// Получаем текущие настройки (если есть)
	current, _ := r.GetAccountSettings(ctx, accountID)

	// Определяем финальные списки разрешений
	var increasePerms, reducingPerms []string

	if req.IncreasePermissions != nil {
		increasePerms = req.IncreasePermissions
	} else if current != nil {
		increasePerms = current.IncreasePermissions
	} else {
		increasePerms = []string{}
	}

	if req.ReducePermissions != nil {
		reducingPerms = req.ReducePermissions
	} else if current != nil {
		reducingPerms = current.ReducePermissions
	} else {
		reducingPerms = []string{}
	}

	// Убираем дубликаты
	increasePerms = config.UniquePermissions(increasePerms)
	reducingPerms = config.UniquePermissions(reducingPerms)

	// Проверяем, что increase и reduce не конфликтуют
	if err := validateNoConflict(increasePerms, reducingPerms); err != nil {
		return nil, err
	}

	// Проверяем валидность разрешений
	if invalid := config.ValidatePermissions(increasePerms); len(invalid) > 0 {
		return nil, fmt.Errorf("invalid increase_permissions: %v", invalid)
	}
	if invalid := config.ValidatePermissions(reducingPerms); len(invalid) > 0 {
		return nil, fmt.Errorf("invalid reduce_permissions: %v", invalid)
	}

	// Маршалим JSON
	increaseJSON, err := json.Marshal(increasePerms)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal increase_permissions: %w", err)
	}

	reducingJSON, err := json.Marshal(reducingPerms)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reduce_permissions: %w", err)
	}

	// Upsert запрос
	query := `
		INSERT INTO accounts_settings (account_id, increase_permissions, reduce_permissions, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (account_id) DO UPDATE SET
			increase_permissions = $2,
			reduce_permissions = $3,
			updated_at = CURRENT_TIMESTAMP
		RETURNING account_id, increase_permissions, reduce_permissions, created_at, updated_at
	`

	var settings AccountSettings
	var returnedIncreaseJSON, returnedReducingJSON []byte

	err = r.pool.QueryRow(ctx, query, accountID, increaseJSON, reducingJSON).Scan(
		&settings.AccountID,
		&returnedIncreaseJSON,
		&returnedReducingJSON,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert account settings: %w", err)
	}

	// Парсим результат
	if len(returnedIncreaseJSON) > 0 {
		if err := json.Unmarshal(returnedIncreaseJSON, &settings.IncreasePermissions); err != nil {
			return nil, fmt.Errorf("failed to parse increase_permissions: %w", err)
		}
	} else {
		settings.IncreasePermissions = []string{}
	}

	if len(returnedReducingJSON) > 0 {
		if err := json.Unmarshal(returnedReducingJSON, &settings.ReducePermissions); err != nil {
			return nil, fmt.Errorf("failed to parse reduce_permissions: %w", err)
		}
	} else {
		settings.ReducePermissions = []string{}
	}

	settings.AccountID = accountID

	return &settings, nil
}

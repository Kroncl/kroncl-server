package dm

import (
	"context"
	"fmt"
	"kroncl-server/internal/core"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ---------
// Вспомогательные методы для обновления связанных сущностей
// ---------

func (r *Repository) updateDealBase(ctx context.Context, tx pgx.Tx, id string, req UpdateDealRequest) error {
	updater := core.NewUpdater("deals")

	if req.Comment != nil {
		if *req.Comment == "" {
			updater.SetNull("comment")
		} else {
			comment := strings.TrimSpace(*req.Comment)
			updater.SetString("comment", comment)
		}
	}

	if req.TypeID != nil {
		if *req.TypeID == "" {
			updater.SetNull("type_id")
		} else {
			updater.SetString("type_id", *req.TypeID)
		}
	}

	query, args := updater.Where("id = $1", id).Build()
	if query == "" {
		return nil
	}

	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update deal base: %w", err)
	}

	return nil
}

func (r *Repository) updateDealClient(ctx context.Context, tx pgx.Tx, dealID, clientID string) error {
	// Удаляем старую связь
	_, err := tx.Exec(ctx, `DELETE FROM deal_client WHERE deal_id = $1`, dealID)
	if err != nil {
		return fmt.Errorf("failed to remove old client link: %w", err)
	}

	// Создаем новую связь
	linkID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO deal_client (id, deal_id, client_id, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, linkID, dealID, clientID)
	if err != nil {
		return fmt.Errorf("failed to link client: %w", err)
	}

	return nil
}

func (r *Repository) updateDealStatus(ctx context.Context, tx pgx.Tx, dealID, statusID string) error {
	// Удаляем старый статус
	_, err := tx.Exec(ctx, `DELETE FROM deal_status WHERE deal_id = $1`, dealID)
	if err != nil {
		return fmt.Errorf("failed to remove old status: %w", err)
	}

	// Создаем новый статус
	linkID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO deal_status (id, deal_id, status_id, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, linkID, dealID, statusID)
	if err != nil {
		return fmt.Errorf("failed to link status: %w", err)
	}

	return nil
}

func (r *Repository) updateDealEmployees(ctx context.Context, tx pgx.Tx, dealID string, employeeIDs []string) error {
	// Удаляем всех старых сотрудников
	_, err := tx.Exec(ctx, `DELETE FROM deal_employees WHERE deal_id = $1`, dealID)
	if err != nil {
		return fmt.Errorf("failed to remove old employees: %w", err)
	}

	// Добавляем новых сотрудников
	for _, empID := range employeeIDs {
		linkID := uuid.New().String()
		_, err = tx.Exec(ctx, `
			INSERT INTO deal_employees (id, deal_id, employee_id, created_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		`, linkID, dealID, empID)
		if err != nil {
			return fmt.Errorf("failed to link employee %s: %w", empID, err)
		}
	}

	return nil
}

func (r *Repository) updateDealPositions(ctx context.Context, tx pgx.Tx, dealID string, positions []UpdateDealPosition) error {
	// Удаляем все позиции, которые помечены на удаление
	for _, pos := range positions {
		if pos.Delete != nil && *pos.Delete && pos.ID != nil {
			_, err := tx.Exec(ctx, `DELETE FROM deal_positions WHERE id = $1`, *pos.ID)
			if err != nil {
				return fmt.Errorf("failed to delete position %s: %w", *pos.ID, err)
			}
		}
	}

	// Обновляем/создаем позиции
	for _, pos := range positions {
		if pos.Delete != nil && *pos.Delete {
			continue
		}

		if pos.ID != nil {
			// Обновляем существующую позицию
			updater := core.NewUpdater("deal_positions")

			if pos.Name != nil {
				name := strings.TrimSpace(*pos.Name)
				if name != "" {
					updater.SetString("name", name)
				}
			}

			if pos.Comment != nil {
				if *pos.Comment == "" {
					updater.SetNull("comment")
				} else {
					comment := strings.TrimSpace(*pos.Comment)
					updater.SetString("comment", comment)
				}
			}

			if pos.Price != nil {
				updater.SetFloat("price", *pos.Price)
			}

			if pos.Quantity != nil {
				updater.SetFloat("quantity", *pos.Quantity)
			}

			if pos.Unit != nil {
				unit := strings.TrimSpace(*pos.Unit)
				if unit != "" {
					updater.SetString("unit", unit)
				}
			}

			if pos.UnitID != nil {
				if *pos.UnitID == "" {
					updater.SetNull("unit_id")
				} else {
					updater.SetString("unit_id", *pos.UnitID)
				}
			}

			if pos.PositionID != nil {
				if *pos.PositionID == "" {
					updater.SetNull("position_id")
				} else {
					updater.SetString("position_id", *pos.PositionID)
				}
			}

			query, args := updater.Where("id = $1", *pos.ID).Build()
			if query != "" {
				_, err := tx.Exec(ctx, query, args...)
				if err != nil {
					return fmt.Errorf("failed to update position %s: %w", *pos.ID, err)
				}
			}
		} else {
			// Создаем новую позицию
			newID := uuid.New().String()

			query := `
				INSERT INTO deal_positions (
					id, deal_id, name, comment, price, quantity, unit, 
					unit_id, position_id, created_at, updated_at
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, 
					CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
				)
			`

			name := ""
			if pos.Name != nil {
				name = *pos.Name
			}

			price := 0.0
			if pos.Price != nil {
				price = *pos.Price
			}

			quantity := 1.0
			if pos.Quantity != nil {
				quantity = *pos.Quantity
			}

			unit := "pcs"
			if pos.Unit != nil {
				unit = *pos.Unit
			}

			_, err := tx.Exec(ctx, query,
				newID,
				dealID, // ← ВАЖНО: передаём deal_id
				name,
				pos.Comment,
				price,
				quantity,
				unit,
				pos.UnitID,
				pos.PositionID,
			)
			if err != nil {
				return fmt.Errorf("failed to create position: %w", err)
			}
		}
	}

	return nil
}

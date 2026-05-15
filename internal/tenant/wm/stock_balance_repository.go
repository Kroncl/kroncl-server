package wm

import (
	"context"
	"fmt"
)

// GetStockBalance возвращает текущие остатки по всем товарам
func (r *Repository) GetStockBalance(ctx context.Context, unitID *string) ([]StockBalanceItem, error) {
	query := `
        SELECT 
            u.id, u.name, u.unit, u.sale_price, u.purchase_price,
            u.inventory_type, u.tracking_detail, u.tracked_type,
            COALESCE(SUM(CASE WHEN sb.direction = 'income' THEN sp.quantity ELSE 0 END), 0) as income_qty,
            COALESCE(SUM(CASE WHEN sb.direction = 'outcome' THEN sp.quantity ELSE 0 END), 0) as outcome_qty
        FROM catalog_units u
        LEFT JOIN stock_positions sp ON u.id = sp.unit_id
        LEFT JOIN stock_position_batch spb ON sp.id = spb.position_id
        LEFT JOIN stock_batches sb ON spb.batch_id = sb.id
        WHERE u.status = 'active'
          AND u.inventory_type = 'tracked'
    `

	var args []interface{}
	argIndex := 1

	if unitID != nil && *unitID != "" {
		query += " AND u.id = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *unitID)
		argIndex++
	}

	query += " GROUP BY u.id ORDER BY u.name ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock balance: %w", err)
	}
	defer rows.Close()

	var balances []StockBalanceItem
	for rows.Next() {
		var item StockBalanceItem
		var unit CatalogUnit
		var incomeQty, outcomeQty float64

		err := rows.Scan(
			&unit.ID, &unit.Name, &unit.Unit, &unit.SalePrice, &unit.PurchasePrice,
			&unit.InventoryType, &unit.TrackingDetail, &unit.TrackedType,
			&incomeQty, &outcomeQty,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock balance: %w", err)
		}

		item.UnitID = unit.ID
		item.UnitName = unit.Name
		item.Unit = unit
		item.Quantity = incomeQty - outcomeQty
		item.Available = item.Quantity // если нет резерва

		balances = append(balances, item)
	}

	return balances, nil
}

// GetUnitStockBalance возвращает остаток по конкретному товару
func (r *Repository) GetUnitStockBalance(ctx context.Context, unitID string) (float64, error) {
	balances, err := r.GetStockBalance(ctx, &unitID)
	if err != nil {
		return 0, err
	}
	if len(balances) == 0 {
		return 0, nil
	}
	return balances[0].Quantity, nil
}

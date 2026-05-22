package wm

import (
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

const stockBalanceFilePrefix = "reports/kroncl_stock_balance_"

func (r *Repository) writeStockBalanceSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	balances, err := r.GetStockBalance(ctx, nil)
	if err != nil {
		return 0, err
	}

	f.NewSheet(sheetName)

	headers := []string{"Товар", "Остаток", "Единица", "Цена продажи", "Валюта", "Тип учета", "ID товара"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	moneyStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 2})

	for i, balance := range balances {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), balance.UnitName)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), balance.Quantity)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), balance.Unit.Unit)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), balance.Unit.SalePrice)
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), moneyStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), balance.Unit.Currency)

		inventoryType := "Складской"
		if balance.Unit.InventoryType == InventoryTypeUntracked {
			inventoryType = "Не складской"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), inventoryType)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), balance.UnitID)
	}

	summaryRow := len(balances) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО ПОЗИЦИЙ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), len(balances))

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 20)
	}

	return len(balances), nil
}

func (r *Repository) GenerateStockBalanceReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeStockBalanceSheet, stockBalanceFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

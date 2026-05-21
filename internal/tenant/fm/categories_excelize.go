package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const categoriesFilePrefix = "reports/kroncl_transactions_categories_"

func (r *Repository) writeCategoriesSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	categories, total, err := r.GetCategories(ctx, 0, config.MAX_EXCEL_SHEET_ROWS, nil, "")
	if err != nil {
		return 0, err
	}

	headers := []string{"Название", "Направление", "Описание", "Системная", "ID категории"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, cat := range categories {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), cat.Name)

		direction := "Расход"
		if cat.Direction == TransactionCategoryDirectionIncome {
			direction = "Доход"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), direction)

		if cat.Description != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), *cat.Description)
		}

		system := "Нет"
		if cat.System {
			system = "Да"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), system)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), cat.ID)
	}

	summaryRow := len(categories) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО КАТЕГОРИЙ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 20)
	}

	return total, nil
}

func (r *Repository) GenerateCategoriesReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeCategoriesSheet, categoriesFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

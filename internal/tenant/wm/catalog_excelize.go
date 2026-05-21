package wm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const catalogCategoriesFilePrefix = "reports/kroncl_catalog_categories_"
const catalogUnitsFilePrefix = "reports/kroncl_catalog_units_"

func (r *Repository) writeCatalogCategoriesSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	req := GetCategoriesRequest{
		Page:  1,
		Limit: config.MAX_EXCEL_SHEET_ROWS,
	}

	categories, total, err := r.GetCatalogCategories(ctx, req)
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many categories: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Название", "Статус", "Родительская категория", "Комментарий", "ID категории"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, cat := range categories {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), cat.Name)

		status := "Активна"
		if cat.Status == CategoryStatusInactive {
			status = "Неактивна"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), status)

		parentName := ""
		if cat.ParentID != nil && *cat.ParentID != "" {
			parent, err := r.GetCatalogCategoryByID(ctx, *cat.ParentID)
			if err == nil {
				parentName = parent.Name
			}
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), parentName)

		if cat.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *cat.Comment)
		}
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

	return int(total), nil
}

func (r *Repository) GenerateCatalogCategoriesReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeCatalogCategoriesSheet, catalogCategoriesFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

func (r *Repository) writeCatalogUnitsSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	req := GetUnitsRequest{
		Page:  1,
		Limit: config.MAX_EXCEL_SHEET_ROWS,
	}

	units, total, err := r.GetCatalogUnits(ctx, req)
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many units: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Название", "Тип", "Статус", "Единица", "Цена продажи", "Валюта", "Категория", "Комментарий", "ID товара"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	moneyStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 2})

	for i, unit := range units {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), unit.Name)

		unitType := "Товар"
		if unit.Type == UnitTypeService {
			unitType = "Услуга"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), unitType)

		status := "Активен"
		if unit.Status == UnitStatusInactive {
			status = "Неактивен"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), status)

		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), unit.Unit)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), unit.SalePrice)
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), moneyStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), unit.Currency)

		categoryName := ""
		if unit.CategoryID != "" {
			cat, err := r.GetCatalogCategoryByID(ctx, unit.CategoryID)
			if err == nil {
				categoryName = cat.Name
			}
		}
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), categoryName)

		if unit.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), *unit.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), unit.ID)
	}

	summaryRow := len(units) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО ТОВАРОВ/УСЛУГ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 20)
	}

	return int(total), nil
}

func (r *Repository) GenerateCatalogUnitsReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeCatalogUnitsSheet, catalogUnitsFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

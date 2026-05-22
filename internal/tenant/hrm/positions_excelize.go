package hrm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const positionsFilePrefix = "reports/kroncl_positions_"

func (r *Repository) writePositionsSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	positions, total, err := r.GetPositions(ctx, 1, config.MAX_EXCEL_SHEET_ROWS, "")
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many positions: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Название", "Описание", "Количество разрешений", "ID должности"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, pos := range positions {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), pos.Name)

		if pos.Description != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), *pos.Description)
		}

		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), len(pos.Permissions))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), pos.ID)

		// Опционально: добавить список разрешений в примечание к ячейке
		if len(pos.Permissions) > 0 {
			comment := "Разрешения: " + strings.Join(pos.Permissions, ", ")
			f.AddComment(sheetName, excelize.Comment{
				Cell:   fmt.Sprintf("C%d", row),
				Author: "Kroncl",
				Text:   comment,
			})
		}
	}

	summaryRow := len(positions) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО ДОЛЖНОСТЕЙ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 25)
	}

	return int(total), nil
}

func (r *Repository) GeneratePositionsReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writePositionsSheet, positionsFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

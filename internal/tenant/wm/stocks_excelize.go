package wm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const stockBatchesFilePrefix = "reports/kroncl_stock_batches_"
const stockPositionsFilePrefix = "reports/kroncl_stock_positions_"

// writeStockBatchesSheet - отчет по партиям (поставкам)
func (r *Repository) writeStockBatchesSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	req := GetStockBatchesParams{
		Page:  1,
		Limit: config.MAX_EXCEL_SHEET_ROWS,
	}

	batches, total, err := r.GetStockBatches(ctx, req)
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many batches: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"ID партии", "Направление", "Комментарий", "Дата создания"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, batch := range batches {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), batch.ID)

		direction := "Приход"
		if batch.Direction == StockDirectionOutcome {
			direction = "Расход"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), direction)

		if batch.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), *batch.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), batch.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	summaryRow := len(batches) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО ПАРТИЙ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 30)
	}

	return int(total), nil
}

func (r *Repository) GenerateStockBatchesReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeStockBatchesSheet, stockBatchesFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

// writeStockPositionsSheet - отчет по позициям склада
func (r *Repository) writeStockPositionsSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	req := GetStockPositionsParams{
		Page:  1,
		Limit: config.MAX_EXCEL_SHEET_ROWS,
	}

	positions, total, err := r.GetStockPositions(ctx, req)
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many positions: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Товар", "Количество", "Тип позиции", "ID партии", "ID позиции", "Дата создания"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, pos := range positions {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), pos.Unit.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), pos.Quantity)

		posType := "Партионный"
		if pos.Type == StockPositionTypeSerial {
			posType = "Поштучный"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), posType)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), pos.BatchID)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), pos.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), pos.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	summaryRow := len(positions) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО ПОЗИЦИЙ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 25)
	}

	return int(total), nil
}

func (r *Repository) GenerateStockPositionsReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeStockPositionsSheet, stockPositionsFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

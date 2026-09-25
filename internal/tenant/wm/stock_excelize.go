package wm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	stockBatchesFilePrefix   = "reports/kroncl_stock_batches_"
	stockPositionsFilePrefix = "reports/kroncl_stock_positions_"
	stockMovementsFilePrefix = "reports/kroncl_stock_movements_"
	stockBalanceFilePrefix   = "reports/kroncl_stock_balance_"
)

// --------
// HELPERS
// --------

// writeHeader — пишет шапку и возвращает стиль
func writeHeader(f *excelize.File, sheet string, headers []string) {
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, style)
	}
}

// setColWidths — устанавливает ширину колонок
func setColWidths(f *excelize.File, sheet string, count, width int) {
	for col := 1; col <= count; col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheet, colLetter, colLetter, float64(width))
	}
}

// writeTotalRow — пишет итоговую строку «ВСЕГО: N»
func writeTotalRow(f *excelize.File, sheet string, row int, label string, value interface{}) {
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), label)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleBold)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), value)
}

// directionRu — человекочитаемое направление
func directionRu(d StockDirection) string {
	if d == StockDirectionOutcome {
		return "Расход"
	}
	return "Приход"
}

// batchStatusRu — человекочитаемый статус
func batchStatusRu(s StockBatchStatus) string {
	switch s {
	case StockBatchStatusDraft:
		return "Черновик"
	case StockBatchStatusLabeled:
		return "Этикетки напечатаны"
	case StockBatchStatusConfirmed:
		return "Проведено"
	case StockBatchStatusCancelled:
		return "Отменено"
	}
	return string(s)
}

// positionTypeRu — человекочитаемый тип позиции
func positionTypeRu(t StockPositionType) string {
	if t == StockPositionTypeSerial {
		return "Поштучный"
	}
	return "Партионный"
}

// movementTypeRu — человекочитаемый тип движения
func movementTypeRu(t StockMovementType) string {
	switch t {
	case StockMovementTypeWriteOff:
		return "Списание"
	case StockMovementTypeReturn:
		return "Возврат"
	case StockMovementTypeTransfer:
		return "Перемещение"
	case StockMovementTypeAdjustment:
		return "Корректировка"
	}
	return string(t)
}

// --------
// STOCK BATCHES
// --------

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

	headers := []string{
		"ID партии",
		"Направление",
		"Статус",
		"Комментарий",
		"Дата создания",
		"Дата обновления",
	}
	writeHeader(f, sheetName, headers)

	for i, batch := range batches {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), batch.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), directionRu(batch.Direction))
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), batchStatusRu(batch.Status))
		if batch.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *batch.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), batch.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), batch.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	writeTotalRow(f, sheetName, len(batches)+4, "ВСЕГО ПАРТИЙ:", total)
	setColWidths(f, sheetName, len(headers), 25)

	return int(total), nil
}

func (r *Repository) GenerateStockBatchesReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeStockBatchesSheet, stockBatchesFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

// --------
// STOCK POSITIONS
// --------

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

	headers := []string{
		"ID позиции",
		"Товар",
		"Тип",
		"Количество",
		"Остаток",
		"Ед. изм.",
		"Цена за ед.",
		"Стоимость",
		"Производитель",
		"ID партии прихода",
		"Дата создания",
	}
	writeHeader(f, sheetName, headers)

	for i, pos := range positions {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), pos.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), pos.Unit.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), positionTypeRu(pos.Type))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), pos.Quantity)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), pos.Remaining)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), pos.Unit.Unit)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), pos.UnitPrice)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), pos.UnitPrice*pos.Quantity)
		if pos.Maker != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), *pos.Maker)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), pos.IncomeBatchID)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), pos.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	writeTotalRow(f, sheetName, len(positions)+4, "ВСЕГО ПОЗИЦИЙ:", total)
	setColWidths(f, sheetName, len(headers), 22)

	return int(total), nil
}

func (r *Repository) GenerateStockPositionsReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeStockPositionsSheet, stockPositionsFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

// --------
// STOCK MOVEMENTS
// --------

func (r *Repository) writeStockMovementsSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	// Тянем все движения. Если у тебя нет метода GetMovements — заведи или через прямой запрос.
	movements, total, err := r.GetMovements(ctx, GetMovementsParams{
		Page:  1,
		Limit: config.MAX_EXCEL_SHEET_ROWS,
	})
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many movements: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{
		"ID отгрузки",
		"ID позиции",
		"Тип движения",
		"Количество",
		"Комментарий",
		"Дата движения",
	}
	writeHeader(f, sheetName, headers)

	for i, m := range movements {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), m.OutcomeBatchID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), m.PositionID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), movementTypeRu(m.Type))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), m.Quantity)
		if m.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), *m.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), m.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	writeTotalRow(f, sheetName, len(movements)+4, "ВСЕГО ДВИЖЕНИЙ:", total)
	setColWidths(f, sheetName, len(headers), 25)

	return int(total), nil
}

func (r *Repository) GenerateStockMovementsReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeStockMovementsSheet, stockMovementsFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

// --------
// STOCK BALANCE
// --------

func (r *Repository) writeStockBalanceSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	items, err := r.GetStockBalance(ctx, nil)
	if err != nil {
		return 0, err
	}

	if len(items) > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many balance rows: %d > %d", len(items), config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{
		"Товар",
		"Всего на складе",
		"Зарезервировано",
		"Доступно",
	}
	writeHeader(f, sheetName, headers)

	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.UnitName)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Quantity)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Reserved)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.Available)
	}

	writeTotalRow(f, sheetName, len(items)+4, "ВСЕГО ПОЗИЦИЙ:", len(items))
	setColWidths(f, sheetName, len(headers), 25)

	return len(items), nil
}

func (r *Repository) GenerateStockBalanceReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeStockBalanceSheet, stockBalanceFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

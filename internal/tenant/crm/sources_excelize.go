// internal/tenant/crm/repository_sources_excel.go
package crm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const sourcesFilePrefix = "reports/kroncl_client_sources_"

func (r *Repository) writeSourcesSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	req := GetSourcesRequest{
		Page:  1,
		Limit: config.MAX_EXCEL_SHEET_ROWS,
	}

	sources, total, err := r.GetClientSources(ctx, req)
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many sources: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Название", "Тип", "Статус", "Сайт/URL", "Комментарий", "Системный", "ID источника"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, source := range sources {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), source.Name)

		sourceType := ""
		switch source.Type {
		case SourceTypeOrganic:
			sourceType = "Органический"
		case SourceTypeSocial:
			sourceType = "Соцсети"
		case SourceTypeReferral:
			sourceType = "Реферальный"
		case SourceTypePaid:
			sourceType = "Платный"
		case SourceTypeEmail:
			sourceType = "Email"
		case SourceTypeOther:
			sourceType = "Другое"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), sourceType)

		status := "Активен"
		if source.Status == SourceStatusInactive {
			status = "Неактивен"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), status)

		if source.URL != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *source.URL)
		}

		if source.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), *source.Comment)
		}

		system := "Нет"
		if source.System {
			system = "Да"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), system)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), source.ID)
	}

	summaryRow := len(sources) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО ИСТОЧНИКОВ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 20)
	}

	return int(total), nil
}

func (r *Repository) GenerateSourcesReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeSourcesSheet, sourcesFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

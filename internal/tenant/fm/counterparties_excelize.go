package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const counterpartiesFilePrefix = "reports/kroncl_counterparties_"

func (r *Repository) writeCounterpartiesSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	filters := GetCounterpartiesRequest{}
	counterparties, total, err := r.GetCounterparties(ctx, 0, config.MAX_EXCEL_SHEET_ROWS, filters)
	if err != nil {
		return 0, err
	}

	headers := []string{"Название", "Тип", "Статус", "Комментарий", "ID контрагента"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, cp := range counterparties {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), cp.Name)

		typeMap := map[CounterpartyType]string{
			CounterpartyTypeBank:         "Банк",
			CounterpartyTypeOrganization: "Организация",
			CounterpartyTypePerson:       "Физ. лицо",
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), typeMap[cp.Type])

		statusMap := map[CounterpartyStatus]string{
			CounterpartyStatusActive:   "Активен",
			CounterpartyStatusInactive: "Неактивен",
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), statusMap[cp.Status])

		if cp.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *cp.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), cp.ID)
	}

	summaryRow := len(counterparties) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО КОНТРАГЕНТОВ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	return total, nil
}

func (r *Repository) GenerateCounterpartiesReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeCounterpartiesSheet, counterpartiesFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

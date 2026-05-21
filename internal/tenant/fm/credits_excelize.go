package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const creditsFilePrefix = "reports/kroncl_credits_"

func (r *Repository) writeCreditsSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	filters := GetCreditsRequest{}
	credits, total, err := r.GetCredits(ctx, 0, config.MAX_EXCEL_SHEET_ROWS, filters)
	if err != nil {
		return 0, err
	}

	headers := []string{"Название", "Тип", "Статус", "Сумма", "Валюта", "Ставка %", "Начало", "Окончание", "Контрагент", "Комментарий", "ID кредита"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	moneyStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 2})

	for i, credit := range credits {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), credit.Name)

		typeMap := map[CreditType]string{
			CreditTypeDebt:   "Мы должны",
			CreditTypeCredit: "Нам должны",
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), typeMap[credit.Type])

		statusMap := map[CreditStatus]string{
			CreditStatusActive: "Активен",
			CreditStatusClosed: "Закрыт",
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), statusMap[credit.Status])

		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), float64(credit.TotalAmount))
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), moneyStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), credit.Currency)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), credit.InterestRate)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), credit.StartDate.Format("2006-01-02"))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), credit.EndDate.Format("2006-01-02"))

		if credit.Counterparty != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), credit.Counterparty.Name)
		}
		if credit.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), *credit.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), credit.ID)
	}

	summaryRow := len(credits) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО КРЕДИТОВ/ЗАЙМОВ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	return total, nil
}

func (r *Repository) GenerateCreditsReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeCreditsSheet, creditsFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

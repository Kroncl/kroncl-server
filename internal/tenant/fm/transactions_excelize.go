package fm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const filePrefix = "reports/kroncl_transactions_"

func (r *Repository) writeTransactionsSheet(ctx context.Context, f *excelize.File, sheetName string, startDate, endDate time.Time) (int, error) {
	filters := GetTransactionsRequest{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     config.MAX_EXCEL_SHEET_ROWS,
		Page:      1,
	}

	transactions, total, err := r.GetTransactions(ctx, 0, config.MAX_EXCEL_SHEET_ROWS, filters)
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many transactions: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Знак", "Тип", "Сумма", "Категория", "Сотрудник", "Дата", "Валюта", "Статус", "Комментарий", "ID транзакции"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	moneyStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 2})

	var totalIncome, totalExpense float64

	for i, t := range transactions {
		row := i + 2
		amount := float64(t.BaseAmount)

		if t.Direction == TransactionDirectionIncome {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "+")
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Доход")
			totalIncome += amount
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "-")
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Расход")
			totalExpense += amount
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), amount)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), moneyStyle)

		if t.CategoryName != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *t.CategoryName)
		}
		if t.EmployeeFirstName != nil {
			name := *t.EmployeeFirstName
			if t.EmployeeLastName != nil && *t.EmployeeLastName != "" {
				name = fmt.Sprintf("%s %s", name, *t.EmployeeLastName)
			}
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), name)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), t.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), t.Currency)

		statusMap := map[TransactionStatus]string{
			TransactionStatusCompleted: "Завершена",
			TransactionStatusPending:   "В обработке",
			TransactionStatusFailed:    "Ошибка",
			TransactionStatusCancelled: "Отменена",
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), statusMap[t.Status])

		if t.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), *t.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), t.ID)
	}

	summaryRow := len(transactions) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	styleTotal, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true},
		NumFmt: 2,
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"#E8F0FE"}, Pattern: 1},
	})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ИТОГО ДОХОДЫ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), totalIncome)
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", summaryRow), fmt.Sprintf("C%d", summaryRow), styleTotal)

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow+1), fmt.Sprintf("B%d", summaryRow+1))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow+1), "ИТОГО РАСХОДЫ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow+1), fmt.Sprintf("A%d", summaryRow+1), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow+1), totalExpense)
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", summaryRow+1), fmt.Sprintf("C%d", summaryRow+1), styleTotal)

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow+2), fmt.Sprintf("B%d", summaryRow+2))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow+2), "БАЛАНС:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow+2), fmt.Sprintf("A%d", summaryRow+2), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow+2), totalIncome-totalExpense)
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", summaryRow+2), fmt.Sprintf("C%d", summaryRow+2), styleTotal)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 20)
	}

	return int(total), nil
}

func (r *Repository) GenerateTransactionsReport(ctx context.Context, startDate, endDate time.Time) (string, int, error) {
	generator := func(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
		return r.writeTransactionsSheet(ctx, f, sheetName, startDate, endDate)
	}

	result, err := r.excelizer.GenerateSingleSheetReport(ctx, generator, filePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

package fm

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

const maxExcelRows = 10000
const filePrefix string = "reports/kroncl_transactions_"

func (r *Repository) GenerateTransactionsReport(ctx context.Context, startDate, endDate time.Time) (string, int, error) {
	filters := GetTransactionsRequest{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     maxExcelRows,
		Page:      1,
	}

	transactions, total, err := r.GetTransactions(ctx, 0, maxExcelRows, filters)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	if total > maxExcelRows {
		return "", int(total), fmt.Errorf("too many transactions: %d > %d", total, maxExcelRows)
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Финансовые операции"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Знак", "Тип", "Сумма", "Категория", "Сотрудник", "Дата", "Валюта", "Статус", "Комментарий", "ID транзакции"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)

		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		f.SetCellStyle(sheet, cell, cell, style)
	}

	moneyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 2,
	})

	var totalIncome float64
	var totalExpense float64

	for i, t := range transactions {
		row := i + 2
		amount := float64(t.BaseAmount)

		var sign string
		var typeName string
		if t.Direction == TransactionDirectionIncome {
			sign = "+"
			typeName = "Доход"
			totalIncome += amount
		} else {
			sign = "-"
			typeName = "Расход"
			totalExpense += amount
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sign)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), typeName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), amount)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), moneyStyle)

		categoryName := ""
		if t.CategoryName != nil {
			categoryName = *t.CategoryName
		}
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), categoryName)

		employeeName := ""
		if t.EmployeeFirstName != nil {
			employeeName = *t.EmployeeFirstName
			if t.EmployeeLastName != nil && *t.EmployeeLastName != "" {
				employeeName = fmt.Sprintf("%s %s", *t.EmployeeFirstName, *t.EmployeeLastName)
			}
		}
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), employeeName)

		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), t.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), t.Currency)

		status := ""
		switch t.Status {
		case TransactionStatusCompleted:
			status = "Завершена"
		case TransactionStatusPending:
			status = "В обработке"
		case TransactionStatusFailed:
			status = "Ошибка"
		case TransactionStatusCancelled:
			status = "Отменена"
		}
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), status)

		if t.Comment != nil {
			f.SetCellValue(sheet, fmt.Sprintf("I%d", row), *t.Comment)
		}

		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), t.ID)
	}

	totalRow := len(transactions) + 4

	styleBold, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	styleTotal, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true},
		NumFmt: 2, // 0.00
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E8F0FE"},
			Pattern: 1,
		},
	})

	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("B%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "ИТОГО ДОХОДЫ:")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("A%d", totalRow), styleBold)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", totalRow), totalIncome)
	f.SetCellStyle(sheet, fmt.Sprintf("C%d", totalRow), fmt.Sprintf("C%d", totalRow), styleTotal)

	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow+1), fmt.Sprintf("B%d", totalRow+1))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow+1), "ИТОГО РАСХОДЫ:")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow+1), fmt.Sprintf("A%d", totalRow+1), styleBold)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", totalRow+1), totalExpense)
	f.SetCellStyle(sheet, fmt.Sprintf("C%d", totalRow+1), fmt.Sprintf("C%d", totalRow+1), styleTotal)

	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow+2), fmt.Sprintf("B%d", totalRow+2))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow+2), "БАЛАНС:")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow+2), fmt.Sprintf("A%d", totalRow+2), styleBold)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", totalRow+2), totalIncome-totalExpense)
	f.SetCellStyle(sheet, fmt.Sprintf("C%d", totalRow+2), fmt.Sprintf("C%d", totalRow+2), styleTotal)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheet, colLetter, colLetter, 20)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", 0, fmt.Errorf("failed to write excel: %w", err)
	}

	objectPath := fmt.Sprintf("%s%s.xlsx", filePrefix, time.Now().Format("20060102_150405"))

	err = r.mediaService.UploadFileToBucket(ctx, objectPath, &buf, int64(buf.Len()), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload report: %w", err)
	}

	return objectPath, int(total), nil
}

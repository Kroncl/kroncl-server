package hrm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const employeesFilePrefix = "reports/kroncl_employees_"

func (r *Repository) writeEmployeesSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	employees, total, err := r.GetEmployees(ctx, 0, config.MAX_EXCEL_SHEET_ROWS, "")
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many employees: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Имя", "Фамилия", "Email", "Телефон", "Статус", "Аккаунт привязан", "ID сотрудника"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, emp := range employees {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), emp.FirstName)

		if emp.LastName != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), *emp.LastName)
		}

		if emp.Email != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), *emp.Email)
		}

		if emp.Phone != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *emp.Phone)
		}

		status := "Активен"
		if emp.Status == EmployeeStatusInactive {
			status = "Неактивен"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), status)

		linked := "Нет"
		if emp.IsAccountLinked {
			linked = "Да"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), linked)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), emp.ID)
	}

	summaryRow := len(employees) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО СОТРУДНИКОВ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 20)
	}

	return int(total), nil
}

func (r *Repository) GenerateEmployeesReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeEmployeesSheet, employeesFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

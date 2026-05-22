package crm

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"time"

	"github.com/xuri/excelize/v2"
)

const clientsFilePrefix = "reports/kroncl_clients_"

func (r *Repository) writeClientsSheet(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
	req := GetClientsRequest{
		Page:  1,
		Limit: config.MAX_EXCEL_SHEET_ROWS,
	}

	clients, total, err := r.GetClients(ctx, req)
	if err != nil {
		return 0, err
	}

	if total > config.MAX_EXCEL_SHEET_ROWS {
		return 0, fmt.Errorf("too many clients: %d > %d", total, config.MAX_EXCEL_SHEET_ROWS)
	}

	headers := []string{"Имя", "Фамилия", "Отчество", "Телефон", "Email", "Тип", "Статус", "Источник", "Комментарий", "ID клиента"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	for i, client := range clients {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), client.FirstName)

		if client.LastName != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), *client.LastName)
		}

		if client.Patronymic != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), *client.Patronymic)
		}

		if client.Phone != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *client.Phone)
		}

		if client.Email != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), *client.Email)
		}

		clientType := "Физ. лицо"
		if client.Type == ClientTypeLegal {
			clientType = "Юр. лицо"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), clientType)

		status := "Активен"
		if client.Status == ClientStatusInactive {
			status = "Неактивен"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), status)

		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), client.Source.Name)

		if client.Comment != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), *client.Comment)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), client.ID)
	}

	summaryRow := len(clients) + 4
	styleBold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "ВСЕГО КЛИЕНТОВ:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), styleBold)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), total)

	for col := 1; col <= len(headers); col++ {
		colLetter, _ := excelize.CoordinatesToCellName(col, 1)
		f.SetColWidth(sheetName, colLetter, colLetter, 20)
	}

	return int(total), nil
}

func (r *Repository) GenerateClientsReport(ctx context.Context) (string, int, error) {
	result, err := r.excelizer.GenerateSingleSheetReport(ctx, r.writeClientsSheet, clientsFilePrefix, 1*time.Hour)
	if err != nil {
		return "", 0, err
	}
	return result.ObjectPath, result.TotalRows, nil
}

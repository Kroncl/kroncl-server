package fm

import (
	"context"
	"fmt"
	"time"

	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/excelizer"

	"github.com/xuri/excelize/v2"
)

const (
	ReportTypeTransactions = "transactions"
	ReportTypeCategories   = "categories"
	ReportTypeCredits      = "credits"
)

type FullReportOptions struct {
	Types     []string
	StartDate *time.Time
	EndDate   *time.Time
	Comment   *string
}

func (r *Repository) GenerateFullReport(ctx context.Context, opts FullReportOptions) (*docs.Doc, error) {
	generators := make(map[string]excelizer.SheetGenerator)

	startDate := opts.StartDate
	endDate := opts.EndDate

	if startDate == nil || endDate == nil {
		now := time.Now()
		startDate = &now
		endDate = &now
		startDate = func() *time.Time { t := now.AddDate(0, -1, 0); return &t }()
	}

	for _, t := range opts.Types {
		switch t {
		case ReportTypeTransactions:
			generators["Операции"] = func(ctx context.Context, f *excelize.File, sheetName string) (int, error) {
				return r.writeTransactionsSheet(ctx, f, sheetName, *startDate, *endDate)
			}
		case ReportTypeCategories:
			generators["Категории"] = r.writeCategoriesSheet
		case ReportTypeCredits:
			generators["Кредиты и займы"] = r.writeCreditsSheet
		default:
			return nil, fmt.Errorf("unknown report type: %s", t)
		}
	}

	if len(generators) == 0 {
		return nil, fmt.Errorf("no valid report types provided")
	}

	result, err := r.excelizer.GenerateMultiSheetReport(ctx, generators, "reports/kroncl_full_report_", 1*time.Hour)
	if err != nil {
		return nil, err
	}

	module := "fm"
	docType := "full"

	doc, err := r.docsService.CreateDoc(ctx, docs.CreateDocRequest{
		ObjectPath: result.ObjectPath,
		Module:     &module,
		Type:       &docType,
		Comment:    opts.Comment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	return doc, nil
}

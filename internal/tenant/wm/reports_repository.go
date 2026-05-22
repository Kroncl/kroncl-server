package wm

import (
	"context"
	"fmt"
	"time"

	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/excelizer"
)

const (
	ReportTypeCatalogCategories = "catalog_categories"
	ReportTypeCatalogUnits      = "catalog_units"
	ReportTypeStockBalance      = "stock_balance"
	ReportTypeStockBatches      = "stock_batches"
	ReportTypeStockPositions    = "stock_positions"
)

type FullReportOptions struct {
	Types   []string
	Comment *string
}

func (r *Repository) GenerateFullReport(ctx context.Context, opts FullReportOptions) (*docs.Doc, error) {
	generators := make(map[string]excelizer.SheetGenerator)

	for _, t := range opts.Types {
		switch t {
		case ReportTypeCatalogCategories:
			generators["Категории каталога"] = r.writeCatalogCategoriesSheet
		case ReportTypeCatalogUnits:
			generators["Товары и услуги"] = r.writeCatalogUnitsSheet
		case ReportTypeStockBalance:
			generators["Остатки на складе"] = r.writeStockBalanceSheet
		case ReportTypeStockBatches:
			generators["Партии (поставки)"] = r.writeStockBatchesSheet
		case ReportTypeStockPositions:
			generators["Складские позиции"] = r.writeStockPositionsSheet
		default:
			return nil, fmt.Errorf("unknown report type: %s", t)
		}
	}

	if len(generators) == 0 {
		return nil, fmt.Errorf("no valid report types provided")
	}

	result, err := r.excelizer.GenerateMultiSheetReport(ctx, generators, "reports/kroncl_wm_full_report_", 1*time.Hour)
	if err != nil {
		return nil, err
	}

	module := "wm"
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

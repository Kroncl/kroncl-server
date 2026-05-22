package crm

import (
	"context"
	"fmt"
	"time"

	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/excelizer"
)

const (
	ReportTypeClients = "clients"
	ReportTypeSources = "sources"
)

type FullReportOptions struct {
	Types   []string
	Comment *string
}

func (r *Repository) GenerateFullReport(ctx context.Context, opts FullReportOptions) (*docs.Doc, error) {
	generators := make(map[string]excelizer.SheetGenerator)

	for _, t := range opts.Types {
		switch t {
		case ReportTypeClients:
			generators["Клиенты"] = r.writeClientsSheet
		case ReportTypeSources:
			generators["Источники"] = r.writeSourcesSheet
		default:
			return nil, fmt.Errorf("unknown report type: %s", t)
		}
	}

	if len(generators) == 0 {
		return nil, fmt.Errorf("no valid report types provided")
	}

	result, err := r.excelizer.GenerateMultiSheetReport(ctx, generators, "reports/kroncl_crm_full_report_", 1*time.Hour)
	if err != nil {
		return nil, err
	}

	module := "crm"
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

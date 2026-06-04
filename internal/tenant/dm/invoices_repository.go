package dm

import (
	"context"
	"fmt"
	"time"

	"kroncl-server/internal/config"
	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/pdfgen"
)

func (r *Repository) GenerateDealInvoice(ctx context.Context, id string, req GenerateInvoiceRequest) (*docs.Doc, error) {
	// Валидация
	if len(req.Positions) == 0 {
		return nil, fmt.Errorf("at least one position is required")
	}

	if req.TotalAmount <= 0 {
		return nil, fmt.Errorf("total_amount must be greater than 0")
	}

	exists, err := r.DealExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check deal existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("deal not found: %s", id)
	}

	invoiceData := InvoiceData{
		DealID:          req.DealID,
		LegalName:       req.LegalName,
		Inn:             req.Inn,
		Ogrn:            req.Ogrn,
		BankName:        req.BankName,
		WarrantyTerms:   req.WarrantyTerms,
		AdditionalTerms: req.AdditionalTerms,
		Positions:       req.Positions,
		TotalAmount:     req.TotalAmount,
		CreatedAt:       time.Now(),
		Comment:         req.Comment,
	}

	pdfBuf, err := r.pdfgen.GenerateFromTemplate(ctx, pdfgen.GenerateFromTemplateRequest{
		TemplatePath: config.DM_TEMPLATE_INVOICE,
		Data:         invoiceData,
		Options: &pdfgen.GeneratePDFOptions{
			PaperSize: "A4",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	objectPath := fmt.Sprintf("invoices/kroncl_deal_%s_%s.pdf", id, time.Now().Format("20060102_150405"))

	err = r.mediaService.UploadBufferToBucket(ctx, objectPath, pdfBuf, "application/pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to upload invoice: %w", err)
	}

	module := "dm"
	docType := "invoice"

	doc, err := r.docsService.CreateDoc(ctx, docs.CreateDocRequest{
		ObjectPath: objectPath,
		Module:     &module,
		Type:       &docType,
		Comment:    req.Comment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	return doc, nil
}

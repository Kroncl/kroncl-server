package dm

import (
	"context"
	"fmt"
	"time"

	"kroncl-server/internal/config"
	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/pdfgen"
)

// GenerateDealInvoice генерирует накладную для сделки
func (r *Repository) GenerateDealInvoice(ctx context.Context, dealID string, comment *string) (*docs.Doc, error) {
	deal, err := r.GetDealWithDetails(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Подсчитываем общую сумму
	var totalAmount float64
	for _, pos := range deal.Positions {
		totalAmount += pos.Price * pos.Quantity
	}

	// Передаём данные вместе с суммой
	invoiceData := struct {
		*DealWithPositions
		TotalAmount float64
	}{
		DealWithPositions: deal,
		TotalAmount:       totalAmount,
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

	// Сохраняем PDF в MinIO
	objectPath := fmt.Sprintf("invoices/kroncl_deal_%s_%s.pdf", dealID, time.Now().Format("20060102_150405"))
	err = r.mediaService.UploadBufferToBucket(ctx, objectPath, pdfBuf, "application/pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to upload invoice: %w", err)
	}

	// Сохраняем запись в docs
	module := "dm"
	docType := "invoice"

	doc, err := r.docsService.CreateDoc(ctx, docs.CreateDocRequest{
		ObjectPath: objectPath,
		Module:     &module,
		Type:       &docType,
		Comment:    comment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	return doc, nil
}

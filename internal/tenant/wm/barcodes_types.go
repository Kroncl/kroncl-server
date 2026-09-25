package wm

import "time"

// Barcode represents a barcode dictionary entry
type Barcode struct {
	ID            string                 `json:"id"`
	Barcode       string                 `json:"barcode"`
	Maker         *string                `json:"maker"`
	CatalogUnitID *string                `json:"catalog_unit_id"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CreateBarcodeRequest represents request to create/update barcode
type CreateBarcodeRequest struct {
	Barcode       string                 `json:"barcode" validate:"required,min=1,max=255"`
	Maker         *string                `json:"maker,omitempty"`
	CatalogUnitID *string                `json:"catalog_unit_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateBarcodeRequest struct {
	Barcode       *string                 `json:"barcode,omitempty"`
	Maker         *string                 `json:"maker,omitempty"`
	CatalogUnitID *string                 `json:"catalog_unit_id,omitempty"`
	Metadata      *map[string]interface{} `json:"metadata,omitempty"`
}

// GetBarcodesParams represents request params for listing barcodes
type GetBarcodesParams struct {
	Page          int     `json:"page" validate:"omitempty,min=1"`
	Limit         int     `json:"limit" validate:"omitempty,min=1,max=100"`
	CatalogUnitID *string `json:"catalog_unit_id,omitempty"`
	Search        *string `json:"search,omitempty"`
}

// BarcodesResponse represents paginated response for barcodes
type BarcodesResponse struct {
	Barcodes []Barcode `json:"barcodes"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
	Pages    int       `json:"pages"`
}

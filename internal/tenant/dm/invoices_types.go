package dm

import "time"

type InvoicePosition struct {
	Name     string  `json:"name" validate:"required"`
	Quantity float64 `json:"quantity" validate:"required,gt=0"`
	Price    float64 `json:"price" validate:"required,gt=0"`
}

type GenerateInvoiceRequest struct {
	DealID          string            `json:"deal_id" validate:"required"`
	LegalName       *string           `json:"legal_name,omitempty"`
	Inn             *string           `json:"inn,omitempty"`
	Ogrn            *string           `json:"ogrn,omitempty"`
	BankName        *string           `json:"bank_name,omitempty"`
	WarrantyTerms   *string           `json:"warranty_terms,omitempty"`
	AdditionalTerms *string           `json:"additional_terms,omitempty"`
	Positions       []InvoicePosition `json:"positions" validate:"required,min=1"`
	TotalAmount     float64           `json:"total_amount" validate:"required,gt=0"`
	Comment         *string           `json:"comment,omitempty"`
}

type InvoiceData struct {
	DealID          string
	LegalName       *string
	Inn             *string
	Ogrn            *string
	BankName        *string
	WarrantyTerms   *string
	AdditionalTerms *string
	Positions       []InvoicePosition
	TotalAmount     float64
	CreatedAt       time.Time
	Comment         *string
}

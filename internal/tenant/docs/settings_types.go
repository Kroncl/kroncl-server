package docs

import "time"

type DocsSettings struct {
	LegalName       *string   `json:"legal_name"`
	LegalAddress    *string   `json:"legal_address"`
	Inn             *string   `json:"inn"`
	Ogrn            *string   `json:"ogrn"`
	BankName        *string   `json:"bank_name"`
	BankBic         *string   `json:"bank_bic"`
	BankAccount     *string   `json:"bank_account"`
	DirectorName    *string   `json:"director_name"`
	AccountantName  *string   `json:"accountant_name"`
	WarrantyTerms   *string   `json:"warranty_terms"`
	AdditionalTerms *string   `json:"additional_termss"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateDocsSettingsRequest struct {
	LegalName       *string `json:"legal_name,omitempty"`
	LegalAddress    *string `json:"legal_address,omitempty"`
	Inn             *string `json:"inn,omitempty"`
	Ogrn            *string `json:"ogrn,omitempty"`
	BankName        *string `json:"bank_name,omitempty"`
	BankBic         *string `json:"bank_bic,omitempty"`
	BankAccount     *string `json:"bank_account,omitempty"`
	DirectorName    *string `json:"director_name,omitempty"`
	AccountantName  *string `json:"accountant_name,omitempty"`
	WarrantyTerms   *string `json:"warranty_terms,omitempty"`
	AdditionalTerms *string `json:"additional_terms,omitempty"`
}

package cpm

import "time"

// -------
// COUNTERPARTIES
// -------

type CounterpartyType string

const (
	CounterpartyTypeBank         CounterpartyType = "bank"
	CounterpartyTypeOrganization CounterpartyType = "organization"
	CounterpartyTypePerson       CounterpartyType = "person"
)

type CounterpartyStatus string

const (
	CounterpartyStatusActive   CounterpartyStatus = "active"
	CounterpartyStatusInactive CounterpartyStatus = "inactive"
)

type Counterparty struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Comment         *string                `json:"comment"`
	Type            CounterpartyType       `json:"type"`
	Status          CounterpartyStatus     `json:"status"`
	INN             *string                `json:"inn"`
	OGRN            *string                `json:"ogrn"`
	KPP             *string                `json:"kpp"`
	Address         *string                `json:"address"`
	DefaultCurrency *string                `json:"default_currency"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type CreateCounterpartyRequest struct {
	Name            string                 `json:"name" validate:"required,min=1,max=255"`
	Comment         string                 `json:"comment,omitempty" validate:"omitempty,max=1000"`
	Type            CounterpartyType       `json:"type" validate:"required,oneof=bank organization person"`
	Status          CounterpartyStatus     `json:"status" validate:"omitempty,oneof=active inactive"`
	INN             string                 `json:"inn,omitempty" validate:"omitempty,max=12"`
	OGRN            string                 `json:"ogrn,omitempty" validate:"omitempty,max=15"`
	KPP             string                 `json:"kpp,omitempty" validate:"omitempty,max=9"`
	Address         string                 `json:"address,omitempty" validate:"omitempty,max=500"`
	DefaultCurrency string                 `json:"default_currency,omitempty" validate:"omitempty,max=10"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateCounterpartyRequest struct {
	Name            *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment         *string                 `json:"comment,omitempty" validate:"omitempty,max=1000"`
	Type            *CounterpartyType       `json:"type,omitempty" validate:"omitempty,oneof=bank organization person"`
	INN             *string                 `json:"inn,omitempty" validate:"omitempty,max=12"`
	OGRN            *string                 `json:"ogrn,omitempty" validate:"omitempty,max=15"`
	KPP             *string                 `json:"kpp,omitempty" validate:"omitempty,max=9"`
	Address         *string                 `json:"address,omitempty" validate:"omitempty,max=500"`
	DefaultCurrency *string                 `json:"default_currency,omitempty" validate:"omitempty,max=10"`
	Metadata        *map[string]interface{} `json:"metadata,omitempty"`
}

type GetCounterpartiesRequest struct {
	Page   int                 `json:"page" validate:"omitempty,min=1"`
	Limit  int                 `json:"limit" validate:"omitempty,min=1,max=100"`
	Type   *CounterpartyType   `json:"type,omitempty"`
	Status *CounterpartyStatus `json:"status,omitempty"`
	Search *string             `json:"search,omitempty"`
}

package wm

import "time"

// --------
// CATEGORIES
// --------

// CategoryStatus represents the status of a category
type CategoryStatus string

const (
	CategoryStatusActive   CategoryStatus = "active"
	CategoryStatusInactive CategoryStatus = "inactive"
)

// CatalogCategory represents a product/service category
type CatalogCategory struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Comment   *string                `json:"comment"`
	Status    CategoryStatus         `json:"status"`
	ParentID  *string                `json:"parent_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// CreateCategoryRequest represents request to create a category
type CreateCategoryRequest struct {
	Name     string                 `json:"name" validate:"required,min=1,max=255"`
	Comment  *string                `json:"comment,omitempty"`
	ParentID *string                `json:"parent_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCategoryRequest represents request to update a category
type UpdateCategoryRequest struct {
	Name     *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment  *string                 `json:"comment,omitempty"`
	ParentID *string                 `json:"parent_id,omitempty"`
	Status   *CategoryStatus         `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
}

// GetCategoriesRequest represents request params for listing categories
type GetCategoriesRequest struct {
	Page     int             `json:"page" validate:"omitempty,min=1"`
	Limit    int             `json:"limit" validate:"omitempty,min=1,max=100"`
	Status   *CategoryStatus `json:"status,omitempty"`
	ParentID *string         `json:"parent_id,omitempty"`
	Search   *string         `json:"search,omitempty"`
}

// CategoriesResponse represents paginated response
type CategoriesResponse struct {
	Categories []CatalogCategory `json:"categories"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Pages      int               `json:"pages"`
}

// --------
// UNITS
// --------

// UnitType represents type of catalog unit
type UnitType string

const (
	UnitTypeProduct UnitType = "product"
	UnitTypeService UnitType = "service"
)

// UnitStatus represents status of catalog unit
type UnitStatus string

const (
	UnitStatusActive   UnitStatus = "active"
	UnitStatusInactive UnitStatus = "inactive"
)

// InventoryType represents inventory tracking type
type InventoryType string

const (
	InventoryTypeTracked   InventoryType = "tracked"
	InventoryTypeUntracked InventoryType = "untracked"
)

// TrackingDetail represents detailed tracking type for tracked items
type TrackingDetail string

const (
	TrackingDetailBatch  TrackingDetail = "batch"  // партионный учет (FIFO/LIFO)
	TrackingDetailSerial TrackingDetail = "serial" // поштучный учет (каждый экземпляр)
)

// TrackedType represents FIFO/LIFO for batch-tracked items
type TrackedType string

const (
	TrackedTypeFIFO TrackedType = "fifo"
	TrackedTypeLIFO TrackedType = "lifo"
)

// CurrencyType represents currency
type CurrencyType string

const (
	CurrencyRUB CurrencyType = "RUB"
	// CurrencyUSD CurrencyType = "USD"
	// CurrencyEUR CurrencyType = "EUR"
	// CurrencyKZT CurrencyType = "KZT"
)

// CatalogUnit represents a product or service
type CatalogUnit struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Comment        *string                `json:"comment"`
	Type           UnitType               `json:"type"`
	Status         UnitStatus             `json:"status"`
	InventoryType  InventoryType          `json:"inventory_type"`
	TrackingDetail *TrackingDetail        `json:"tracking_detail"` // batch/serial - только для tracked
	TrackedType    *TrackedType           `json:"tracked_type"`    // только для batch-учета
	Unit           string                 `json:"unit"`            // pcs, kg, l, etc
	SalePrice      float64                `json:"sale_price"`
	PurchasePrice  *float64               `json:"purchase_price"` // only for tracked
	Currency       CurrencyType           `json:"currency"`
	CategoryID     string                 `json:"category_id"` // обязательное поле
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreateUnitRequest represents request to create a catalog unit
type CreateUnitRequest struct {
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Comment        *string                `json:"comment,omitempty"`
	Type           UnitType               `json:"type" validate:"required,oneof=product service"`
	Status         *UnitStatus            `json:"status,omitempty"` // defaults to active
	InventoryType  InventoryType          `json:"inventory_type" validate:"required,oneof=tracked untracked"`
	TrackingDetail *TrackingDetail        `json:"tracking_detail,omitempty" validate:"omitempty,oneof=batch serial"`
	TrackedType    *TrackedType           `json:"tracked_type,omitempty" validate:"omitempty,oneof=fifo lifo"`
	Unit           string                 `json:"unit" validate:"required"`
	SalePrice      float64                `json:"sale_price" validate:"required,min=0"`
	PurchasePrice  *float64               `json:"purchase_price,omitempty" validate:"omitempty,min=0"`
	Currency       CurrencyType           `json:"currency" validate:"required,oneof=RUB"`
	CategoryID     string                 `json:"category_id" validate:"required"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUnitRequest represents request to update a catalog unit
type UpdateUnitRequest struct {
	Name           *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment        *string                 `json:"comment,omitempty"`
	Type           *UnitType               `json:"type,omitempty" validate:"omitempty,oneof=product service"`
	Status         *UnitStatus             `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	InventoryType  *InventoryType          `json:"inventory_type,omitempty" validate:"omitempty,oneof=tracked untracked"`
	TrackingDetail *TrackingDetail         `json:"tracking_detail,omitempty" validate:"omitempty,oneof=batch serial"`
	TrackedType    *TrackedType            `json:"tracked_type,omitempty" validate:"omitempty,oneof=fifo lifo"`
	Unit           *string                 `json:"unit,omitempty"`
	SalePrice      *float64                `json:"sale_price,omitempty" validate:"omitempty,min=0"`
	PurchasePrice  *float64                `json:"purchase_price,omitempty" validate:"omitempty,min=0"`
	Currency       *CurrencyType           `json:"currency,omitempty" validate:"omitempty,oneof=RUB"`
	CategoryID     *string                 `json:"category_id,omitempty" validate:"omitempty"`
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
}

// GetUnitsRequest represents request params for listing catalog units
type GetUnitsRequest struct {
	Page           int             `json:"page" validate:"omitempty,min=1"`
	Limit          int             `json:"limit" validate:"omitempty,min=1,max=100"`
	Type           *UnitType       `json:"type,omitempty"`
	Status         *UnitStatus     `json:"status,omitempty"`
	InventoryType  *InventoryType  `json:"inventory_type,omitempty"`
	TrackingDetail *TrackingDetail `json:"tracking_detail,omitempty"`
	CategoryID     *string         `json:"category_id,omitempty"`
	Search         *string         `json:"search,omitempty"`
}

// UnitsResponse represents paginated response
type UnitsResponse struct {
	Units []CatalogUnit `json:"units"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Pages int           `json:"pages"`
}

package wm

import (
	"time"
)

// --------
// STOCK TYPES
// --------

// StockDirection represents direction of stock movement
type StockDirection string

const (
	StockDirectionIncome  StockDirection = "income"  // приход на склад
	StockDirectionOutcome StockDirection = "outcome" // расход со склада
)

// StockPositionType represents type of stock position
type StockPositionType string

const (
	StockPositionTypeBatch  StockPositionType = "batch"  // партионный учет (количество)
	StockPositionTypeSerial StockPositionType = "serial" // поштучный учет (каждый экземпляр)
)

// StockBatch represents a stock movement document (income/outcome)
type StockBatch struct {
	ID        string                 `json:"id"`
	Direction StockDirection         `json:"direction"`
	Comment   *string                `json:"comment"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// StockPosition represents a physical stock position (batch or serial)
type StockPosition struct {
	ID        string            `json:"id"`
	Type      StockPositionType `json:"type"`     // batch or serial
	UnitID    string            `json:"unit_id"`  // ссылка на catalog_units
	Quantity  float64           `json:"quantity"` // для batch > 0, для serial = 1
	CreatedAt time.Time         `json:"created_at"`
}

// StockBatchPosition represents a position in batch creation request
type StockBatchPosition struct {
	UnitID   string  `json:"unit_id" validate:"required"`
	Quantity float64 `json:"quantity" validate:"required,min=0.001"`
	Price    float64 `json:"price" validate:"required,min=0"` // purchase_price для income, sale_price для outcome
}

// CreateStockBatchRequest represents request to create a stock batch with positions
type CreateStockBatchRequest struct {
	Direction StockDirection         `json:"direction" validate:"required,oneof=income outcome"`
	Comment   *string                `json:"comment,omitempty"`
	Positions []StockBatchPosition   `json:"positions" validate:"required,min=1"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type CreateStockBatchOnlyRequest struct {
	Direction StockDirection         `json:"direction" validate:"required,oneof=income outcome"`
	Comment   *string                `json:"comment,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BatchWithPositionsResponse represents batch with its positions (монолитно)
type BatchWithPositionsResponse struct {
	ID        string                     `json:"id"`
	Direction StockDirection             `json:"direction"`
	Comment   *string                    `json:"comment"`
	Metadata  map[string]interface{}     `json:"metadata"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
	Positions []PositionWithUnitResponse `json:"positions"`
}

// PositionWithUnitResponse represents position with unit info (монолитно)
type PositionWithUnitResponse struct {
	ID        string            `json:"id"`
	Type      StockPositionType `json:"type"`
	UnitID    string            `json:"unit_id"`
	Quantity  float64           `json:"quantity"`
	CreatedAt time.Time         `json:"created_at"`
	BatchID   string            `json:"batch_id"`
	Unit      CatalogUnit       `json:"unit"`
}

// CreateStockBatchResponse represents response after creating batch with positions
type CreateStockBatchResponse struct {
	BatchID   string                     `json:"batch_id"`
	Direction StockDirection             `json:"direction"`
	Comment   *string                    `json:"comment"`
	Metadata  map[string]interface{}     `json:"metadata"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
	Positions []PositionWithUnitResponse `json:"positions"`
}

// GetStockBatchesParams represents request params for listing stock batches
type GetStockBatchesParams struct {
	Page      int             `json:"page" validate:"omitempty,min=1"`
	Limit     int             `json:"limit" validate:"omitempty,min=1,max=100"`
	Direction *StockDirection `json:"direction,omitempty"`
	UnitID    *string         `json:"unit_id,omitempty"`
	Search    *string         `json:"search,omitempty"`
}

// GetStockPositionsParams represents request params for listing stock positions
type GetStockPositionsParams struct {
	Page    int                `json:"page" validate:"omitempty,min=1"`
	Limit   int                `json:"limit" validate:"omitempty,min=1,max=100"`
	Type    *StockPositionType `json:"type,omitempty"`
	UnitID  *string            `json:"unit_id,omitempty"`
	BatchID *string            `json:"batch_id,omitempty"`
	InStock *bool              `json:"in_stock,omitempty"`
}

// StockPositionsResponse represents paginated response for positions
type StockPositionsResponse struct {
	Positions []PositionWithUnitResponse `json:"positions"`
	Total     int64                      `json:"total"`
	Page      int                        `json:"page"`
	Limit     int                        `json:"limit"`
	Pages     int                        `json:"pages"`
}

// StockBatchesResponse represents paginated response for batches
type StockBatchesResponse struct {
	Batches []StockBatch `json:"batches"`
	Total   int64        `json:"total"`
	Page    int          `json:"page"`
	Limit   int          `json:"limit"`
	Pages   int          `json:"pages"`
}

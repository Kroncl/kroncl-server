package wm

import (
	"time"
)

// --------
// STOCK ENUMS
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
	StockPositionTypeBatch  StockPositionType = "batch"  // партионный учёт
	StockPositionTypeSerial StockPositionType = "serial" // поштучный учёт
)

// StockBatchStatus represents document status
type StockBatchStatus string

const (
	StockBatchStatusDraft     StockBatchStatus = "draft"     // черновик
	StockBatchStatusLabeled   StockBatchStatus = "labeled"   // этикетки напечатаны (income)
	StockBatchStatusConfirmed StockBatchStatus = "confirmed" // проведено
	StockBatchStatusCancelled StockBatchStatus = "cancelled" // отменено
)

// StockMovementType represents type of position movement
type StockMovementType string

const (
	StockMovementTypeWriteOff   StockMovementType = "write_off"  // списание при отгрузке
	StockMovementTypeReturn     StockMovementType = "return"     // возврат
	StockMovementTypeTransfer   StockMovementType = "transfer"   // перемещение
	StockMovementTypeAdjustment StockMovementType = "adjustment" // корректировка
)

// --------
// STOCK MODELS
// --------

type StockBatch struct {
	ID        string                     `json:"id"`
	Direction StockDirection             `json:"direction"`
	Status    StockBatchStatus           `json:"status"`
	Comment   *string                    `json:"comment"`
	Metadata  map[string]interface{}     `json:"metadata"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
	Positions []PositionWithUnitResponse `json:"positions"`
}

// StockPosition represents a physical stock position (batch or serial)
type StockPosition struct {
	ID            string            `json:"id"`
	Type          StockPositionType `json:"type"`
	IncomeBatchID string            `json:"income_batch_id"`
	UnitID        string            `json:"unit_id"`
	Quantity      float64           `json:"quantity"`
	UnitPrice     float64           `json:"unit_price"`
	Maker         *string           `json:"maker"`
	BarcodeID     *string           `json:"barcode_id"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// StockPositionMovement represents a movement of stock position
type StockPositionMovement struct {
	OutcomeBatchID string                 `json:"outcome_batch_id"`
	PositionID     string                 `json:"position_id"`
	Type           StockMovementType      `json:"type"`
	Quantity       float64                `json:"quantity"`
	Comment        *string                `json:"comment"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
}

// --------
// REQUESTS
// --------

// StockBatchPosition represents a position in batch creation request
type StockBatchPosition struct {
	UnitID    string  `json:"unit_id" validate:"required"`
	Quantity  float64 `json:"quantity" validate:"required,min=0.001"`
	UnitPrice float64 `json:"unit_price" validate:"required,min=0"`
	Maker     *string `json:"maker,omitempty"`
	Barcode   *string `json:"barcode,omitempty"`
}

// CreateStockBatchRequest represents request to create a stock batch with positions
type CreateStockBatchRequest struct {
	Direction StockDirection         `json:"direction" validate:"required,oneof=income outcome"`
	Comment   *string                `json:"comment,omitempty"`
	Positions []StockBatchPosition   `json:"positions" validate:"required,min=1"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CreateStockBatchOnlyRequest represents request to create an empty stock batch
type CreateStockBatchOnlyRequest struct {
	Direction StockDirection         `json:"direction" validate:"required,oneof=income outcome"`
	Comment   *string                `json:"comment,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateStockBatchStatusRequest represents request to update batch status
type UpdateStockBatchStatusRequest struct {
	Status StockBatchStatus `json:"status" validate:"required,oneof=draft labeled confirmed cancelled"`
}

// CreateStockMovementRequest represents request to create a movement (write-off)
type CreateStockMovementRequest struct {
	OutcomeBatchID string                 `json:"outcome_batch_id" validate:"required"`
	PositionID     string                 `json:"position_id" validate:"required"`
	Type           StockMovementType      `json:"type" validate:"required,oneof=write_off return transfer adjustment"`
	Quantity       float64                `json:"quantity" validate:"required,min=0.001"`
	Comment        *string                `json:"comment,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// --------
// RESPONSES
// --------

// PositionWithUnitResponse represents position with unit info
type PositionWithUnitResponse struct {
	ID            string            `json:"id"`
	Type          StockPositionType `json:"type"`
	IncomeBatchID string            `json:"income_batch_id"`
	UnitID        string            `json:"unit_id"`
	Quantity      float64           `json:"quantity"`
	UnitPrice     float64           `json:"unit_price"`
	Maker         *string           `json:"maker"`
	BarcodeID     *string           `json:"barcode_id"`
	Remaining     float64           `json:"remaining"` // остаток = quantity - сумма движений
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Unit          CatalogUnit       `json:"unit"`
}

// -----------
// STOCK BALANCE
// -----------

type StockBalanceItem struct {
	UnitID    string  `json:"unit_id"`
	UnitName  string  `json:"unit_name"`
	Quantity  float64 `json:"quantity"`
	Reserved  float64 `json:"reserved"`
	Available float64 `json:"available"`
}

// --------
// FILTERS
// --------

// GetStockBatchesParams represents request params for listing stock batches
type GetStockBatchesParams struct {
	Page      int               `json:"page" validate:"omitempty,min=1"`
	Limit     int               `json:"limit" validate:"omitempty,min=1,max=100"`
	Direction *StockDirection   `json:"direction,omitempty"`
	Status    *StockBatchStatus `json:"status,omitempty"`
	UnitID    *string           `json:"unit_id,omitempty"`
	Search    *string           `json:"search,omitempty"`
}

// GetStockPositionsParams represents request params for listing stock positions
type GetStockPositionsParams struct {
	Page          int                `json:"page" validate:"omitempty,min=1"`
	Limit         int                `json:"limit" validate:"omitempty,min=1,max=100"`
	Type          *StockPositionType `json:"type,omitempty"`
	UnitID        *string            `json:"unit_id,omitempty"`
	IncomeBatchID *string            `json:"income_batch_id,omitempty"`
	InStock       *bool              `json:"in_stock,omitempty"`
	Search        *string            `json:"search,omitempty"`
}

type GetMovementsParams struct {
	Page           int                `json:"page" validate:"omitempty,min=1"`
	Limit          int                `json:"limit" validate:"omitempty,min=1,max=1000"`
	Type           *StockMovementType `json:"type,omitempty"`
	OutcomeBatchID *string            `json:"outcome_batch_id,omitempty"`
	PositionID     *string            `json:"position_id,omitempty"`
}

// --------
// PAGINATED RESPONSES
// --------

// StockBatchesResponse represents paginated response for batches
type StockBatchesResponse struct {
	Batches []StockBatch `json:"batches"`
	Total   int64        `json:"total"`
	Page    int          `json:"page"`
	Limit   int          `json:"limit"`
	Pages   int          `json:"pages"`
}

// StockPositionsResponse represents paginated response for positions
type StockPositionsResponse struct {
	Positions []PositionWithUnitResponse `json:"positions"`
	Total     int64                      `json:"total"`
	Page      int                        `json:"page"`
	Limit     int                        `json:"limit"`
	Pages     int                        `json:"pages"`
}

// --------
// UNITS CONFIG
// --------

// AllowedUnits — список разрешённых единиц измерения
var AllowedUnits = []string{
	"pcs",
	"kg",
	"g",
	"l",
	"ml",
	"m",
	"cm",
}

// allowedUnitsSet — быстрый поиск для валидации
var allowedUnitsSet = func() map[string]bool {
	m := make(map[string]bool, len(AllowedUnits))
	for _, u := range AllowedUnits {
		m[u] = true
	}
	return m
}()

// IsValidUnit проверяет, что единица измерения допустима
func IsValidUnit(unit string) bool {
	return allowedUnitsSet[unit]
}

// --------
// VALIDATION HELPERS
// --------

// IsValidStockDirection проверяет валидность направления
func IsValidStockDirection(d StockDirection) bool {
	return d == StockDirectionIncome || d == StockDirectionOutcome
}

// IsValidStockBatchStatus проверяет валидность статуса
func IsValidStockBatchStatus(s StockBatchStatus) bool {
	switch s {
	case StockBatchStatusDraft, StockBatchStatusLabeled,
		StockBatchStatusConfirmed, StockBatchStatusCancelled:
		return true
	}
	return false
}

// IsValidStockPositionType проверяет валидность типа позиции
func IsValidStockPositionType(t StockPositionType) bool {
	return t == StockPositionTypeBatch || t == StockPositionTypeSerial
}

// IsValidStockMovementType проверяет валидность типа движения
func IsValidStockMovementType(t StockMovementType) bool {
	switch t {
	case StockMovementTypeWriteOff, StockMovementTypeReturn,
		StockMovementTypeTransfer, StockMovementTypeAdjustment:
		return true
	}
	return false
}

// --------
// HELPERS
// --------

func (s StockBatchStatus) CanTransitionTo(next StockBatchStatus) bool {
	switch s {
	case StockBatchStatusDraft:
		return next == StockBatchStatusLabeled ||
			next == StockBatchStatusConfirmed ||
			next == StockBatchStatusCancelled
	case StockBatchStatusLabeled:
		return next == StockBatchStatusConfirmed ||
			next == StockBatchStatusCancelled
	case StockBatchStatusConfirmed, StockBatchStatusCancelled:
		return false // финальные статусы
	}
	return false
}

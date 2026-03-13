package dm

import (
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/wm"
	"time"
)

// ---------
// TYPES
// ---------

type DealType struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateDealTypeRequest struct {
	Name    string  `json:"name" validate:"required,min=1,max=255"`
	Comment *string `json:"comment,omitempty"`
}

type UpdateDealTypeRequest struct {
	Name    *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment *string `json:"comment,omitempty"`
}

// ---------
// STATUSES
// ---------

type DealStatus struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Comment   *string   `json:"comment"`
	SortOrder int       `json:"sort_order"`
	Color     *string   `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateDealStatusRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=255"`
	Comment   *string `json:"comment,omitempty"`
	SortOrder int     `json:"sort_order" validate:"min=1"`
	Color     *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

type UpdateDealStatusRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Comment   *string `json:"comment,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty" validate:"omitempty,min=1"`
	Color     *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// ---------
// DEALS
// ---------

// DealPosition represents a position in a deal
type DealPosition struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Comment    *string   `json:"comment"`
	Price      float64   `json:"price"`
	Quantity   float64   `json:"quantity"`
	Unit       string    `json:"unit"`
	UnitID     *string   `json:"unit_id"`     // ссылка на catalog_units, может быть null
	PositionID *string   `json:"position_id"` // ссылка на stock_positions, может быть null
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Вложенные сущности (могут быть null)
	CatalogUnit     *wm.CatalogUnit   `json:"catalog_unit"`
	CatalogPosition *wm.StockPosition `json:"catalog_position"`
}

// Deal represents a deal without positions (for list views)
type Deal struct {
	ID        string    `json:"id"`
	Comment   *string   `json:"comment"`
	TypeID    *string   `json:"type_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Вложенные сущности (могут быть null)
	ClientID  *string                `json:"client_id"`
	Client    *crm.ClientDetail      `json:"client"`
	Employees []hrm.EmployeeListItem `json:"employees"`
	Status    *DealStatus            `json:"status"`
	Type      *DealType              `json:"type"`
}

// DealWithPositions represents a deal with positions (for detail view)
type DealWithPositions struct {
	ID        string    `json:"id"`
	Comment   *string   `json:"comment"`
	TypeID    *string   `json:"type_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Вложенные сущности (могут быть null)
	ClientID  *string                `json:"client_id"`
	Client    *crm.ClientDetail      `json:"client"`
	Employees []hrm.EmployeeListItem `json:"employees"`
	Status    *DealStatus            `json:"status"`
	Type      *DealType              `json:"type"`
	Positions []DealPosition         `json:"positions"`
}

// CreateDealRequest represents request to create a deal
type CreateDealRequest struct {
	Comment *string `json:"comment,omitempty"`
	TypeID  *string `json:"type_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateDealPosition represents a position update in a deal
type UpdateDealPosition struct {
	ID         *string  `json:"id,omitempty"`   // для существующих позиций
	Name       *string  `json:"name,omitempty"` // для новых или обновления
	Comment    *string  `json:"comment,omitempty"`
	Price      *float64 `json:"price,omitempty"`
	Quantity   *float64 `json:"quantity,omitempty"`
	Unit       *string  `json:"unit,omitempty"`
	UnitID     *string  `json:"unit_id,omitempty"`
	PositionID *string  `json:"position_id,omitempty"`
	Delete     *bool    `json:"delete,omitempty"` // true - удалить позицию
}

// UpdateDealRequest represents request to update a deal with all related entities
type UpdateDealRequest struct {
	Comment   *string              `json:"comment,omitempty"`
	TypeID    *string              `json:"type_id,omitempty"`
	ClientID  *string              `json:"client_id,omitempty"`
	StatusID  *string              `json:"status_id,omitempty"`
	Employees []string             `json:"employees,omitempty"` // полная замена списка сотрудников
	Positions []UpdateDealPosition `json:"positions,omitempty"` // полная замена/обновление позиций
}

// GetDealsParams represents request params for listing deals
type GetDealsParams struct {
	Page       int     `json:"page" validate:"omitempty,min=1"`
	Limit      int     `json:"limit" validate:"omitempty,min=1,max=100"`
	TypeID     *string `json:"type_id,omitempty"`
	StatusID   *string `json:"status_id,omitempty"`
	ClientID   *string `json:"client_id,omitempty"`
	EmployeeID *string `json:"employee_id,omitempty"`
	Search     *string `json:"search,omitempty"`
	GroupBy    *string `json:"group_by,omitempty"` // "status" - группировка по статусам
}

// DealsResponse represents paginated response for deals
type DealsResponse struct {
	Deals []Deal `json:"deals"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Pages int    `json:"pages"`
}

// DealGroup represents a group of deals by status
type DealGroup struct {
	StatusID    string  `json:"status_id"`
	StatusName  string  `json:"status_name"`
	StatusColor *string `json:"status_color"`
	SortOrder   int     `json:"sort_order"`
	Deals       []Deal  `json:"deals"`
	Count       int     `json:"count"`
}

// DealsGroupedResponse represents grouped response for deals
type DealsGroupedResponse struct {
	Groups []DealGroup `json:"groups"`
	Total  int64       `json:"total"`
}

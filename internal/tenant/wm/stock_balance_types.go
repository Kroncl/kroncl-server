package wm

// StockBalanceItem представляет остаток по конкретному товару
type StockBalanceItem struct {
	UnitID    string      `json:"unit_id"`
	UnitName  string      `json:"unit_name"`
	Unit      CatalogUnit `json:"unit"`
	Quantity  float64     `json:"quantity"`  // текущий остаток
	Reserved  float64     `json:"reserved"`  // зарезервировано (если есть)
	Available float64     `json:"available"` // доступно для продажи
}

package adminpricing

type UpdatePlanRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	PricePerMonth     *int    `json:"price_per_month,omitempty"`
	PricePerYear      *int    `json:"price_per_year,omitempty"`
	LimitDbMB         *int    `json:"limit_db_mb,omitempty"`
	LimitObjectsMB    *int    `json:"limit_objects_mb,omitempty"`
	LimitObjectsCount *int    `json:"limit_objects_count,omitempty"`
}

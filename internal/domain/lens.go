package domain

import "time"

type Lens struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
	Dimensions  []LensDimension `json:"dimensions,omitempty"`
	DimCount    int             `json:"dim_count,omitempty"`
}

type LensDimension struct {
	DimensionID   string  `json:"dimension_id"`
	DimensionName string  `json:"dimension_name"`
	Weight        float64 `json:"weight"`
	SortOrder     int     `json:"sort_order"`
}

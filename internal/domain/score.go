package domain

import "time"

type ReviewDimension struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	SortOrder   int     `json:"sort_order"`
}

type Scorecard struct {
	ID       int64            `json:"id"`
	RepoID   string           `json:"repo_id"`
	ScoredAt time.Time        `json:"scored_at"`
	Overall  float64          `json:"overall"`
	Notes    string           `json:"notes"`
	Scores   []DimensionScore `json:"scores,omitempty"`
}

type DimensionScore struct {
	ID          int64   `json:"id"`
	ScorecardID int64   `json:"scorecard_id"`
	DimensionID string  `json:"dimension_id"`
	Score       float64 `json:"score"`
	Evidence    string  `json:"evidence"`
	Notes       string  `json:"notes"`
}

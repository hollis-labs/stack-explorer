package domain

import "time"

type Finding struct {
	ID          int64     `json:"id"`
	RepoID      *string   `json:"repo_id,omitempty"`
	Title       string    `json:"title"`
	Category    string    `json:"category"` // "gap", "strength", "opportunity", "risk"
	Severity    string    `json:"severity"` // "critical", "high", "medium", "low", "info"
	Description string    `json:"description"`
	Status      string    `json:"status"` // "open", "acknowledged", "resolved", "wontfix"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ComparisonSet struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ComparisonSetRepo struct {
	SetID  string `json:"set_id"`
	RepoID string `json:"repo_id"`
	Role   string `json:"role"` // "subject" or "reference"
}

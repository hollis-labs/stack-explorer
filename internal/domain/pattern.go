package domain

import "time"

type ArchitecturePattern struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "pattern" or "anti_pattern"
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RepoPattern struct {
	RepoID    string `json:"repo_id"`
	PatternID string `json:"pattern_id"`
	Quality   string `json:"quality"` // "exemplary", "present", "partial", "absent"
	Notes     string `json:"notes"`
}

type CodeReference struct {
	ID          int64     `json:"id"`
	RepoID      string    `json:"repo_id"`
	FilePath    string    `json:"file_path"`
	LineStart   *int      `json:"line_start,omitempty"`
	LineEnd     *int      `json:"line_end,omitempty"`
	Description string    `json:"description"`
	RefType     string    `json:"ref_type"`
	PatternID   *string   `json:"pattern_id,omitempty"`
	FindingID   *int64    `json:"finding_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

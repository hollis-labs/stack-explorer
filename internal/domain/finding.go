package domain

import (
	"fmt"
	"strings"
	"time"
)

// Canonical finding lifecycle statuses. These are the only values accepted by
// the finding write paths (CLI + MCP); anything else is rejected so status
// filters and `finding list --status` queries stay reliable.
const (
	FindingStatusOpen         = "open"
	FindingStatusAcknowledged = "acknowledged"
	FindingStatusResolved     = "resolved"
	FindingStatusWontfix      = "wontfix"
)

// FindingStatuses is the canonical, ordered set of allowed finding statuses.
var FindingStatuses = []string{
	FindingStatusOpen,
	FindingStatusAcknowledged,
	FindingStatusResolved,
	FindingStatusWontfix,
}

// ValidFindingStatus reports whether status is a recognized finding status.
func ValidFindingStatus(status string) bool {
	switch status {
	case FindingStatusOpen, FindingStatusAcknowledged, FindingStatusResolved, FindingStatusWontfix:
		return true
	default:
		return false
	}
}

// ValidateFindingStatus returns a descriptive error when status is not one of
// the canonical FindingStatuses, and nil otherwise.
func ValidateFindingStatus(status string) error {
	if ValidFindingStatus(status) {
		return nil
	}
	return fmt.Errorf("invalid finding status %q: must be one of %s", status, strings.Join(FindingStatuses, ", "))
}

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

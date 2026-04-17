package config

import (
	"os"
	"path/filepath"
)

// DefaultDBPath returns the default database path relative to the binary location.
func DefaultDBPath() string {
	// Check if running from project root (data/ exists)
	if _, err := os.Stat("data"); err == nil {
		return filepath.Join("data", "stack-explorer.db")
	}
	// Fall back to home directory
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".stack-explorer", "stack-explorer.db")
}

package domain

import "time"

type Relationship struct {
	ID           int64     `json:"id"`
	RepoID       string    `json:"repo_id"`
	SrcSymbolID  int64     `json:"src_symbol_id"`
	DstSymbolID  int64     `json:"dst_symbol_id"`
	Kind         string    `json:"kind"`
	Weight       float64   `json:"weight"`
	Source       string    `json:"source"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

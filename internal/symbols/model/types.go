package model

import (
	"context"
	"time"
)

type Symbol struct {
	ID               int64
	RepoID           string
	Kind             string
	Name             string
	QualifiedName    string
	FilePath         string
	LineStart        *int
	LineEnd          *int
	ContentHash      string
	SignatureHash    string
	ParentSymbolID   *int64
	Language         string
	Visibility       string
	Docstring        string
	StaleSinceCommit *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SearchFilter struct {
	RepoID   string
	Query    string
	Kind     string
	Language string
	Limit    int
}

type StatsFilter struct {
	RepoID string
}

type Stats struct {
	RepoID         string
	Total          int
	Stale          int
	ByKind         map[string]int
	ByLanguage     map[string]int
	ByLanguageKind map[string]map[string]int
}

type Store interface {
	FindSymbolByQualifiedName(repoID, qualifiedName string) (*Symbol, error)
	FindSymbolByFileQualifiedName(repoID, filePath, qualifiedName string) (*Symbol, error)
	FindSymbolByContentHash(repoID, contentHash string) (*Symbol, error)
	FindSymbolByFileContentHash(repoID, filePath, contentHash string) (*Symbol, error)
	SearchSymbols(filter SearchFilter) ([]Symbol, error)
	UpsertSymbol(sym *Symbol) error
}

type IngestRequest struct {
	RepoID    string
	RepoPath  string
	Languages []string
	CommitRef string
}

type IngestResult struct {
	Inserted int
	Updated  int
	Drifted  int
}

type Ingester interface {
	Ingest(context.Context, IngestRequest) (IngestResult, error)
}

type Config struct {
	WorkDir string
	Store   Store
}

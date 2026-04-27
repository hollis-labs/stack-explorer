package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/chrispian/stack-explorer/internal/domain"
	"github.com/chrispian/stack-explorer/internal/embed"
	"github.com/chrispian/stack-explorer/internal/store/sqlite"
	"github.com/chrispian/stack-explorer/internal/symbols/model"
)

func TestStoreUpsertAndSearchSymbols(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "symbols.db")
	store, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	if err := store.CreateRepo(&domain.Repo{ID: "stack-explorer", Name: "Stack Explorer"}); err != nil {
		t.Fatalf("create repo: %v", err)
	}

	lineStart := 10
	lineEnd := 42
	sym := &model.Symbol{
		RepoID:        "stack-explorer",
		Kind:          "function",
		Name:          "CreateRepo",
		QualifiedName: "github.com/chrispian/stack-explorer/internal/store/sqlite.(*Store).CreateRepo",
		FilePath:      "internal/store/sqlite/repo.go",
		LineStart:     &lineStart,
		LineEnd:       &lineEnd,
		ContentHash:   "body-hash",
		SignatureHash: "sig-hash",
		Language:      "go",
		Visibility:    "public",
		Docstring:     "CreateRepo inserts a repo row.",
	}
	if err := store.UpsertSymbol(sym); err != nil {
		t.Fatalf("upsert symbol: %v", err)
	}

	got, err := store.GetSymbol(sym.ID)
	if err != nil {
		t.Fatalf("get symbol: %v", err)
	}
	if got == nil || got.QualifiedName != sym.QualifiedName {
		t.Fatalf("got %#v, want qualified name %q", got, sym.QualifiedName)
	}

	foundByName, err := store.FindSymbolByQualifiedName(sym.RepoID, sym.QualifiedName)
	if err != nil {
		t.Fatalf("find by qualified name: %v", err)
	}
	if foundByName == nil || foundByName.ID != sym.ID {
		t.Fatalf("find by qualified name = %#v, want id %d", foundByName, sym.ID)
	}

	results, err := store.SearchSymbols(model.SearchFilter{RepoID: sym.RepoID, Query: "CreateRepo", Limit: 10})
	if err != nil {
		t.Fatalf("search symbols: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(search results) = %d, want 1", len(results))
	}

	stats, err := store.SymbolStats(model.StatsFilter{RepoID: sym.RepoID})
	if err != nil {
		t.Fatalf("symbol stats: %v", err)
	}
	if stats.Total != 1 || stats.ByLanguage["go"] != 1 || stats.ByKind["function"] != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}

	targets, err := store.ListSymbolEmbeddingTargets(sym.RepoID)
	if err != nil {
		t.Fatalf("list symbol embedding targets: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("len(symbol embedding targets) = %d, want 1", len(targets))
	}
	wantText := sym.QualifiedName + "\n" + sym.FilePath + "\n" + sym.Docstring
	if targets[0].Text != wantText {
		t.Fatalf("embedding text = %q, want %q", targets[0].Text, wantText)
	}
	if targets[0].ContentHash != embed.HashText(wantText) {
		t.Fatalf("embedding content hash = %q, want hash of embedding text", targets[0].ContentHash)
	}
}

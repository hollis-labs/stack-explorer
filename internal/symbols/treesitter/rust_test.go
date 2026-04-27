//go:build treesitter

package treesitter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRustFile(t *testing.T) {
	path := filepath.Join("testdata", "rust", "sample.rs")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	items, err := ParseRustFile("stack-explorer", "testdata/rust/sample.rs", src)
	if err != nil {
		t.Fatalf("parse rust file: %v", err)
	}
	if len(items) != 7 {
		t.Fatalf("len(items) = %d, want 7", len(items))
	}
	got := map[string]string{}
	for _, sym := range items {
		got[sym.QualifiedName] = sym.Kind
	}
	expect := map[string]string{
		"Store":               "type",
		"Store::save":         "method",
		"create_repo":         "function",
		"VERSION":             "const",
		"config":              "var",
		"nested":              "module",
		"nested::NestedStore": "type",
	}
	for qn, kind := range expect {
		if got[qn] != kind {
			t.Fatalf("symbol %q = %q, want %q", qn, got[qn], kind)
		}
	}
}

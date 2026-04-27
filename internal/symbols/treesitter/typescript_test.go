//go:build treesitter

package treesitter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTypeScriptFile(t *testing.T) {
	path := filepath.Join("testdata", "typescript", "sample.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	items, err := ParseTypeScriptFile("stack-explorer", "testdata/typescript/sample.ts", src, false)
	if err != nil {
		t.Fatalf("parse typescript file: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("len(items) = %d, want 4", len(items))
	}
	got := map[string]string{}
	for _, sym := range items {
		got[sym.QualifiedName] = sym.Kind
	}
	expect := map[string]string{
		"Store":      "type",
		"Store.save": "method",
		"createRepo": "function",
		"version":    "var",
	}
	for qn, kind := range expect {
		if got[qn] != kind {
			t.Fatalf("symbol %q = %q, want %q", qn, got[qn], kind)
		}
	}
}

//go:build treesitter

package treesitter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGoFile(t *testing.T) {
	path := filepath.Join("testdata", "go", "sample.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	symbols, err := ParseGoFile("stack-explorer", "testdata/go/sample.go", src)
	if err != nil {
		t.Fatalf("parse go file: %v", err)
	}
	if len(symbols) != 5 {
		t.Fatalf("len(symbols) = %d, want 5", len(symbols))
	}

	got := map[string]string{}
	for _, sym := range symbols {
		got[sym.QualifiedName] = sym.Kind
		if sym.ContentHash == "" || sym.SignatureHash == "" {
			t.Fatalf("missing hashes for %#v", sym)
		}
	}

	expect := map[string]string{
		"demo.Store":      "type",
		"demo.Version":    "const",
		"demo.CreateRepo": "function",
		"demo.Store.Save": "method",
	}
	for qualifiedName, kind := range expect {
		if got[qualifiedName] != kind {
			t.Fatalf("symbol %q = %q, want %q", qualifiedName, got[qualifiedName], kind)
		}
	}
	if got["demo.global"] != "var" {
		t.Fatalf("symbol %q = %q, want %q", "demo.global", got["demo.global"], "var")
	}
}

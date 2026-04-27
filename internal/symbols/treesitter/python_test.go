//go:build treesitter

package treesitter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePythonFile(t *testing.T) {
	path := filepath.Join("testdata", "python", "sample.py")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	items, err := ParsePythonFile("stack-explorer", "testdata/python/sample.py", src)
	if err != nil {
		t.Fatalf("parse python file: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("len(items) = %d, want 4", len(items))
	}
	got := map[string]string{}
	for _, sym := range items {
		got[sym.QualifiedName] = sym.Kind
	}
	expect := map[string]string{
		"Store":       "type",
		"Store.save":  "method",
		"create_repo": "function",
		"VERSION":     "var",
	}
	for qn, kind := range expect {
		if got[qn] != kind {
			t.Fatalf("symbol %q = %q, want %q", qn, got[qn], kind)
		}
	}
}

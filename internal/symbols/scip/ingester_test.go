package scip

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chrispian/stack-explorer/internal/symbols/model"
	scippb "github.com/scip-code/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"
)

func TestCollectSymbolsFromIndex(t *testing.T) {
	repoPath := t.TempDir()
	source := "package demo\n\nfunc Hello(name string) string {\n\treturn name\n}\n"
	sourcePath := filepath.Join(repoPath, "hello.go")
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	rawSymbol := scippb.VerboseSymbolFormatter.FormatSymbol(&scippb.Symbol{
		Scheme: "scip-go",
		Package: &scippb.Package{
			Manager: "gomod",
			Name:    "example.com/demo",
			Version: "v0.0.0",
		},
		Descriptors: []*scippb.Descriptor{
			{Name: "Hello", Suffix: scippb.Descriptor_Method},
		},
	})

	index := &scippb.Index{
		Documents: []*scippb.Document{
			{
				Language:     "go",
				RelativePath: "hello.go",
				Symbols: []*scippb.SymbolInformation{
					{
						Symbol:      rawSymbol,
						DisplayName: "Hello",
						Kind:        scippb.SymbolInformation_Function,
						Documentation: []string{
							"Hello returns the provided name.",
						},
						SignatureDocumentation: &scippb.Document{Language: "go", Text: "func Hello(name string) string"},
					},
				},
				Occurrences: []*scippb.Occurrence{
					{
						Range:       []int32{2, 0, 2, 5},
						Symbol:      rawSymbol,
						SymbolRoles: int32(scippb.SymbolRole_Definition),
					},
				},
			},
		},
	}
	data, err := proto.Marshal(index)
	if err != nil {
		t.Fatalf("marshal index: %v", err)
	}

	parsed, err := collectSymbols(&scippb.Index{Documents: index.Documents}, model.IngestRequest{
		RepoID:   "demo",
		RepoPath: repoPath,
	})
	if err != nil {
		t.Fatalf("collect symbols: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("len(parsed) = %d, want 1", len(parsed))
	}
	got := parsed[0].symbol
	if got.QualifiedName == "" || got.ContentHash == "" || got.SignatureHash == "" {
		t.Fatalf("unexpected parsed symbol: %#v", got)
	}
	if got.LineStart == nil || *got.LineStart != 3 {
		t.Fatalf("line start = %#v, want 3", got.LineStart)
	}
	_ = data
}

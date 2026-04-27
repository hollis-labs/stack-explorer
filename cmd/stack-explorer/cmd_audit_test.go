package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrispian/stack-explorer/internal/domain"
	"github.com/chrispian/stack-explorer/internal/store/sqlite"
)

func TestAuditImportAndListCLI(t *testing.T) {
	fixtureDir := t.TempDir()
	writeAuditFixture(t, fixtureDir)

	dbDir := t.TempDir()
	dbPath = filepath.Join(dbDir, "test.db")
	tempStore, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := tempStore.CreateRepo(&domain.Repo{ID: "nanite", Name: "Nanite"}); err != nil {
		t.Fatalf("seed repo: %v", err)
	}
	_ = tempStore.Close()

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"audit", "import", fixtureDir, "--repo", "nanite"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("execute import: %v", err)
		}
		rootCmd.SetArgs([]string{"audit", "list", "--repo", "nanite"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("execute list: %v", err)
		}
	})
	if !strings.Contains(output, "Imported audit #") {
		t.Fatalf("import output missing success line: %s", output)
	}
	if !strings.Contains(output, "example-scope") {
		t.Fatalf("list output missing audit scope: %s", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	_ = w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	return buf.String()
}

func writeAuditFixture(t *testing.T, root string) {
	t.Helper()
	index := `# Example scope — deep review

**Date:** 2026-04-10
**Scope:** ` + "`example-scope`" + `
**Reviewer agent:** ` + "`nanite-reviewer`" + `
**Skill:** ` + "`deep-review`" + `

## Scope

Fixture summary.

## Findings

### By severity

- [01 — Example critical issue](01-critical-example-critical-issue.md)

### By topic

**Security**
- [01 — Example critical issue](01-critical-example-critical-issue.md)
`
	finding := `# [Critical] Example critical issue

**Scope:** example-scope

## Problem

The parser should preserve the first problem paragraph.

## References

- ` + "`internal/chat/engine.go:L10-L14`" + `
`
	if err := os.WriteFile(filepath.Join(root, "index.md"), []byte(index), 0o644); err != nil {
		t.Fatalf("write index fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "01-critical-example-critical-issue.md"), []byte(finding), 0o644); err != nil {
		t.Fatalf("write finding fixture: %v", err)
	}
}

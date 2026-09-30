#!/usr/bin/env bash
# Export Stack Explorer runtime state for transfer to another machine.
#
# Produces a tarball in the parent directory containing:
#   - db.sql          SQLite dump (portable, re-importable via sqlite3)
#   - scans/          data/scans/ raw scan outputs (gitignored)
#   - reports/        generated reports (gitignored)
#   - repos.yaml      local repo catalog, if present (gitignored; see repos.example.yaml)
#   - MANIFEST.txt    metadata: counts, DB stats, source host, timestamp
#
# Usage:
#   ./scripts/export-state.sh            # exports to ../stack-explorer-state-YYYYMMDD-HHMMSS.tar.gz
#   ./scripts/export-state.sh /tmp       # custom output dir
#
# Import on the new machine:
#   1. Clone repo, `make build`
#   2. Extract tarball
#   3. sqlite3 data/stack-explorer.db < db.sql
#   4. Copy scans/ -> data/scans/  and  reports/ -> reports/
#   5. Copy repos.yaml into the repo root and fix any local_path entries
#      that point at the source machine.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${1:-$(dirname "$REPO_ROOT")}"
STAMP="$(date +%Y%m%d-%H%M%S)"
STAGE="$(mktemp -d)"
BUNDLE="$STAGE/stack-explorer-state-$STAMP"
TARBALL="$OUT_DIR/stack-explorer-state-$STAMP.tar.gz"

mkdir -p "$BUNDLE"

DB="$REPO_ROOT/data/stack-explorer.db"
if [[ ! -f "$DB" ]]; then
  echo "error: database not found at $DB" >&2
  exit 1
fi

echo "==> dumping DB"
sqlite3 "$DB" ".dump" > "$BUNDLE/db.sql"

echo "==> copying scans (if present)"
if [[ -d "$REPO_ROOT/data/scans" ]]; then
  cp -R "$REPO_ROOT/data/scans" "$BUNDLE/scans"
fi

echo "==> copying reports (if present)"
if [[ -d "$REPO_ROOT/reports" ]]; then
  mkdir -p "$BUNDLE/reports"
  # copy only .md files (skip .gitkeep)
  find "$REPO_ROOT/reports" -maxdepth 1 -type f -name '*.md' -exec cp {} "$BUNDLE/reports/" \;
fi

if [[ -f "$REPO_ROOT/repos.yaml" ]]; then
  echo "==> copying repos.yaml"
  cp "$REPO_ROOT/repos.yaml" "$BUNDLE/repos.yaml"
fi

echo "==> writing MANIFEST.txt"
{
  echo "Stack Explorer state export"
  echo "Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Source host: $(hostname)"
  echo "Source repo: $REPO_ROOT"
  echo ""
  echo "DB size (bytes): $(stat -f%z "$DB" 2>/dev/null || stat -c%s "$DB")"
  echo "DB tables:"
  sqlite3 "$DB" ".tables" | tr -s ' ' '\n' | sed 's/^/  /'
  echo ""
  echo "Scan files: $(find "$BUNDLE/scans" -type f 2>/dev/null | wc -l | tr -d ' ')"
  echo "Report files: $(find "$BUNDLE/reports" -type f 2>/dev/null | wc -l | tr -d ' ')"
  echo "Repos in catalog: $(grep -cE '^    - id:' "$BUNDLE/repos.yaml" 2>/dev/null || echo 0)"
} > "$BUNDLE/MANIFEST.txt"

echo "==> creating tarball"
tar -czf "$TARBALL" -C "$STAGE" "$(basename "$BUNDLE")"
rm -rf "$STAGE"

echo ""
echo "✓ Exported to: $TARBALL"
echo "  size: $(du -h "$TARBALL" | cut -f1)"

#!/usr/bin/env bash
# cleanup.sh — Remove cloned repos from tmp/repos/.
#
# Usage:
#   ./scripts/cleanup.sh              # remove all cloned repos
#   ./scripts/cleanup.sh <repo-id>    # remove one
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TMP_DIR="$(dirname "$SCRIPT_DIR")/tmp/repos"

if [[ ! -d "$TMP_DIR" ]]; then
  echo "Nothing to clean up"
  exit 0
fi

if [[ $# -gt 0 ]]; then
  target="$TMP_DIR/$1"
  if [[ -d "$target" ]]; then
    rm -rf "$target"
    echo "Removed $1"
  else
    echo "Not found: $1"
  fi
else
  count=$(find "$TMP_DIR" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')
  if [[ "$count" -eq 0 ]]; then
    echo "Nothing to clean up"
  else
    du -sh "$TMP_DIR" 2>/dev/null || true
    rm -rf "$TMP_DIR"/*
    echo "Removed $count cloned repo(s)"
  fi
fi

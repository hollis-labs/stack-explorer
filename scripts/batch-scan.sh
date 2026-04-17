#!/usr/bin/env bash
# batch-scan.sh — Run Hadron blueprints across repos in the stack-explorer DB.
#
# Repos are resolved in order:
#   1. local_path override (if set and exists) — for own/unreleased projects
#   2. Clone from url to tmp/repos/<id> — for everything else
#
# Usage:
#   ./scripts/batch-scan.sh                    # scan all repos
#   ./scripts/batch-scan.sh --own              # scan own projects only
#   ./scripts/batch-scan.sh --category memory  # scan one category
#   ./scripts/batch-scan.sh --audit            # feature audit only (skip repo-scan)
#   ./scripts/batch-scan.sh --scan             # repo scan only (skip feature audit)
#   ./scripts/batch-scan.sh --dry-run          # show what would run
#   ./scripts/batch-scan.sh --no-cleanup       # keep cloned repos after scan
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DB="$PROJECT_DIR/data/stack-explorer.db"
BP_DIR="$PROJECT_DIR/blueprints"
TMP_DIR="$PROJECT_DIR/tmp/repos"

CATEGORY=""
OWN_ONLY=false
DRY_RUN=false
RUN_SCAN=true
RUN_AUDIT=true
CLEANUP=true

while [[ $# -gt 0 ]]; do
  case "$1" in
    --category) CATEGORY="$2"; shift 2 ;;
    --own) OWN_ONLY=true; shift ;;
    --dry-run) DRY_RUN=true; shift ;;
    --audit) RUN_SCAN=false; shift ;;
    --scan) RUN_AUDIT=false; shift ;;
    --no-cleanup) CLEANUP=false; shift ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

# Build SQL query — fetch id, local_path, url
QUERY="SELECT id, local_path, url FROM repos WHERE 1=1"
if $OWN_ONLY; then
  QUERY="$QUERY AND is_own = 1"
fi
if [[ -n "$CATEGORY" ]]; then
  QUERY="$QUERY AND category = '$CATEGORY'"
fi
QUERY="$QUERY ORDER BY category, id"

REPOS=$(sqlite3 "$DB" "$QUERY")
TOTAL=$(echo "$REPOS" | grep -c '|' || true)

echo "=== Stack Explorer Batch Scan ==="
echo "Repos to process: $TOTAL"
echo ""

mkdir -p "$TMP_DIR"

COUNT=0
FAILED=0
CLONED=()

for line in $REPOS; do
  IFS='|' read -r id local_path url <<< "$line"
  COUNT=$((COUNT + 1))

  # Resolve repo path: local_path override > clone from URL
  repo_path=""
  cloned=false

  if [[ -n "$local_path" && -d "$local_path" ]]; then
    repo_path="$local_path"
  elif [[ -n "$url" && "$url" != "" ]]; then
    clone_target="$TMP_DIR/$id"
    if [[ -d "$clone_target" ]]; then
      repo_path="$clone_target"
    elif ! $DRY_RUN; then
      echo "[$COUNT/$TOTAL] $id — cloning..."
      if git -c filter.lfs.smudge= -c filter.lfs.process= -c filter.lfs.required=false clone --depth 1 --quiet "$url" "$clone_target" 2>/dev/null; then
        repo_path="$clone_target"
        cloned=true
        CLONED+=("$id")
      else
        echo "[$COUNT/$TOTAL] SKIP $id — clone failed"
        FAILED=$((FAILED + 1))
        continue
      fi
    else
      echo "[$COUNT/$TOTAL] $id"
      echo "  [dry-run] git clone --depth 1 $url $clone_target"
      if $RUN_AUDIT; then
        echo "  [dry-run] hadron run $BP_DIR/se-feature-audit.yaml --input repo_path=$clone_target --input repo_id=$id"
      fi
      if $RUN_SCAN; then
        echo "  [dry-run] hadron run $BP_DIR/se-repo-scan.yaml --input repo_path=$clone_target --input repo_id=$id"
      fi
      echo ""
      continue
    fi
  else
    echo "[$COUNT/$TOTAL] SKIP $id — no local_path or url"
    FAILED=$((FAILED + 1))
    continue
  fi

  echo "[$COUNT/$TOTAL] $id"

  if $RUN_AUDIT; then
    if $DRY_RUN; then
      echo "  [dry-run] hadron run $BP_DIR/se-feature-audit.yaml --input repo_path=$repo_path --input repo_id=$id"
    else
      echo "  Running feature audit..."
      hadron run "$BP_DIR/se-feature-audit.yaml" \
        --input "repo_path=$repo_path" \
        --input "repo_id=$id" 2>&1 | tail -3
    fi
  fi

  if $RUN_SCAN; then
    if $DRY_RUN; then
      echo "  [dry-run] hadron run $BP_DIR/se-repo-scan.yaml --input repo_path=$repo_path --input repo_id=$id"
    else
      echo "  Running repo scan..."
      hadron run "$BP_DIR/se-repo-scan.yaml" \
        --input "repo_path=$repo_path" \
        --input "repo_id=$id" 2>&1 | tail -3
    fi
  fi

  echo ""
done

echo "=== Complete ==="
echo "Processed: $COUNT | Failed: $FAILED | Cloned: ${#CLONED[@]}"

if $CLEANUP && [[ ${#CLONED[@]} -gt 0 ]]; then
  echo ""
  echo "Cleaning up ${#CLONED[@]} cloned repo(s)..."
  "$SCRIPT_DIR/cleanup.sh"
fi

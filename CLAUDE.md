# Stack Explorer

AI agent tooling research and comparison platform. Tracks, scores, and compares repos across the AI agent ecosystem for gap analysis, pattern discovery, and publishable scorecards.

## Build

```bash
make build     # -> ./stack-explorer
make install   # -> ~/go/bin/stack-explorer
make test
make lint
```

## Stack

- **Language:** Go 1.26+
- **Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGo)
- **CLI:** Cobra
- **Config:** YAML (gopkg.in/yaml.v3)
- **Analysis:** scc (LoC/complexity), golangci-lint (Go quality)

## Key Paths

| Path | Purpose |
|------|---------|
| `cmd/stack-explorer/` | Cobra CLI commands |
| `internal/store/sqlite/` | SQLite store, migrations, CRUD |
| `internal/domain/` | Domain types |
| `internal/config/` | App config |
| `data/stack-explorer.db` | SQLite database (gitignored) |
| `blueprints/` | Hadron blueprints for automated analysis |
| `skills/` | agentrc skills |
| `scripts/` | Batch scan, ingest, seed, cleanup scripts |
| `reports/` | Generated output (index.md is the entry point) |
| `tmp/repos/` | Temporary cloned repos (gitignored, auto-cleaned) |

## Database

- 13 migrations in `internal/store/sqlite/migrations/` (embedded SQL, numbered)
- WAL mode + foreign keys enabled
- 18 review dimensions, 9 scoring lenses seeded on first run
- Run `./stack-explorer db stats` to check table counts

## CLI Commands

```
repo add|update|list|show|remove|import|export  # Manage catalog
tag add|remove|list                              # Manage tags
snapshot take|history                            # Time-series metrics
dimension list|add                               # Review dimensions
lens list|show|create|add-dim                    # Scoring lenses
score set|get|recalc                             # Score repos (supports --lens)
scorecard generate                               # Markdown scorecards
pattern add|list|link                            # Architecture patterns
finding add|list                                 # Research findings
gap compare-set create|list                      # Comparison sets
db stats                                         # Database statistics
```

## Scoring Lenses

Lenses define which dimensions matter and how they're weighted for different perspectives. A repo can be viewed through any lens without re-scoring.

| Lens | Use For |
|------|---------|
| `agent-platform` | Full agent platforms (Conduit, Codex, Gemini CLI) |
| `chat-app` | GUI AI chat apps — tool calls, security, enterprise, features |
| `automation` | Workflow/DAG tools (Hadron, Dagster, Prefect) |
| `memory-system` | Memory/RAG systems (Cortex, GraphRAG) |
| `agent-framework` | Agent SDKs (agentrc, ADK, Strands) |
| `desktop-app` | Desktop apps (Nanite, Sigil) |
| `infra-tool` | Infrastructure tools (Cerberus) |
| `content-pipeline` | Content pipelines (Carrier) |
| `general` | All 18 dimensions equally weighted |

Usage: `./stack-explorer score get conduit --lens chat-app`

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/batch-scan.sh` | Run blueprints across repos (`--own`, `--category`, `--audit`, `--scan`, `--dry-run`, `--no-cleanup`) |
| `scripts/ingest-scans.sh` | Parse scan JSON and create snapshots |
| `scripts/seed-data.sh` | Seed patterns and comparison sets (safe to re-run) |
| `scripts/cleanup.sh` | Remove cloned repos from tmp/ |

## Conventions

- Repo IDs are lowercase slugs matching the repo name
- URLs are HTTPS (not SSH) — used for clone-on-demand
- `local_path` is an optional override for unreleased/speculative projects
- Scores are 0.0-10.0 with evidence text
- All timestamps are RFC3339 UTC
- Patterns are "pattern" or "anti_pattern" type
- Findings use categories: gap, strength, opportunity, risk
- Own projects have `is_own = true`

## Hadron Blueprints

- `se-repo-scan` — scc + golangci-lint + git stats (JSON output)
- `se-security-scan` — Security tool runner
- `se-feature-audit` — Pattern detection (hooks, MCP, memory, agents, loops, tools)

Run via: `hadron run /Users/chrispian/Projects-apps/stack-explorer/blueprints/se-repo-scan.yaml --input repo_path=... --input repo_id=...`

**Important:** Hadron resolves blueprint paths server-side. Always use absolute paths.

## Skills

| Skill | Purpose |
|-------|---------|
| `onboard-repo` | Full onboarding: add, tag, scan, ingest, score |
| `batch-score` | Score multiple repos efficiently |
| `create-lens` | Define custom scoring perspectives |
| `setup-comparison` | Create gap analysis comparison sets |
| `add-dimension` | Add new review dimensions |
| `add-repo` | Quick repo registration |
| `run-analysis` | Execute Hadron blueprints |
| `generate-scorecard` | Render markdown scorecards |
| `gap-analysis` | Run gap analysis on comparison sets |

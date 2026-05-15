# Stack Explorer — Agent Orientation

## What is this and why

Stack Explorer is a Go-based research and analytics platform for the AI agent and
developer-tooling ecosystem. It tracks, scores, and compares repositories through
configurable **lenses** (perspectives) and **dimensions** (review criteria), and
produces publishable scorecards, gap analyses, and findings.

It serves a second, emerging role: an **app-specific RAG-for-code** knowledge base.
A layered knowledge stack (symbols → knowledge → retrieval) lets agents retrieve
prior-art and architectural findings during reasoning. In this role Stack Explorer
is positioned as the portfolio's ingester of framework/code knowledge and a pipeline
partner for Tesseract's Knowledge domain.

Why it exists: a portfolio-wide audit (April 2026) showed every project produces
findings and architecture observations that live as scattered markdown. Stack
Explorer consolidates audits, symbols, and findings into one structured, queryable,
daemon-accessible store.

## Where to start

- `README` equivalent / quick reference: `CLAUDE.md` (build, stack, key paths, CLI).
- **Build:** `make build` → `./stack-explorer`. Also `make build-full` (treesitter
  build tag), `make install`, `make test`, `make lint`, `make eval`.
- **CLI entry point:** `cmd/stack-explorer/` — Cobra command groups (one file per
  group: `cmd_repo.go`, `cmd_score.go`, `cmd_audit.go`, `cmd_symbol.go`, etc.).
- **HTTP API:** `internal/api/` — Chi-based REST server. Start with
  `./stack-explorer serve --port 8080`.
- **MCP server:** `./stack-explorer mcp --transport stdio` (or `--transport http`).
  Tool reference: `docs/mcp-tools.md`.
- **The blueprint document:** `docs/proposed-knowledge-base-architecture.md` — the
  layered architecture (phases A–G), open questions, and schema additions.
- **Product framing:** `docs/product-vision.md`, `docs/portfolio-summary-2026-04.md`.

## Key domain concepts

- **Repo** — a tracked repository in the catalog (lowercase-slug ID, HTTPS clone
  URL, optional `local_path`, `is_own` flag).
- **Lens** — a scoring perspective that selects and weights dimensions
  (`agent-platform`, `chat-app`, `automation`, `memory-system`, `agent-framework`,
  `desktop-app`, `infra-tool`, `content-pipeline`, `general`). A repo can be viewed
  through any lens without re-scoring.
- **Dimension** — a review criterion (18 seeded). Scores are 0.0–10.0 with evidence text.
- **Snapshot** — a point-in-time metrics capture for a repo (time-series).
- **Pattern** — an architecture `pattern` or `anti_pattern`, linkable to repos.
- **Finding** — a research observation categorized as `gap`, `strength`,
  `opportunity`, or `risk`.
- **Audit** — a deep-review lifecycle (start/finish/import/export/diff); the ingest
  path for structured review findings.
- **Symbol** — function/type/const anchors extracted via tree-sitter, content-hashed.
- **Knowledge / Retrieval layers** — audits, themes, code refs; FTS5 BM25 + vector
  search (vectors via Tesseract) with reciprocal-rank fusion.

## Common operations + examples

```bash
# Build
make build

# Inspect the database
./stack-explorer db stats

# Score a repo through a specific lens
./stack-explorer score get conduit --lens chat-app

# Generate a markdown scorecard
./stack-explorer scorecard generate

# Add and onboard a repo
./stack-explorer repo add <slug> --url https://github.com/...
./stack-explorer snapshot take <slug>

# Run the REST API for the Sigil frontend
./stack-explorer serve --port 8080

# Run the MCP server (stdio)
./stack-explorer mcp --transport stdio

# Retrieval-quality eval
make eval

# Run a Hadron analysis blueprint (use absolute paths — Hadron resolves server-side)
hadron run /Users/chrispian/dev/hollis-labs/apps/stack-explorer/blueprints/se-repo-scan.yaml \
  --input repo_path=... --input repo_id=...
```

Batch scripts live in `scripts/`: `batch-scan.sh`, `ingest-scans.sh`,
`seed-data.sh` (safe to re-run), `cleanup.sh`.

## Where to look for more

- `CLAUDE.md` — the canonical build/CLI/conventions quick reference.
- `docs/proposed-knowledge-base-architecture.md` — full layered architecture.
- `docs/mcp-tools.md` — MCP tool surface (`repo_context`, `prior_art_for_file`, …).
- `docs/audit-format.md` — deep-review audit authoring format.
- `docs/product-vision.md`, `docs/portfolio-summary-2026-04.md` — strategy/context.
- `docs/install-on-new-machine.md` — setup on a fresh machine.
- `internal/store/sqlite/migrations/` — numbered SQL migrations (schema source of truth).
- `blueprints/` — three Hadron blueprints (`se-repo-scan`, `se-security-scan`,
  `se-feature-audit`).
- `skills/` — nine agentrc skills (`onboard-repo`, `batch-score`, `create-lens`, …).
- `.agentrc/` — agent definitions (researcher, analyst, curator, backend).
- Knowledge file: `~/dev/agent-os/knowledge/projects/stack-explorer.md`.

> Note: `docs/stack-explorer-api-prompt.md` is a stale boot-spec. The HTTP API it
> describes as "to be built" is already live in `internal/api/`. Treat it as
> historical reference only.

## Conventions

- Repo IDs are lowercase slugs matching the repo name.
- Clone URLs are HTTPS (not SSH); `local_path` overrides for unreleased projects.
- All timestamps are RFC3339 UTC.
- SQLite via `modernc.org/sqlite` — pure Go, no CGo (tree-sitter build is the one
  optional CGo path, gated behind the `treesitter` build tag).

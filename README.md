# Stack Explorer

Stack Explorer is a research and knowledge-retrieval platform for the AI
agent and developer-tooling ecosystem. It keeps a SQLite catalog of
repositories, scores them through weighted lenses, extracts symbols and code
relationships, and serves the result over a CLI, a REST API, and an MCP
endpoint. It does not scan code itself — quantitative analysis comes from
Hadron blueprints — and it is not a task tracker, an agent launcher, or a
durable store for project knowledge.

> **Pre-release.** Stack Explorer already runs internally — a live catalog,
> real imported audits, and an MCP surface other tools call today — but it
> has no public release, no external consumers, and no compatibility
> guarantees. Built in the open: this README describes what's running, not a
> pitch for what's planned.

## What it is today

- **Repo catalog.** Repos → snapshots (time-series metrics) → scores
  (per-dimension, per-lens) → patterns → findings → tags. A repo scored once
  can be viewed through any lens (`agent-platform`, `chat-app`,
  `memory-system`, `agent-framework`, and others).
- **Audits as data.** Deep-review audits (Nanite's, first) are imported as
  queryable rows — `audit start|finish|list|show|import|export|diff` — rather
  than living as one-off markdown.
- **Hybrid retrieval.** `internal/retrieval/` fuses FTS5 BM25 with vector
  search over `internal/embed/`, so `search "panic recovery" --repo nanite`
  ranks findings and symbols together.
- **An MCP surface.** `internal/mcp/` exposes six read tools —
  `repo_context`, `prior_art_for_file`, `finding_search`, `symbol_lookup`,
  `audit_show`, `knowledge_query` — over stdio or HTTP. See
  [`docs/mcp-tools.md`](docs/mcp-tools.md).
- **Deployed, not demo.** Runs as a standing local process
  (`stack-explorer serve`, Cerberus-managed) with a live, multi-hundred-MB
  catalog behind it.

## Where it sits in the stack

```
   agents / sessions     Nanite, Claude Code, or any MCP-capable session
         │  MCP: repo_context, prior_art_for_file, finding_search, ...
    ┌────────────────┐
    │ Stack Explorer │   catalog + scoring + retrieval — read-heavy
    └────────────────┘
         │
   Hadron blueprints     quantitative scans that feed snapshots/scores
                          (Stack Explorer doesn't scan; Hadron does)
```

It's a peer to Tesseract and Torque, not a replacement for either: Stack
Explorer answers "what does the ecosystem look like / what did we already
find here," not "what's the plan" or "what did we decide."

## Examples

**Daily driver.** Before touching a file, pull prior art on it —
`prior_art_for_file(path="internal/chat/engine.go")` — to see findings and
audit history without re-deriving them. Comparing tools across the ecosystem
(leaderboards, gap analyses against a lens) lives in `reports/`.

**Composition.** A Hadron blueprint scans a repo and produces the raw
snapshot/score data; Stack Explorer catalogs and ranks it. An agent session
(Nanite, Claude Code) calls Stack Explorer's MCP tools mid-session to pull
that context before editing code. When a gap surfaces that's worth acting on,
it becomes a Torque task or a Tesseract decision — Stack Explorer doesn't
track that itself.

## Roadmap

- **Comparative re-audits.** Diff a fresh audit against a prior one to see
  what changed, instead of re-auditing a repo from scratch. Nanite's 31
  deep-review audits are the first case.
- **Autonomous multi-phase audits.** Using the audit + diff surface as the
  proving ground for running staged, multi-step audit work through Torque
  unattended once it's tested.
- **Ecosystem signal enrichment.** GitHub activity (stars, releases,
  contributors), stack/framework cross-referencing, and trend tracking —
  directional per `docs/product-vision.md`, not scheduled.
- **Sigil-based GUI.** A lens-switchable leaderboard and drill-down dashboard
  on top of the existing CLI/API/MCP surfaces.

## Commands

```bash
make build          # -> ./stack-explorer, pure Go
make test
make lint           # go vet ./...
make build-full     # adds the treesitter tag — the only CGo path here
```

`make eval` rewrites the tracked `eval/baseline.json`. Run it only when you
intend to move the retrieval baseline, and commit that diff deliberately.

See [`AGENTS.md`](AGENTS.md) for the subsystem map and repo boundaries.

# Stack Explorer — Knowledge Base Architecture (Proposed)

**Status:** Proposed / pre-planning draft
**Purpose:** Framing document for an architectural planning session. Not yet committed.
**Audience:** Future planning session (self + collaborators). Assumes familiarity with Stack Explorer's current codebase and the Nanite deep-review audit format.

---

## 1. Executive summary

Stack Explorer today is a repo catalog with quantitative scoring, dimensions, lenses, and a scan pipeline. The long-term goal is for it to become **THE source of architectural knowledge for a software project** — a daemon that tracks, audits, and keeps current a queryable knowledge base spanning findings, symbols, relationships, metrics, and narrative reviews, exposed over HTTP and MCP so any agent (Nanite, Claude Code, others) can pull prior knowledge into a session instead of rediscovering it.

This spec proposes the architecture to get there. It covers: the audit-as-first-class-entity extension (near-term), a symbol/relationship/embedding substrate (medium-term), scheduled refresh jobs (daemon-native), and how Stack Explorer reuses existing infrastructure (Vanta Conduit, Hadron, established OSS libraries) to avoid reinventing wheels.

**One-line positioning:** Stack Explorer is app-specific RAG for code — prior-art retrieval for agent reasoning, not documentation search.

---

## 2. Goals and non-goals

### Goals

- **Queryable architectural knowledge.** Every audit, finding, code reference, symbol, and relationship is a row in SQLite, not a file on disk.
- **Hybrid search.** BM25 + vector over narrative bodies and symbol descriptions, partitioned per project.
- **Symbol-level granularity.** Findings and embeddings anchor to stable code symbols, not line numbers.
- **Time travel.** Re-running an audit produces a diff against the previous pass, not a replacement.
- **Daemon-first delivery.** HTTP API + MCP server. No bundled UI; Sigil or any MCP-capable agent consumes it.
- **Scheduled maintenance.** Audits, git-history sweeps, and embedding refreshes run on schedule without operator intervention.
- **Project source flexibility.** Git repos first; expand to arbitrary local folders and installed binaries once the primitives are stable.
- **Low-churn foundation.** Once stable, the core shouldn't need rewriting. Favor boring, well-maintained dependencies.

### Non-goals

- **No bundled frontend.** Sigil already exists as a standalone UI for anyone who wants GUI access. Stack Explorer stays server-only.
- **No automated score mutation from audits.** Audits produce observations; scoring stays agent-assisted, human-reviewed.
- **No attempt to become a bug tracker.** Findings have lifecycles but Stack Explorer is not a workflow system. Don't add assignees, labels, or sprint planning.
- **No general-purpose RAG.** This is code knowledge, not a document store. Don't generalize prematurely.
- **No automated "fix this" agent loop.** Knowledge is read-mostly. Mutations come from explicit scans, imports, and CLI actions.

---

## 3. Current state (grounding)

Facts verified against the repo:

- **15 migrations** shipped. Last is `015_create_scans.sql`.
- **`findings` table has 1 row**, category `gap`. Namespace is effectively empty — the "collision risk" the earlier assessment flagged is a non-issue. Extending `findings` in place is safe.
- **`code_references` has 0 rows.** The primitive exists in schema but has never been exercised. The design is sound; it just hasn't been loaded.
- **`internal/api/` already exists** — chi-based HTTP server, `scan_worker.go` for async scan execution, handlers for repos/tags/snapshots/dimensions/lenses/scores/patterns/findings/comparison-sets/scans/reports. **The daemon shape is already real.** This reframes the roadmap: we are not building a daemon, we are extending the one that exists.
- **No `internal/mcp/` yet.** This is a gap.
- **No embeddings, FTS, graph, or symbol tables yet.** Clean slate for the knowledge layer.
- **Stack already locked to pure-Go SQLite** (`modernc.org/sqlite`). No CGo in the current build.

---

## 4. Architecture — layers

Think of Stack Explorer's knowledge base as a stack of layers, each building on the ones below. Higher layers can be built and shipped incrementally without disturbing lower ones.

```
┌─────────────────────────────────────────────────────────┐
│ Layer 6: Delivery        HTTP API │ MCP server           │
├─────────────────────────────────────────────────────────┤
│ Layer 5: Orchestration   Scheduler │ Event stream │ Jobs │
├─────────────────────────────────────────────────────────┤
│ Layer 4: Retrieval       BM25 (FTS5) │ Vector │ Hybrid   │
├─────────────────────────────────────────────────────────┤
│ Layer 3: Knowledge       Audits │ Findings │ Themes      │
│                          Patterns │ Relationships        │
├─────────────────────────────────────────────────────────┤
│ Layer 2: Symbols         Symbol extraction │ Anchors     │
├─────────────────────────────────────────────────────────┤
│ Layer 1: Projects        Repos │ Folders │ Binaries      │
│                          (generalized from today's repos)│
└─────────────────────────────────────────────────────────┘
```

### Layer 1 — Projects (generalized)

Today's `repos` table stays as the default case but generalizes to `source_type ∈ {git, local_folder, binary, archive}`. `binary_path`, `binary_hash` fields added for the binary case. This is **not** a first-phase change — only touch it once Layers 2–4 are solid. Noted here so we don't paint ourselves into a corner with `repo`-specific foreign keys downstream.

### Layer 2 — Symbols (the keystone)

**New table: `symbols`** — the identity layer for "this thing in the code." Populated by a tree-sitter ingestor.

```sql
CREATE TABLE symbols (
    id               INTEGER PRIMARY KEY,
    repo_id          TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    kind             TEXT NOT NULL,  -- function, method, type, const, var, module, package
    name             TEXT NOT NULL,
    qualified_name   TEXT NOT NULL,  -- e.g. pkg/store.(*Store).CreateRepo
    file_path        TEXT NOT NULL,
    line_start       INTEGER,
    line_end         INTEGER,
    content_hash     TEXT NOT NULL,  -- hash of normalized body
    signature_hash   TEXT,           -- hash of signature only (stable across body edits)
    parent_symbol_id INTEGER REFERENCES symbols(id) ON DELETE CASCADE,
    language         TEXT NOT NULL,
    visibility       TEXT,           -- public / private / internal
    docstring        TEXT,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL
);
```

**Staleness strategy: content-hash the enclosing symbol.** On re-scan, look up by `(repo_id, qualified_name)` first, then by `content_hash`. A symbol whose content hash changed is flagged as "drifted" — anything anchored to it (findings, embeddings) gets a review marker, not silent deletion.

**Ingestion options (to decide in planning):**
1. Tree-sitter via official Go binding (`github.com/tree-sitter/go-tree-sitter`). CGo. Production-proven, full grammar ecosystem. **Recommended for correctness.**
2. SCIP ingestion (`github.com/sourcegraph/scip`). Run `scip-go`, `scip-typescript`, `scip-python` against repos; parse the resulting protobuf. Zero CGo in Stack Explorer. Gives cross-file definitions/references for free. **Recommended if the CGo on #1 is a dealbreaker** — likely the right primary path given the pure-Go preference.

**Strong recommendation: do both.** Tree-sitter for symbols and local queries (where we control the schema); SCIP for cross-file reference graphs (where we'd have to reinvent otherwise). They're complementary, not competing.

### Layer 3 — Knowledge (audits, findings, themes, relationships)

#### Audits as first-class entities

**New table: `audits`** — wrapper for a cohesive review pass.

```sql
CREATE TABLE audits (
    id                INTEGER PRIMARY KEY,
    repo_id           TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    scope             TEXT NOT NULL,        -- human label: "memory system", "websocket layer"
    scope_paths       TEXT NOT NULL DEFAULT '[]', -- JSON glob array
    audit_type        TEXT NOT NULL,        -- deep-review, plan-review, security, observability, ...
    auditor           TEXT NOT NULL,        -- agent or human identifier
    audited_at_ref    TEXT,                 -- git sha the audit applies to
    summary_markdown  TEXT NOT NULL DEFAULT '',
    verdict           TEXT,                 -- approve, approve-with-changes, reject, advisory
    status            TEXT NOT NULL DEFAULT 'in_progress', -- in_progress, completed, superseded
    supersedes_id     INTEGER REFERENCES audits(id),
    started_at        TEXT NOT NULL,
    finished_at       TEXT,
    created_at        TEXT NOT NULL,
    updated_at        TEXT NOT NULL
);
```

**`findings` extended** — add columns to the existing table. Empty namespace (1 row) means this is safe:

```sql
ALTER TABLE findings ADD COLUMN audit_id        INTEGER REFERENCES audits(id) ON DELETE SET NULL;
ALTER TABLE findings ADD COLUMN body_markdown   TEXT NOT NULL DEFAULT '';
ALTER TABLE findings ADD COLUMN audited_at_ref  TEXT;
ALTER TABLE findings ADD COLUMN is_out_of_scope INTEGER NOT NULL DEFAULT 0;
ALTER TABLE findings ADD COLUMN symbol_id       INTEGER REFERENCES symbols(id) ON DELETE SET NULL;
```

**`code_references` extended** — add symbol anchor and staleness marker:

```sql
ALTER TABLE code_references ADD COLUMN symbol_id          INTEGER REFERENCES symbols(id) ON DELETE SET NULL;
ALTER TABLE code_references ADD COLUMN anchor_content_hash TEXT;
ALTER TABLE code_references ADD COLUMN stale_since_commit  TEXT;
```

#### Cross-cutting themes

**New table: `audit_themes`** + **`finding_themes`** join. Themes are audit-level ("we keep finding no panic recovery"), distinct from `architecture_patterns` which are ecosystem-level ("how the ecosystem does memory"). M:N with findings. Themes can span audits.

#### Relationships (the code graph, without a graph DB)

**New table: `relationships`** — one table, many edge types.

```sql
CREATE TABLE relationships (
    id              INTEGER PRIMARY KEY,
    repo_id         TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    src_symbol_id   INTEGER NOT NULL REFERENCES symbols(id) ON DELETE CASCADE,
    dst_symbol_id   INTEGER NOT NULL REFERENCES symbols(id) ON DELETE CASCADE,
    kind            TEXT NOT NULL, -- calls, imports, implements, references, tests, documents, co-changed-with
    weight          REAL NOT NULL DEFAULT 1.0,
    source          TEXT NOT NULL, -- static, scip, git-history, audit, manual
    discovered_at   TEXT NOT NULL
);
CREATE INDEX idx_rel_src ON relationships(src_symbol_id, kind);
CREATE INDEX idx_rel_dst ON relationships(dst_symbol_id, kind);
```

**Traversal strategy:** SQLite recursive CTEs for shallow queries; Go-side BFS with prepared statements for deep traversal. No graph DB. At single-repo scale (<10M edges) this is correct. Revisit only when recursive CTEs demonstrably fail.

**Git history as a graph signal.** A scheduled nightly job walks `git log --name-only` and fills `relationships` with `kind=co-changed-with`, `source=git-history`. Co-change frequency is one of the strongest "these are actually coupled" signals in the research literature, and it's free to compute.

### Layer 4 — Retrieval

**Two indexes, one query surface.**

1. **BM25 via SQLite FTS5.** Free, stays in-DB, single backup. Mirror `findings.body_markdown`, `symbols.qualified_name + docstring`, and `code_references.description` into FTS5 virtual tables. No external search engine.

2. **Vectors via Vanta Conduit** (see §6 Integration Decisions below). Symbol-level and finding-level embeddings, stored in Vanta's `embeddings` table, partitioned by namespace (e.g. `app/stack-explorer/repo-<id>/symbols`).

3. **Hybrid retrieval service** — `internal/retrieval/` package. Input: query + filters. Output: ranked results merging BM25 + cosine + optional metadata filters. Reciprocal rank fusion for the merge (well-understood, no training needed). Per-query weight override for cases where one signal should dominate.

### Layer 5 — Orchestration (scheduler + events)

**New tables:** `schedules`, `jobs`, `job_events`.

- `schedules(id, name, kind, cron_expr, target_repo_id, payload_json, enabled, last_run_at, next_run_at)` — cron-spec'd recurring work.
- `jobs(id, schedule_id, kind, status, payload_json, started_at, finished_at, error)` — execution records.
- `job_events(id, job_id, kind, payload_json, ts)` — append-only event stream, queryable and streamable.

**Runtime:** `robfig/cron/v3` in-process. On startup, load enabled schedules from SQLite into the cron runtime. On mutation, rehydrate. This is the idiomatic pattern — persistent metadata, in-memory runtime.

**Job kinds (initial):**
- `scan` — run a Hadron blueprint against a repo
- `audit-refresh` — re-ingest audit markdown, diff against prior
- `git-history-sweep` — compute co-change relationships for a repo
- `symbol-reindex` — tree-sitter / SCIP ingest
- `embedding-refresh` — recompute vectors for new/drifted content
- `complexity-metrics` — run Lizard, store results as snapshots

**Event streaming:** SSE endpoint on the HTTP API (`GET /api/events?filter=...`). MCP surface exposes `event_stream_subscribe` as a long-running tool. Backed by the `job_events` table plus an in-memory subscriber channel set (patterns borrowed from Hadron's `execution.Manager` — see §6).

### Layer 6 — Delivery (HTTP + MCP)

**HTTP** — extends existing chi router with:
- `/api/audits`, `/api/audits/{id}/findings`, `/api/audits/{id}/diff`
- `/api/symbols/{id}`, `/api/symbols/search`
- `/api/relationships?src=&kind=`
- `/api/search` — hybrid BM25+vector with filters
- `/api/schedules`, `/api/jobs`, `/api/events` (SSE)

**MCP server — new `internal/mcp/` package.** The official Go SDK (`github.com/modelcontextprotocol/go-sdk`). Tools:

| Tool | Purpose |
|---|---|
| `project_list` | List known projects with source type and health |
| `repo_context` | Summary of a repo: top findings, top themes, recent audits, last activity |
| `prior_art_for_file` | Given a file path, return every open finding, theme, and code reference that touches it |
| `prior_art_for_symbol` | Same, but anchored to a symbol ID or qualified name |
| `finding_search` | Hybrid search over findings with filters (severity, status, category, audit) |
| `symbol_lookup` | Resolve name → symbol(s), with neighbors from the relationships graph |
| `audit_show` | Get an audit with its findings and themes |
| `audit_diff` | Compare two audits (same repo, different git refs) |
| `pattern_examples` | Show exemplars of a pattern across repos |
| `knowledge_query` | Freeform hybrid RAG query with optional namespace scoping |
| `event_stream_subscribe` | Subscribe to job/audit events for long-running agents |

**Authoring tools (carefully scoped):**

| Tool | Purpose |
|---|---|
| `finding_add` | Create a finding (optionally attached to an audit) |
| `finding_update_status` | Move finding through open → acknowledged → resolved → wontfix |
| `audit_import` | Ingest an audit folder (INDEX.md + finding-NN-*.md) |
| `audit_export` | Render an audit back to markdown |

Write tools have a clear provenance trail (auditor identity stored on the audit; `source=` on the finding). Don't expose destructive operations over MCP — deletion stays CLI-only.

---

## 5. Data model migration roadmap

Ordered. Each migration is independently shippable.

| # | Migration | Layer | Notes |
|---|---|---|---|
| 016 | Create `audits`; extend `findings` | 3 | Safe — `findings` has 1 row |
| 017 | `audit_import` CLI + `audit_export` | 3 | First real data in: ingest Nanite's 6 audits |
| 018 | Create `audit_themes`, `finding_themes` | 3 | M:N join |
| 019 | Create `symbols` | 2 | Empty until ingestor lands |
| 020 | Tree-sitter + SCIP ingestor (code, not schema) | 2 | Go first, then TS/Py/Rust |
| 021 | Extend `code_references` with symbol anchor | 3 | Backfill from existing rows is trivial (0 rows) |
| 022 | Create `relationships` | 3 | Empty until scip-ingest runs |
| 023 | FTS5 virtual tables over findings/symbols/code-refs | 4 | Triggers to keep them in sync |
| 024 | Integrate Vanta Conduit (see §6) | 4 | Adds Vanta's schema alongside SE's |
| 025 | Hybrid retrieval service | 4 | Code, not schema |
| 026 | Create `schedules`, `jobs`, `job_events` | 5 | Cron runtime wired up |
| 027 | MCP server (`internal/mcp/`) | 6 | First callable tools |
| 028 | Git-history sweeper (scheduled job) | 3/5 | Populates co-change edges |
| 029 | Lizard integration (complexity metrics snapshot) | 3/5 | Shell-out, CSV parse |
| 030 | Generalize `repos` → `projects` with source_type | 1 | Defer until 016–029 settled |

Expect this list to shift during planning; the **order is load-bearing** though. Audits first (cheapest, highest-value data). Symbols before relationships. FTS before vectors before hybrid. Scheduler before MCP (so MCP tools can report job status). Project generalization last (so we don't paint the schema into a corner before we know the shape).

---

## 6. Integration decisions

### Vanta Conduit — **embed as a Go library dependency**

**Use it for:** vector storage, embedding lifecycle, namespace partitioning, memory activation/decay, MCP tool patterns.

**Rationale:**
- Already a Go module (`github.com/hollis-labs/vanta-conduit`) using `modernc.org/sqlite` — same stack constraints.
- Proven embeddable via Nanite's `internal/memory/service.go`, which imports `conduitMemory "github.com/hollis-labs/vanta-conduit/memory"` and wraps `memory.Store` directly. No HTTP, no sidecar.
- First-class namespaces (`context_namespace_*`) directly map to Stack Explorer's "per-repo partitioning" requirement. Namespace convention: `app/stack-explorer/<repo-id>/{symbols,findings,audits}`.
- Memory lifecycle with status transitions (draft → reviewed → canonical) aligns with how audit findings mature.
- Writing to Vanta from within Stack Explorer's process means **one SQLite file, two coexisting schemas** — single backup story, one connection pool.

**Gaps to close ourselves:**
1. **Vanta has no BM25.** → Stack Explorer adds SQLite FTS5 virtual tables alongside Vanta's storage. The retrieval service layers on top, combining BM25 from FTS5 + vector cosine from Vanta + metadata filters from SE's own schema via reciprocal rank fusion.
2. **Vanta's embedder is OpenAI-only, hardcoded in `createEmbedder()`.** → Either (a) upstream a local-provider patch to Vanta (Ollama via HTTP), (b) bypass `createEmbedder` by calling the library API with pre-computed vectors and handling embedding in Stack Explorer's own code, or (c) accept OpenAI as the default with Ollama as a SE-side override. **Recommend (b) short-term, (a) medium-term** — it's a patch worth landing upstream and it benefits Nanite too.
3. **Vanta's chunking is opinionated** (sentence/paragraph/fixed). Stack Explorer wants symbol-granularity chunks, not text-length chunks. → Don't use `context_chunked_ingest`. Call `context_write` directly with pre-chunked symbol content. Vanta's chunker is a separate concern from its storage.

**Cost:** Stack Explorer adopts Vanta's schema migrations into its own DB file. Schema versions must not conflict — namespace migrations with a prefix or separate migration runner. Planning session needs to nail this down.

### Hadron — **keep as external peer, borrow patterns for in-process scheduler**

**Verdict from the survey:** Hadron is not practically embeddable today. `internal/` package restriction, tight coupling to fragments-engine OTel, required global init, and no published `pkg/` public API. Extracting a clean `pkg/scheduler` would require 2–4 weeks of Hadron refactoring — not worth it for our needs.

**What we do instead:**

1. **Keep Hadron as an external CLI for complex blueprint execution** — exactly as today. `scripts/batch-scan.sh`, `hadron run ...`, the `se-repo-scan`/`se-security-scan`/`se-feature-audit` blueprints stay. Stack Explorer shells out for these.

2. **Build a minimal in-process scheduler/runner in `internal/jobs/`** — ~500–800 LOC borrowing patterns from Hadron:
   - **Scheduler tick loop** (1-second ticker, fetch due jobs, dispatch). Pattern from Hadron's `internal/scheduler/engine.go:103-116`.
   - **Worker pool with per-job context cancellation.** Pattern from `internal/execution/manager.go:203-246`.
   - **Retry backoff** (fixed/exponential/linear with caps). Hadron has battle-tested versions in `internal/execution/retry_test.go`.
   - **Event emitter** (subscriber channels, non-blocking broadcast). Pattern from `internal/execution/manager.go:648-669`. Events also persist to `job_events` for durable history.
   - **Library:** `github.com/robfig/cron/v3` (same as Hadron). Don't duplicate cron parsing.

3. **Job kinds native to Stack Explorer** — symbol-reindex, embedding-refresh, git-history-sweep, audit-refresh — run in-process via this minimal scheduler. They're tightly coupled to SE's own tables and don't benefit from blueprint-style YAML indirection.

4. **`scan` job kind** — when the scheduler fires a scan job, it invokes `hadron run ...` as subprocess (matching current behavior) and tails the result. Hadron stays the blueprint execution engine; Stack Explorer owns the schedule and the result ingestion.

**Reconsider later:** If in-process scheduling and event streaming start mattering across both projects, prototype a minimal `pkg/scheduler` extracted from Hadron — but that's a v2 conversation, not a blocker for shipping.

### OSS library picks (verified via survey)

For each capability, one boring and durable pick. All chosen for low-churn profile.

| # | Capability | Pick | License | CGo? | Why |
|---|---|---|---|---|---|
| 1 | Multi-lang parsing (symbols) | Tree-sitter official Go binding | Apache-2.0 | **Yes** (only place) | Production-proven, full grammar ecosystem. Accept CGo for correctness here. |
| 1b | Cross-file reference graph | `sourcegraph/scip` | Apache-2.0 | No | Ingest protobuf from `scip-go`/`scip-typescript`/`scip-python`. Gives definitions + references free. Run SCIP indexers as subprocess. |
| 1c | Structural pattern matching | `ast-grep` CLI (shell-out) | MIT | No (external binary) | Declarative queries over tree-sitter ASTs. Use for "find every call-site matching X." |
| 2 | BM25 full-text search | SQLite FTS5 | Public domain | No | Stays in the same DB. Don't add Bleve/Bluge. |
| 3 | Vector search | Vanta Conduit (primary); sqlite-vec as fallback | Apache-2.0 / MIT | No | See above. Vanta owns the vector substrate. |
| 4 | Embedding generation | Ollama HTTP (`/api/embeddings`) | MIT | No | Local-first. No CGo. Users install Ollama once; we send JSON. Also wire upstream OpenAI as optional for users who want it. |
| 5 | Graph traversal | SQLite recursive CTEs + Go BFS | Public domain / stdlib | No | Single-repo scale. No graph DB. |
| 6 | Cyclomatic complexity (multi-lang) | Lizard (Python CLI, shell-out) | MIT | No | Only mature multi-language CC tool. Stable CSV output. Treat like we treat scc. |
| 7 | Code graph ingestion | `sourcegraph/scip` protobuf | Apache-2.0 | No | See #1b. |
| 8 | MCP server SDK | `github.com/modelcontextprotocol/go-sdk` | MIT | No | Official, Anthropic-maintained. Community SDKs predate it. |
| 9 | Cron scheduling | `github.com/robfig/cron/v3` | MIT | No | Same lib as Hadron. SQLite for persistence, in-memory for runtime. |
| 10 | Git inspection | `git` CLI (heavy ops) + `go-git` v5 (cheap ops) | Various / Apache-2.0 | No | Shell-out for blame/log across large histories (10–35× faster). go-git for refs, tree reads, clone-on-demand. |

**CGo budget:** exactly one place — tree-sitter. Everything else stays pure-Go or out-of-process. This is a deliberate constraint to preserve single-binary distribution.

**Worth calling out that we're *not* using:** Bleve/Bluge (FTS5 is enough), Cayley/Dgraph/Kuzu (no graph DB), Chroma/Qdrant/Milvus (Vanta is the vector layer), llama.cpp bindings (shell out to Ollama).

---

## 7. MCP and HTTP API — design principles

- **HTTP first, MCP second.** MCP tools are thin wrappers over HTTP handlers in most cases, so adding a new operation means adding an HTTP endpoint once and exposing it via MCP. Avoid divergent implementations.
- **Read-heavy.** Most MCP tools are queries. Write tools are scoped to audit authoring and finding status transitions. No destructive operations over MCP.
- **Namespace-aware.** Every query tool accepts an optional `repo_id` filter. Global queries return aggregated results; per-repo queries are cheap.
- **Pagination and limits.** Every list endpoint is paginated (`page`, `pageSize`) and every search endpoint accepts a `limit`. Default small — 25 for lists, 10 for searches.
- **Explicit provenance.** Every finding and audit records its source (agent ID, human user, importer run). Provenance is a first-class field, not a comment.
- **SSE for events.** Agents that need to wait on a long-running job subscribe to the event stream. One connection, many job types, filtered server-side.

---

## 8. Proposed delivery phases

Phases are shipping units. Each ends in something an agent could actually use.

### Phase A — Audits as data (1–2 weeks)

**Goal:** Nanite's 6 completed audits become queryable rows in Stack Explorer's DB.

- Migration 016 (`audits` + `findings` extension)
- Migration 018 (`audit_themes`, `finding_themes`)
- CLI: `audit start|finish|list|show|import|export|diff`
- `audit import <folder>` parser for the deep-review format
- Re-run the importer against Nanite; verify the schema holds up against real data
- HTTP endpoints for audits and extended findings

**Exit criterion:** `./stack-explorer audit import /path/to/nanite/docs/audits && ./stack-explorer audit list --repo nanite` returns real results.

### Phase B — Symbol substrate (1–2 weeks)

**Goal:** Stack Explorer knows what functions, methods, and types exist in its indexed repos.

- Migration 019 (`symbols`)
- Migration 021 (`code_references` symbol anchor columns)
- Ingestor in `internal/symbols/` — tree-sitter for Go first, then the others. SCIP ingestion as the cross-file reference pass.
- CLI: `symbol ingest <repo>`, `symbol show <name>`, `symbol search <query>`
- Backfill `code_references` anchors where possible (there are 0 today, so effectively a no-op on current data)

**Exit criterion:** `./stack-explorer symbol search "CreateRepo" --repo stack-explorer` returns the actual function with file:line.

### Phase C — Retrieval (1 week)

**Goal:** Hybrid search works against Phase A + B data.

- Migration 023 (FTS5 virtual tables + sync triggers)
- Migration 024 (Vanta Conduit integration — schema + library import)
- `internal/retrieval/` package (BM25 + vector + RRF merge)
- CLI: `search <query> --repo ... --kind finding|symbol|code-ref`
- Embedding generation via Ollama adapter (see §6 Vanta gap #2)

**Exit criterion:** `./stack-explorer search "panic recovery"` returns findings and symbols, ranked sensibly, within 200ms on Nanite-sized data.

### Phase D — MCP surface (1 week)

**Goal:** Any MCP-capable agent can pull Stack Explorer knowledge into a session.

- `internal/mcp/` package (official Go SDK)
- First 6 tools: `repo_context`, `prior_art_for_file`, `finding_search`, `symbol_lookup`, `audit_show`, `knowledge_query`
- Stdio and HTTP transport
- Integration test with Nanite consuming Stack Explorer over MCP

**Exit criterion:** Nanite in a session can call `stack-explorer.prior_art_for_file(path="internal/chat/engine.go")` and receive real findings before touching the code.

### Phase E — Scheduler and maintenance (1–2 weeks)

**Goal:** Stack Explorer keeps its own knowledge fresh without human intervention.

- Migration 026 (`schedules`, `jobs`, `job_events`)
- `internal/jobs/` — scheduler, worker pool, retry, event emitter (patterns borrowed from Hadron)
- Job kinds: `scan` (shells out to `hadron run`), `symbol-reindex`, `embedding-refresh`, `audit-refresh`, `git-history-sweep`
- SSE event endpoint + MCP `event_stream_subscribe` tool
- CLI: `schedule add|list|remove|trigger`

**Exit criterion:** A nightly schedule on a real repo runs, recomputes symbols and embeddings if git has changed, emits events, and updates timestamps in place.

### Phase F — Graph relationships (1 week, can run parallel to E)

**Goal:** Stack Explorer can answer "what else is connected to this symbol?"

- Migration 022 (`relationships`)
- Git-history sweeper job (populates `co-changed-with`)
- SCIP `calls`/`imports`/`references` ingestion
- CLI: `graph neighbors <symbol>`, `graph co-change <symbol>`
- HTTP + MCP surface

**Exit criterion:** Given a symbol, Stack Explorer returns its callers, callees, and co-changed siblings with a single query.

### Phase G — Project generalization (deferred)

**Goal:** Index arbitrary folders and binaries, not just git repos.

Only touch this once A–F have been in real use long enough to know what the schema should look like. Migration 030.

---

## 9. Open questions for the planning session

These are the real architectural decisions — each one has defensible options and we should walk through them together before committing.

1. **Vanta embedder swap: upstream or bypass?** Do we land an Ollama provider patch in Vanta itself (benefits Nanite too) or does Stack Explorer compute embeddings itself and call `context_write` with pre-computed vectors? What's the right interface contract between SE and Vanta here?

2. **Shared DB file for Vanta + SE schemas.** Single SQLite file with both sets of tables, or separate files? Migration coordination? Connection pool sharing? Any FK constraints across the boundary (probably: no — keep them logically separate via namespaces)?

3. **Symbol ingestor primary: tree-sitter or SCIP?** The survey recommends both, but which is the *primary* anchor for `symbols.id`? My lean: tree-sitter owns symbol identity, SCIP adds cross-file reference edges into `relationships`. Confirm or change.

4. **CGo budget.** Is tree-sitter acceptable as the one CGo dependency? If not, the fallback is `malivvan/tree-sitter` (wazero-based, pre-release) or SCIP-only ingestion (cedes some local symbol detail). What's the tolerance?

5. **Line-number staleness strategy.** Content-hash the enclosing symbol is the recommendation. Alternatives: blame re-anchoring, commit-SHA tagging with opportunistic repair. Which does the planning session actually want to implement first?

6. **Audit import format lock-in.** The deep-review INDEX.md + finding-NN-*.md format exists today. Do we commit to it as Stack Explorer's canonical audit authoring format, or design a more structured YAML frontmatter variant and migrate?

7. **Finding provenance fields.** What exactly do we store? `(actor_kind, actor_id, session_id, tool_name, model_name, timestamp)`? Worth nailing down before Phase A to avoid migration churn later.

8. **Scheduler fault tolerance.** What happens to a running job if Stack Explorer crashes? Mark `in_progress` jobs as `failed` on startup? Retry from where they left off (hard)? No retry (easy, correct-ish)? Hadron's model is "mark failed" — adopt it?

9. **Embedding refresh policy.** Refresh on every git change, on a schedule, on explicit demand, or only when content hash actually changes at the symbol level? The last is clearly most efficient but adds complexity to the job runner. Acceptable?

10. **MCP tool namespacing.** Do we prefix tool names with `stack_explorer__` to avoid collisions with other MCP servers in a session, or rely on the server-level namespace? Convention matters for agent UX.

11. **Storage size budget.** Per-repo, per-symbol embeddings can grow fast. For a 50k-symbol repo at 1536 dims float32, that's ~300MB just in vectors. Planning question: is that acceptable? Do we need a smaller embedding model, dimension reduction, or per-symbol content filters that skip low-value symbols (getters, trivial wrappers)?

12. **Evaluation and ground truth.** How will we know retrieval quality is actually good? A handful of canonical queries against Nanite's audit data with expected results baked into test cases? Recall@10 against a labeled set? Worth establishing before we start tuning.

---

## 10. Risks and constraints

- **Schema churn.** Getting data model right early is critical because migrations are one-way. The content-hash anchor and symbol-as-identity decisions are the two highest-consequence calls. Mistakes there ripple through everything.
- **Embedding lock-in.** Choosing a model means a world of vectors in that model's space. Swapping requires re-embedding everything. Pick a model we're willing to live with for a year and make the choice explicit, not accidental.
- **Vanta upstream pace.** If Stack Explorer needs Vanta features that Vanta doesn't prioritize (local embeddings, hybrid search hooks), we're on the hook for upstream patches or a fork. Plan for both; keep conversations with Vanta maintainers open.
- **Tree-sitter grammar quality variance.** Go and TypeScript grammars are excellent; Python is good; Rust has rough edges. Test with real repos before committing to a "supports all four" claim.
- **Scheduled jobs that fail silently.** The event stream + durable `job_events` is the mitigation. Every schedule run must produce either a success or a failure row; absence is a bug.
- **CGo single-point.** Tree-sitter is the one CGo dependency. A broken C toolchain breaks all builds. Distribution story: ship pre-built binaries, don't ask users to compile.
- **Audit data staleness.** An imported audit is correct at a commit; code moves. The `audit_diff` tool and `audited_at_ref` field are the mitigation, but we need a UX for "here's a finding whose anchored symbol has drifted" that doesn't either hide or overwhelm.

---

## 11. Success criteria (how we'll know this worked)

- **Nanite's 6 existing audits are real data** in Stack Explorer, queryable by file, symbol, theme, and severity.
- **An agent about to touch a file in Nanite can call `prior_art_for_file` over MCP** and get findings back before wasting context on rediscovery. End-to-end round trip under 200ms.
- **A re-run of an audit produces a diff** — resolved, still-open, new, regressed — not a replacement.
- **Cross-repo pattern queries work.** `theme show "missing panic recovery"` returns instances across every audited repo.
- **Stack Explorer runs as a daemon** with scheduled maintenance that users never have to touch. The event stream shows what's happening in real time.
- **New repo onboarding in under 10 minutes** from `repo add` to a working symbol search and embedding-backed query.
- **The same binary runs against git repos, local folders, and binaries** once Phase G lands.

---

## 12. Appendix — file map of proposed additions

```
internal/
├── audits/           # Phase A — audit authoring, import/export, diff
├── symbols/          # Phase B — tree-sitter + SCIP ingestors, symbol store
├── retrieval/        # Phase C — BM25, vector, hybrid RRF
├── vanta/            # Phase C — wrapper around vanta-conduit library
├── mcp/              # Phase D — official Go SDK, tool handlers
├── jobs/             # Phase E — scheduler, worker pool, events
├── graph/            # Phase F — relationships traversal helpers
├── git/              # Phase E/F — git CLI wrappers, co-change sweeper
└── api/              # Existing — extended with new endpoints each phase

cmd/stack-explorer/
└── (new subcommands per phase — audit, symbol, search, schedule, graph)

docs/
├── proposed-knowledge-base-architecture.md   # this file
├── product-vision.md                          # existing
└── adr/                                       # recommended: one ADR per major decision from §9
```

---

## 13. What the next planning session needs to decide

Ordered by urgency — early decisions block later ones.

1. **Phase A scope lock.** Commit to Migration 016 shape before writing code.
2. **Vanta integration model.** Shared DB vs separate. Embedder swap strategy.
3. **Symbol ingestor primary path.** Tree-sitter + SCIP combo, or one of them.
4. **CGo tolerance.** Yes/no on tree-sitter.
5. **Provenance field shape.** Lock before Phase A so we don't migrate again.
6. **Audit authoring format.** Commit to deep-review INDEX.md or design a variant.
7. **Retrieval quality eval harness.** Agree on a small labeled set to benchmark against.

Everything else in §9 can wait until its phase comes up. These seven can't.

---

*This document is a starting point for discussion. None of it is committed. Expect Phases D, E, and F to shift as we learn from Phases A–C.*

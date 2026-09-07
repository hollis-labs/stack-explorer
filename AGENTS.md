# Stack Explorer

Stack Explorer is a local research platform for the AI agent and developer-tooling
ecosystem. It keeps a SQLite catalog of repositories, scores them through weighted
lenses, extracts symbols and code relationships, and serves the result over a CLI,
a REST API and an MCP endpoint. It does not perform the scans itself — quantitative
analysis comes from Hadron blueprints — and it is not a task tracker, an agent
launcher, or a durable store for project knowledge.

## Start Here

- `cmd/stack-explorer/` is the CLI surface: one `cmd_<group>.go` per command group.
- `internal/store/sqlite/migrations/` is the schema source of truth — numbered,
  embedded SQL, applied on open.
- `internal/api/` is the Chi REST server (`serve`, port 8080); `server.go` owns
  routing, CORS, pagination and the sort whitelist.
- `internal/mcp/` is the MCP tool surface (`mcp --transport stdio|http`);
  `docs/mcp-tools.md` is the tool reference.
- `internal/retrieval/` fuses FTS5 BM25 with vector search over `internal/embed/`;
  `internal/symbols/` and `internal/graph/` produce what it retrieves.
- `docs/proposed-knowledge-base-architecture.md` is the layered plan the symbol,
  graph and retrieval work follows; `docs/audit-format.md` defines the deep-review
  ingest format.
- `blueprints/` holds the Hadron blueprints; `scripts/` holds the batch scan,
  ingest, seed and export shells.

## Commands

```bash
make build          # -> ./stack-explorer, pure Go
make test
make lint           # go vet ./...
make build-full     # adds the treesitter tag — the only CGo path here
```

`make eval` rewrites the tracked `eval/baseline.json`. Run it only when you intend
to move the retrieval baseline, and commit that diff deliberately.

## Boundaries

The default build is pure Go on `modernc.org/sqlite`, with no CGo. Tree-sitter
symbol extraction sits behind the `treesitter` build tag
(`internal/symbols/factory_full.go`), so `make test` neither compiles nor exercises
it — use `go test -tags treesitter ./internal/symbols/...` when you change it.

`data/stack-explorer.db` is gitignored, live, and hundreds of megabytes; nothing in
the repo regenerates it. Treat it as real data — CLI subcommands act on the actual
catalog, and `scripts/export-state.sh` is the supported way to move it.

`allowedSort()` in `internal/api/server.go` whitelists ORDER BY columns. Sort
parameters arrive from query strings, so every list handler routes them through it.

**The LongMemEval benchmark work is not on this branch.** It lives on
`feature/longmemeval-vanta`, usually checked out in a gitignored worktree under
`.worktrees/`, and it has diverged from `main` in both directions. Its
`benchmarks/longmemeval/` tree does not exist here, and the tracked
`docs/superpowers/{plans,specs}/2026-04-16-longmemeval-vanta*.md` describe that
branch rather than this one. Do not recreate that tree from those documents, and do
not read the worktree's own `.agentrc/`, `.nanite/` or `CLAUDE.md` as this repo's
configuration — they are an April snapshot of files `main` no longer carries.

`docs/stack-explorer-api-prompt.md` is a historical boot spec: the API it calls
unbuilt shipped in `internal/api/`. It, `docs/install-on-new-machine.md` and the
`local_path` entries in `repos.yaml` still cite `~/Projects-apps/`, a prefix that
no longer exists. A path named in a document is not evidence the path is there.

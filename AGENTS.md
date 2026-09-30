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
- `docs/product-vision.md` is the data model and direction; `docs/audit-format.md`
  defines the deep-review ingest format.
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

**The LongMemEval benchmark work is not on this branch.** It lives on the
`feature/longmemeval-vanta` branch, which has diverged from `main` in both
directions. Its `benchmarks/longmemeval/` tree does not exist here; do not
recreate it from planning documents that describe that branch.

The `local_path` entries in `repos.example.yaml` are examples; real checkout
paths are machine-specific and belong in your gitignored `repos.yaml`. A path
named in a document is not evidence the path is there.

`stack-explorer serve` binds `127.0.0.1` and refuses a non-loopback bind without
a bearer token (`internal/api/security.go`, guarded by `security_test.go`). Keep
that default; the MCP HTTP transport is unauthenticated and must stay on
loopback.

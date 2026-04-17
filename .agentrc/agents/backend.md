# Backend Context — Stack Explorer

> Project-specific backend conventions. Loaded by the backend agent role when working in this project.

## Stack

- **Go version:** 1.26+
- **Module path:** `github.com/chrispian/stack-explorer`
- **Router:** Chi (`github.com/go-chi/chi/v5`)
- **Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGo)
- **CLI framework:** Cobra (`github.com/spf13/cobra`)
- **Config format:** YAML (`gopkg.in/yaml.v3`)
- **Analysis tools:** scc (LoC/complexity), golangci-lint (Go quality)

## Project Structure

```
cmd/
└── stack-explorer/           # Cobra CLI commands
internal/
├── api/
│   ├── server.go             # Chi router, CORS, JSON response helpers, pagination
│   ├── handler_repos.go      # /api/repos CRUD
│   ├── handler_scores.go     # /api/scores CRUD
│   ├── handler_lenses.go     # /api/lenses CRUD
│   ├── handler_dimensions.go # /api/dimensions CRUD
│   ├── handler_patterns.go   # /api/patterns CRUD
│   ├── handler_findings.go   # /api/findings CRUD
│   └── handler_comparisons.go# /api/comparison-sets CRUD
├── store/
│   └── sqlite/               # SQLite store, 13 embedded SQL migrations, CRUD methods
├── domain/                   # Domain types (Repo, Score, Lens, Dimension, Pattern, Finding, etc.)
├── config/                   # App config loading
├── analysis/                 # Analysis runners
└── render/                   # Output renderers
data/
└── stack-explorer.db         # SQLite database (gitignored)
blueprints/                   # Hadron blueprints for automated scanning
scripts/                      # Batch scan, ingest, seed, cleanup scripts
reports/                      # Generated markdown scorecards
```

## API

- **Base path:** `/api`
- **CORS:** Allows `http://localhost:3334` (frontend dev server)
- **Content-Type:** All responses are `application/json`
- **Pagination:** `?page=1&pageSize=25&sort=name&direction=asc`
- **Filtering:** `?filter[field]=value` query params parsed by `parseListParams()`
- **List responses:** `{ data: [...], meta: { total, page, pageSize, pageCount } }`
- **Item responses:** `{ data: { ... } }`
- **Errors:** `{ error: "message" }` with appropriate HTTP status

### Routes

| Method | Path | Handler |
|--------|------|---------|
| GET/POST | `/api/repos` | list, create |
| GET/PUT/DELETE | `/api/repos/{id}` | get, update, delete |
| GET/POST | `/api/tags` | list, create |
| DELETE | `/api/tags/{id}` | delete |
| GET/POST | `/api/snapshots` | list, create |
| GET | `/api/snapshots/{id}` | get |
| GET/POST | `/api/dimensions` | list, create |
| GET | `/api/dimensions/{id}` | get |
| GET/POST | `/api/lenses` | list, create |
| GET/DELETE | `/api/lenses/{id}` | get, delete |
| GET/POST | `/api/scores` | list, create |
| PUT | `/api/scores/{id}` | update |
| GET | `/api/scorecards` | list |
| GET | `/api/scorecards/{id}` | get |
| GET/POST | `/api/patterns` | list, create |
| GET/PUT/DELETE | `/api/patterns/{id}` | get, update, delete |
| GET/POST | `/api/findings` | list, create |
| GET/PUT | `/api/findings/{id}` | get, update |
| GET/POST | `/api/comparison-sets` | list, create |
| GET/DELETE | `/api/comparison-sets/{id}` | get, delete |
| GET/POST | `/api/reports` | list, create |
| GET/DELETE | `/api/reports/{id}` | get, delete |

## Database

- 13 migrations in `internal/store/sqlite/migrations/` (embedded SQL, numbered)
- WAL mode + foreign keys enabled
- 18 review dimensions, 9 scoring lenses seeded on first run
- Run `./stack-explorer db stats` to check table counts

## Conventions

- Repo IDs are lowercase slugs matching the repo name
- URLs are HTTPS (not SSH) — used for clone-on-demand
- Scores are 0.0–10.0 with evidence text
- All timestamps are RFC3339 UTC
- Patterns are `pattern` or `anti_pattern` type
- Findings use categories: `gap`, `strength`, `opportunity`, `risk`
- Own projects have `is_own = true`
- `allowedSort()` validates ORDER BY columns against a whitelist — always use it

## Build & Run

- **Build:** `make build` → `./stack-explorer`
- **Install:** `make install` → `~/go/bin/stack-explorer`
- **Test:** `make test`
- **Lint:** `make lint`
- **Run API:** `./stack-explorer serve` (default port configurable)

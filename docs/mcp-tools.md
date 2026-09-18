# Stack Explorer MCP Tools

Stack Explorer exposes a read-heavy MCP surface for repo knowledge, findings, symbols, and audits.

## Server

- Stdio: `./stack-explorer mcp --transport stdio`
- HTTP: `./stack-explorer mcp --transport http --host 127.0.0.1 --port 8090 --path /mcp`

Tool names are intentionally unprefixed.

## Defaults

- List-style tools default to `page=1` and `page_size=25`.
- Search-style tools default to `limit=10`.
- Query tools accept optional `repo_id` filters when relevant.
- Write-tool provenance is derived from MCP context where available, with environment fallback:
  - `STACK_EXPLORER_MCP_ACTOR_KIND`
  - `STACK_EXPLORER_MCP_ACTOR_ID`
  - `STACK_EXPLORER_SESSION_ID`
  - `STACK_EXPLORER_MCP_MODEL`

## Read Tools

### `repo_context`

Returns repo metadata, top findings, top themes, recent audits, the latest snapshot, and recent activity.

Example input:

```json
{"repo_id":"nanite"}
```

### `prior_art_for_file`

Returns findings, themes, code references, and symbols touching a repo-relative file path.

Example input:

```json
{"repo_id":"nanite","path":"internal/chat/engine.go"}
```

### `prior_art_for_symbol`

Returns findings, themes, and code references linked to a symbol. Accepts either `symbol_id` or `repo_id` plus `qualified_name`.

Example input:

```json
{"repo_id":"nanite","qualified_name":"nanite.Engine.Run"}
```

### `finding_search`

Runs hybrid retrieval with findings as the default result kind.

Example input:

```json
{"repo_id":"nanite","query":"panic recovery","limit":5}
```

### `symbol_lookup`

Resolves names to symbols with file and line metadata.

Example input:

```json
{"repo_id":"nanite","query":"Engine.Run","limit":5}
```

### `audit_show`

Returns a stored audit plus findings and themes.

Example input:

```json
{"audit_id":1}
```

### `knowledge_query`

Runs freeform hybrid retrieval across findings, symbols, and code references.

Example input:

```json
{"repo_id":"nanite","query":"panic recovery","limit":5}
```

## Write Tools

### `finding_add`

Creates a finding with MCP-derived provenance.

Example input:

```json
{
  "repo_id":"nanite",
  "title":"Panic recovery is inconsistent",
  "category":"risk",
  "severity":"high",
  "status":"open",
  "description":"Nested tool execution does not consistently recover from panics."
}
```

### `finding_update_status`

Updates a finding lifecycle state. `status` must be one of the canonical
values — `open`, `acknowledged`, `resolved`, or `wontfix` — any other value is
rejected with an error listing the allowed set.

Example input:

```json
{"id":123,"status":"resolved"}
```

### `audit_import`

Imports one or more deep-review audits from a folder. Use `dry_run=true` to parse without writing.

Example input:

```json
{"folder":"/path/to/audits","repo_id":"nanite","dry_run":true}
```

### `audit_export`

Exports a stored audit to deep-review markdown.

Example input:

```json
{"audit_id":1,"out_dir":"/tmp/audit-1-export"}
```

## Notes

- The server is built on `github.com/hollis-labs/go-mcp` (`server` + `transport/http`), the portfolio's shared MCP library on top of the official `modelcontextprotocol/go-sdk`. The HTTP transport is stateless (2026-07-28 / SEP-2567): no `Mcp-Session-Id` is read or set, each request is served independently.
- Write-tool provenance (`actor_kind`, `actor_id`, `session_id`, `model_name`) is derived from environment variables, overridden per-call by the `X-Stack-Explorer-Actor-Kind` / `X-Stack-Explorer-Actor-Id` / `X-Stack-Explorer-Session-Id` / `X-Stack-Explorer-Model` headers on an HTTP call. go-mcp's tool handlers see only `(ctx, args)`, so these headers are threaded through request context ahead of the protocol layer (`internal/mcp/provenance_http.go`) rather than read off the request/session object.

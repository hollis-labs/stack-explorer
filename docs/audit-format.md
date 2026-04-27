# Audit Format

Stack Explorer supports two audit lanes during the Phase A migration window:

- Existing deep-review markdown folders are imported as-is.
- New audits should converge on a YAML-frontmatter format so authorship is deterministic and machine-validated.

For legacy deep-review imports, `model_name` is not present in the markdown corpus today. Importers must supply it explicitly via the caller provenance:

- CLI: `audit import ... --model-name <name>` takes precedence.
- CLI fallback: `STACK_EXPLORER_MODEL=<name>`.
- MCP / scheduler: the caller's provenance `model_name`.

If no model is available, Stack Explorer stores `''` for `model_name` rather than `NULL`.

## Folder shape

One audit lives in one directory:

```text
<audit-dir>/
  index.md
  01-critical-example.md
  02-medium-example.md
```

`index.md` is the audit manifest. Each finding file is one finding.

## Audit Manifest

The manifest keeps a markdown body but starts with YAML frontmatter:

```yaml
---
repo_id: nanite
scope: chat-engine
scope_paths:
  - internal/chat/**
audit_type: deep-review
auditor: nanite-reviewer-backend
audited_at_ref: 9c8d8e7
verdict: advisory
status: completed
actor_kind: agent
actor_id: nanite-reviewer-backend
session_id: session-20260426-1234
tool_name: deep-review
model_name: gpt-5.5
started_at: 2026-04-10T00:00:00Z
finished_at: 2026-04-10T01:15:00Z
themes:
  - Security
  - Concurrency correctness
---
```

Required fields:

- `repo_id`
- `scope`
- `audit_type`
- `auditor`
- `status`
- `actor_kind`
- `actor_id`
- `started_at`

Optional fields:

- `scope_paths`
- `audited_at_ref`
- `verdict`
- `session_id`
- `tool_name`
- `model_name`
- `finished_at`
- `themes`

Allowed enums:

- `status`: `in_progress`, `completed`, `superseded`
- `verdict`: `approve`, `approve-with-changes`, `reject`, `advisory`
- `actor_kind`: `agent`, `human`, `importer`

## Finding Files

Each finding file starts with YAML frontmatter:

```yaml
---
id: 01
title: ScopeGuard is dead code
severity: critical
category: risk
scope: chat-engine
audited_at_ref: 9c8d8e7
is_out_of_scope: false
actor_kind: agent
actor_id: nanite-reviewer-backend
session_id: session-20260426-1234
tool_name: deep-review
model_name: gpt-5.5
themes:
  - Security
code_refs:
  - file_path: internal/chat/engine.go
    line_start: 120
    line_end: 133
    ref_type: audit
---
```

Required fields:

- `title`
- `severity`
- `category`
- `actor_kind`
- `actor_id`

Optional fields:

- `id`
- `scope`
- `audited_at_ref`
- `is_out_of_scope`
- `session_id`
- `tool_name`
- `model_name`
- `themes`
- `code_refs`

Allowed enums:

- `severity`: `critical`, `high`, `medium`, `low`, `info`
- `category`: `gap`, `strength`, `opportunity`, `risk`

The markdown body after frontmatter is the full finding narrative and is stored verbatim in `findings.body_markdown`.

## Validation Rules

- One directory must map to one audit.
- Every finding file in the directory must parse cleanly or import fails.
- `themes` are deduped per repo by `(repo_id, name)`.
- `code_refs.file_path` is required when a `code_refs` item exists.
- `finished_at` must be greater than or equal to `started_at`.
- `audited_at_ref` should be a git commit SHA when available. Existing deep-review imports may leave it empty.

## Migration Path

- Deep-review markdown remains supported for Nanite’s legacy audits.
- New audits should prefer the YAML-frontmatter shape above.
- A future conversion tool can rewrite imported deep-review folders into YAML-frontmatter without changing stored rows.

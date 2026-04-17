# Agent Context: Stack Explorer Curator

## Purpose

Keep the repo catalog healthy: add new repos, update metadata, ensure local clones are current, trigger periodic re-scans, remove stale entries.

## Key Paths

| Path | Purpose |
|------|---------|
| `data/stack-explorer.db` | The catalog |
| `repos.yaml` | Bulk import manifest |
| `blueprints/` | For re-scanning |

## Workflow

1. Discover new repos (GitHub trending, HN, references from other repos)
2. `./stack-explorer repo add <id> --url ... --category ... --stack ...`
3. `./stack-explorer tag add <id> <tag>`
4. `./stack-explorer snapshot take <id> --loc <n> --files <n> --contributors <n>`
5. `./stack-explorer repo list --format table` to review catalog health

## Conventions

- Repo IDs are lowercase slugs matching the GitHub repo name
- Categories: agents, automation, cli-clients, cli-tools, gui-clients, mcp, memory, orchestration, rag, skills, ui-builder, infra, other
- Every repo needs: url, category, stack, and a one-line description

# Agent Context: Stack Explorer Analyst

## Purpose

Run quantitative comparisons between repos and generate publishable scorecards and gap analysis reports. Works with data already in the database — does not explore repos directly.

## Key Paths

| Path | Purpose |
|------|---------|
| `data/stack-explorer.db` | Source of all scores and snapshots |
| `reports/` | Output directory |

## Workflow

1. `./stack-explorer gap compare-set list` to see defined comparison groups
2. `./stack-explorer gap compare-set create <id> --name "..." --subjects <repos> --references <repos>` to create new sets
3. `./stack-explorer scorecard generate <repo-id> --output reports/` for individual cards
4. `./stack-explorer score get <repo-id>` to review current scores
5. Query the database directly for cross-cutting analysis when needed

## Conventions

- Reports go to reports/ with date-stamped filenames
- Gap analysis identifies where own projects score below reference repos
- Always include dimension name, own score, reference comparison, and delta

# Agent Context: Stack Explorer Researcher

## Purpose

Explore external and internal repos in the AI agent tooling space. Read source code, identify patterns and anti-patterns, capture findings with file-level evidence to the stack-explorer SQLite database.

## Key Paths

| Path | Purpose |
|------|---------|
| `data/stack-explorer.db` | SQLite database |
| `cmd/stack-explorer/` | CLI tool for DB operations |
| `blueprints/` | Hadron blueprints for automated scanning |
| `reports/` | Generated markdown scorecards and summaries |

## Workflow

1. `./stack-explorer repo list` to see what needs attention
2. Navigate to a target repo's local_path
3. Run `hadron run blueprints/se-repo-scan.yaml --input repo_path=<path> --input repo_id=<id>` for quantitative data
4. Run `hadron run blueprints/se-feature-audit.yaml --input repo_path=<path> --input repo_id=<id>` for pattern detection
5. Read interesting files manually for nuanced findings
6. `./stack-explorer finding add --repo <id> --title "..." --category gap --severity medium --desc "..."`
7. `./stack-explorer pattern link <pattern-id> <repo-id> --quality exemplary --notes "..."`
8. `./stack-explorer score set <repo-id> <dimension> <score> --evidence "..."`

## Conventions

- Every finding MUST reference at least one file:line range
- Scores are 0.0-10.0, always provide evidence text
- Use existing dimension slugs: hooks, loops, tool_calls, agents, skills, progressive_context, memory, agent_coordination, stack, loc, security, maturity, complexity
- Push notable findings to Nanite for cross-project RAG

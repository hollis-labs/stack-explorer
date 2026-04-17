# Run Analysis

Execute analysis blueprints against a repo.

## When to use

When you need to collect quantitative metrics and feature audit data for a repo.

## Procedure

1. Look up repo: `cd /Users/chrispian/Projects-apps/stack-explorer && ./stack-explorer repo show <id>` to get local_path
2. If no local_path, check if the repo needs cloning first
3. Run selected blueprints via Hadron:
   - **Scan:** `hadron run blueprints/se-repo-scan.yaml --input repo_path=<path> --input repo_id=<id>`
   - **Security:** `hadron run blueprints/se-security-scan.yaml --input repo_path=<path> --input repo_id=<id>`
   - **Features:** `hadron run blueprints/se-feature-audit.yaml --input repo_path=<path> --input repo_id=<id>`
4. After scan, capture snapshot: `./stack-explorer snapshot take <id> --loc <n> --files <n> --contributors <n>`
5. Report which tools ran and any failures

## Invariants

- Never run analysis on a repo without a valid local_path
- Always take a snapshot after scan completes

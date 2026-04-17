# Onboard Repo

Add a new repo to stack-explorer and run initial analysis.

## When to use

When you discover a new repo to track and want it fully integrated.

## Procedure

1. Add the repo:
   ```
   cd /Users/chrispian/Projects-apps/stack-explorer
   ./stack-explorer repo add <id> \
     --url <https-url> \
     --category <category> \
     --stack <language> \
     --desc "One-line description"
   ```
2. Tag it: `./stack-explorer tag add <id> <tag1> <tag2> ...`
3. Run the scan (clones to tmp, auto-cleans):
   ```
   ./scripts/batch-scan.sh --category <category>
   ```
   Or manually:
   ```
   git clone --depth 1 <url> tmp/repos/<id>
   hadron run /path/to/blueprints/se-feature-audit.yaml --input repo_path=tmp/repos/<id> --input repo_id=<id>
   hadron run /path/to/blueprints/se-repo-scan.yaml --input repo_path=tmp/repos/<id> --input repo_id=<id>
   ./scripts/cleanup.sh <id>
   ```
4. Ingest scan results: `./scripts/ingest-scans.sh <id>`
5. Score across relevant dimensions: `./stack-explorer score set <id> <dim> <score> --evidence "..."`
6. Add to relevant comparison sets if applicable
7. Export backup: `./stack-explorer repo export repos-backup.yaml`

## Categories

agents, automation, cli-clients, cli-tools, gui-clients, infra, mcp, memory, orchestration, rag, skills, ui-builder, other

## Invariants

- Always use HTTPS URLs (not SSH)
- Repo ID should be lowercase slug matching the repo name
- Run scan before scoring — scan data informs scores

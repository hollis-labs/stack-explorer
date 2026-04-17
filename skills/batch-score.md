# Batch Score

Score multiple repos across dimensions efficiently.

## When to use

When you need to score a category of repos or re-score repos after adding new dimensions.

## Procedure

1. List repos to score: `./stack-explorer repo list --category <cat>`
2. For each repo, review its scan data: `cat data/scans/<repo-id>-*.json | python3 -m json.tool | head -50`
3. Run the feature audit blueprint for context:
   ```
   hadron run /Users/chrispian/Projects-apps/stack-explorer/blueprints/se-feature-audit.yaml \
     --input repo_path=<path> --input repo_id=<id>
   ```
4. Score each dimension:
   ```
   ./stack-explorer score set <repo-id> <dimension> <score> --evidence "..."
   ```
5. Optionally recalculate through a specific lens:
   ```
   ./stack-explorer score recalc <repo-id> --lens <lens-id>
   ```

## Scoring Scale

- 0: Not present / not applicable
- 1-2: Minimal / barely present
- 3-4: Basic implementation
- 5-6: Solid implementation
- 7-8: Strong implementation with good coverage
- 9-10: Best-in-class / exemplary

## Invariants

- Always provide evidence text with scores
- Score 0 explicitly when a dimension doesn't apply (don't skip it)
- Run the feature audit first for objective signal

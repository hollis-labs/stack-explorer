# Gap Analysis

Compare own projects against external repos on specific dimensions.

## When to use

When you need to identify where own projects fall behind or lead compared to external repos.

## Procedure

1. Create a comparison set if needed:
   `cd /Users/chrispian/Projects-apps/stack-explorer && ./stack-explorer gap compare-set create <id> --name "..." --subjects <own-repos> --references <ext-repos>`
2. Query scores for all repos in the set using `./stack-explorer score get <repo-id>` for each
3. For each dimension, compute:
   - Median score of reference repos
   - Each subject repo's score
   - Delta (subject - reference median)
4. Rank gaps by severity (largest negative deltas first)
5. For critical gaps (delta <= -3.0), create findings:
   `./stack-explorer finding add --repo <subject> --title "Gap: <dimension>" --category gap --severity high`
6. Generate a markdown report summarizing the analysis

## Invariants

- Both subject and reference repos must have scored dimensions
- Never generate a gap report if fewer than 2 reference repos have scores

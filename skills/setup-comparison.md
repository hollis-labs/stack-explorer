# Setup Comparison

Create a comparison set for gap analysis between own projects and reference repos.

## When to use

When you need to define a new gap analysis comparing own projects against a category of external tools.

## Procedure

1. Identify subject repos (own projects to evaluate) and reference repos (external tools to compare against)
2. Choose a lens that matches the comparison context (run `./stack-explorer lens list` to see options)
3. Create the comparison set:
   ```
   cd /Users/chrispian/Projects-apps/stack-explorer
   ./stack-explorer gap compare-set create <id> \
     --name "Descriptive Name" \
     --desc "What this comparison evaluates" \
     --subjects repo1,repo2 \
     --references ref1,ref2,ref3
   ```
4. Verify: `./stack-explorer gap compare-set list`
5. Ensure all repos in the set have scores: `./stack-explorer score get <repo-id>`

## Invariants

- Both subjects and references need scored dimensions before gap analysis is meaningful
- Use at least 2 reference repos for meaningful averages
- Reference repos should be in the same domain as the subjects

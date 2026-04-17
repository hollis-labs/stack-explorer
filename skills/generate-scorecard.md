# Generate Scorecard

Pull scored data from the DB and render a markdown scorecard.

## When to use

When you need to produce a publishable scorecard for one or more repos.

## Procedure

1. Run: `cd /Users/chrispian/Projects-apps/stack-explorer && ./stack-explorer scorecard generate <repo-id> --output reports/`
2. Read the generated file and display it
3. Emit: `Scorecard: reports/scorecard-<repo-id>.md`

## Invariants

- If a repo has no scores yet, warn and suggest running analysis first

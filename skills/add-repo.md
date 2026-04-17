# Add Repo

Add a new repo to the stack-explorer database with metadata.

## When to use

When you need to register a new external or own repo in the catalog.

## Procedure

1. Parse the repo ID and URL from arguments
2. Determine category and stack from context or flags
3. Run: `cd /Users/chrispian/Projects-apps/stack-explorer && ./stack-explorer repo add <id> --url <url> --category <cat> --stack <stack> [--own] [--desc "..."]`
4. Tag the repo: `./stack-explorer tag add <id> <tag1> <tag2> ...`
5. Emit: `Added repo: <id> (<category>/<stack>)`

## Invariants

- Never add a repo without at least URL and category
- Always verify the URL looks valid before adding

# Add Dimension

Add a new review dimension to the system.

## When to use

When the existing 18 dimensions don't cover an evaluation criteria you need (e.g., "accessibility", "i18n", "offline-support").

## Procedure

1. Choose a slug ID: lowercase, underscores (e.g., `offline_support`)
2. Pick a category: `agent_patterns`, `code_quality`, `project_health`, or `features`
3. Run:
   ```
   cd /Users/chrispian/Projects-apps/stack-explorer
   ./stack-explorer dimension add <id> \
     --name "Offline Support" \
     --category features \
     --weight 1.0
   ```
4. Add the new dimension to relevant lenses:
   ```
   ./stack-explorer lens add-dim <lens-id> <dimension-id> <weight>
   ```
5. Verify: `./stack-explorer dimension list`

## Invariants

- Dimension IDs must be unique
- New dimensions have no scores until explicitly set via `score set`
- Remember to add the dimension to at least one lens or it won't affect any overall scores

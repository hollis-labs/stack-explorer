# Create Lens

Create a custom scoring lens with weighted dimensions.

## When to use

When you need to define a new perspective for scoring repos (e.g., a new app category or evaluation criteria).

## Procedure

1. Choose a slug ID (lowercase, hyphens): e.g., `my-lens`
2. Decide which dimensions apply and their weights (0.5 = low importance, 1.0 = normal, 1.5 = high, 2.0 = critical)
3. Run:
   ```
   cd /Users/chrispian/Projects-apps/stack-explorer
   ./stack-explorer lens create <id> \
     --name "My Lens" \
     --desc "Description of what this lens evaluates" \
     --dimensions hooks:1.0,memory:2.0,security:1.5,maturity:1.0
   ```
4. Verify: `./stack-explorer lens show <id>`
5. Use it: `./stack-explorer score get <repo-id> --lens <id>`

## Available Dimensions

Agent patterns: hooks, loops, tool_calls, agents, skills, progressive_context, memory, agent_coordination
Code quality: stack, loc, complexity
Project health: security, maturity, enterprise, data_governance
Features: multi_provider, multi_modal, ui_extensibility

## Invariants

- Lens IDs must be unique lowercase slugs
- At least one dimension required
- Weights should be 0.5-2.0 (higher = more influence on overall score)

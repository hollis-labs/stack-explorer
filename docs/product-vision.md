# Stack Explorer — Product Vision

*Living document | Last updated: 2026-04-03*

## What It Is

A multi-audience analytics platform for understanding the AI agent and developer tooling landscape. Tracks, scores, compares, and surfaces insights across repos, frameworks, and ecosystems.

## Core Concepts

### Data Model

```
Repos → Snapshots (time-series metrics)
     → Scores (per-dimension, per-lens)
     → Patterns (architecture patterns/anti-patterns)
     → Findings (observations, gaps, opportunities)
     → Tags (faceted classification)

Lenses → Dimensions + Weights (scoring perspectives)
Reports → Lens + Repo Filters (curated views)
```

### Three Consumer Tracks

| Track | Cares About | Example Views |
|-------|-------------|---------------|
| **Engineer** | Architecture patterns, code quality, complexity, dependency health, "how did they build X?" | Dimension breakdowns, code references, pattern examples |
| **Product** | Feature gaps, competitive positioning, market trends, "what should we build next?" | Gap analysis, feature matrices, trend charts |
| **Leadership** | Market landscape, investment signals, risk, "where is the market going?" | Heatmaps, rising/falling trends, ecosystem health |

### Scoring Lenses (current)

Lenses reweight the same raw scores for different perspectives. A repo scored once can be viewed through any lens:

- agent-platform, chat-app, automation, memory-system, agent-framework
- desktop-app, infra-tool, content-pipeline, general

### Report Configs (current)

Reports bind a lens to a filtered set of repos. Filters can be by category, tag, stack, explicit repo ID, or is_own. Filters compose as union (match any).

## Near-Term Roadmap

### Data Enrichment
- **GitHub API integration** — Live stars, forks, issues, contributors, release history (go-github already in cerberus)
- **CVE/vulnerability tracking** — Link repos to known CVEs via dependency analysis
- **Release timeline** — Track release cadence, breaking changes, semver compliance
- **Contributor analysis** — Bus factor, contributor diversity, commit velocity by author

### Time-Series & Trends
- **Popularity heatmaps** — Stars/forks over time, rising vs falling repos
- **Activity trends** — Commit velocity, issue response time, PR merge rate
- **Growth signals** — LoC growth rate, dependency adoption, community growth
- **Decay signals** — Stale repos, declining activity, abandoned projects

### Cross-Reference Data
- **Stack/framework mapping** — TALL stack, MEAN stack, JAMstack — filter by stack combination
- **Job market correlation** — Cross-reference stack popularity with job demand data (LinkedIn, Indeed APIs)
- **Ecosystem graphs** — Which repos depend on which? Who uses what SDK?

### Faceted Search & Filtering
- **Primary stack/language** — "Show me all Go agent frameworks"
- **Framework** — "Everything using Next.js" or "All Wails apps"
- **Stack combination** — "Go backend + React frontend" (TALL, etc.)
- **Feature flags** — "Has MCP" + "Has RAG" + "Multi-provider"
- **Audience** — Filter reports by engineer/product/leadership track

## Long-Term Features

### GUI (Sigil-based)
- Dashboard with lens-switchable leaderboards
- Interactive gap analysis with drill-down
- Repo detail pages with dimension radar charts
- Trend charts (stars, LoC, activity over time)
- Side-by-side comparison tool
- Pattern catalog with code examples

### Publishing
- Static site generation from reports/ (or Sigil-rendered)
- RSS/newsletter for weekly landscape updates
- API endpoint for programmatic access
- Embeddable widgets (scorecard badges, leaderboard tables)

### Automation
- Scheduled re-scans via Hadron cron
- GitHub webhook triggers on new releases
- Auto-discovery of trending repos (GitHub trending API)
- Alert on significant score changes (repo improved/degraded)

## Architecture Principles

1. **SQLite is the source of truth.** All data lives in the DB. Reports are generated views.
2. **Lenses don't change data.** Same raw scores, different weights. No re-scoring needed.
3. **Filters compose.** Reports are lens + filters. Filters are additive (union).
4. **Clone-on-demand.** Repos are cloned to tmp/ for analysis, cleaned after. URLs are canonical.
5. **Snapshot everything.** Time-series by default. Track changes over time, not just current state.
6. **Dimensions are extensible.** Users add custom dimensions and lenses for their context.
7. **Reports are reproducible.** Report config + DB state = deterministic output.

## Non-Goals (for now)

- Real-time monitoring (batch analysis is fine)
- Hosted SaaS (local-first CLI tool, static site for publishing)
- Automated scoring (agent-assisted but human-reviewed)
- Social features (comments, upvotes — this is a research tool)

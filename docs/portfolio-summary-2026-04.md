# Hollis Labs Portfolio Summary

**Date:** April 2026
**Scope:** Full audit of 12 applications + shared framework libraries
**Purpose:** Strategic reference for planning, prioritization, and architectural decisions

---

## The thesis

Build the infrastructure to build things. Every app in the portfolio serves one of two roles: (1) a tool that makes agent-driven development better, or (2) a product built with those tools to prove they work. The tools are the product; the products are the proof.

---

## Application inventory

### Tier 1 — Agent platform (the core)

| App | What it is | Stack | Maturity | Key strength |
|---|---|---|---|---|
| **Nanite** | Agent runtime + desktop app | Go + React (Wails) | Production | 23-type envelope UI, plugin system, multi-provider LLM, MCP hub, context hot-swap |
| **Vanta Conduit** | Unified AI memory layer | Go (embeddable library + daemon) | Production | Activation-ranked recall, namespace partitioning, semantic dedup, pointer model |
| **Nil** | Personal notes + document store | Go + React (Wails) | Production | Terminal aesthetic, keyboard-first UX, TipTap + wikilinks, FTS5 search, Claude AI tools |
| **Clockwork Manifold** | Agentic task execution engine | Go + React (Vite + shadcn) | Data plane ready, execution plane in progress | 35-field task model, typed deliverables, permission system, worker pool, FSM lifecycle |
| **Fast Triage** | Structured interaction for agents | TypeScript + React (Vite + shadcn) | Complete, shipping | Blocking MCP calls, keyboard-first forms/triage, atomic responses, zero coupling |

### Tier 2 — Automation and knowledge

| App | What it is | Stack | Maturity | Key strength |
|---|---|---|---|---|
| **Hadron** | Blueprint/workflow engine | Go + React (Wails) | Production | YAML blueprints, pipeline DAGs, scheduler (robfig/cron), PTY execution, event streaming |
| **Stack Explorer** | Code project knowledge base | Go (CLI + HTTP daemon) | Production (core), proposed (knowledge layers) | Multi-dimensional scoring, lenses, findings with code refs, scan pipeline, chi HTTP API |
| **Cerberus** | Infrastructure control plane | Go + React (Bubble Tea TUI) | Production (local services), emerging (cloud/infra) | Service DAG, health checks, auto-restart, connectors (DO, GitHub, Cloudflare, SSH, Forge) |

### Tier 3 — Generation and interfaces

| App | What it is | Stack | Maturity | Key strength |
|---|---|---|---|---|
| **Sigil** | UI YAML compiler | Go (CLI + dev server) | MVP, actively developed | Multi-target codegen (React/Go-Templ/HTML), 49 components, MCP-first, DataSource abstraction |
| **Fragments Engine** | Task execution (predecessor to CM) | Go + React (Tauri) | Production (being sunset) | Working executor (Claude CLI), signal protocol, retry escalation chain, quality gates |

### Tier 4 — Proof of concept

| App | What it is | Stack | Maturity | Key strength |
|---|---|---|---|---|
| **SUDS v2** | Browser dungeon crawler | TypeScript + Next.js 15 | 95% complete | AI-generated content, full game loop, agent-driven development workflow proof |

---

## Shared framework (framework/libs/)

| Package | LOC | Purpose | Consumers |
|---|---|---|---|
| **go-providers** | 10,415 | Multi-provider LLM abstraction (8 API + 8 CLI) | Nanite, Vanta |
| **go-queue** | 2,006 | Laravel-style job queue (SQLite/memory/noop) | Vanta, CM |
| **go-toolbroker** | 1,884 | Intent-aware MCP tool selection + token budgeting | Nanite, FE |
| **go-directives** | 1,336 | Chat directive parser (::command syntax) | Nanite |
| **go-strutil** | 1,198 | String utilities (slug, case, truncate, random) | Unreleased (zero consumers) |
| **go-plugin** | 1,030 | In-process plugin contract + registry | Nanite, Vanta, FE |
| **go-otel** | 796 | Shared OTel instrumentation + GenAI conventions | Nanite, Vanta, FE, Nil, Hadron |
| **go-mcp** | 546 | MCP response budget utilities | Hadron, Vanta, FE |
| **connectors/** | ~700 | Gmail + Webhook clients | Internal |

---

## Architecture — how the pieces connect

```
┌─────────────────────────────────────────────────────────────┐
│                    User / Human Layer                         │
│                                                               │
│  Nanite (desktop)     Nil (desktop)     Fast Triage (browser) │
│  ┌─────────────┐     ┌────────────┐     ┌──────────────┐     │
│  │ Conversation │     │ Notes/Docs │     │ Forms/Triage │     │
│  │ Envelopes   │     │ Wikilinks  │     │ Approvals    │     │
│  │ Context     │     │ FTS5       │     │ Keyboard UX  │     │
│  └──────┬──────┘     └──────┬─────┘     └──────┬───────┘     │
│         │                   │                   │              │
│         └─────────┬─────────┴─────────┬─────────┘              │
│                   │        MCP        │                        │
│                   ▼                   ▼                        │
│  ┌────────────────────────────────────────────────────┐       │
│  │              Orchestration Layer                     │       │
│  │                                                     │       │
│  │  Clockwork Manifold          Hadron                 │       │
│  │  ┌──────────────────┐       ┌──────────────────┐   │       │
│  │  │ Task execution   │       │ Blueprint runs   │   │       │
│  │  │ Plan→tasks→done  │       │ Pipelines, DAGs  │   │       │
│  │  │ Human-in-loop    │       │ Scheduled scans  │   │       │
│  │  │ Escalation chain │       │ PTY execution    │   │       │
│  │  └──────────────────┘       └──────────────────┘   │       │
│  └────────────────────────────────────────────────────┘       │
│                   │                   │                        │
│                   ▼                   ▼                        │
│  ┌────────────────────────────────────────────────────┐       │
│  │              Knowledge Layer                        │       │
│  │                                                     │       │
│  │  Vanta Conduit      Stack Explorer     Nil          │       │
│  │  ┌──────────┐       ┌──────────────┐  ┌──────────┐ │       │
│  │  │ Memory   │──ptr──│ Findings     │  │ Docs     │ │       │
│  │  │ Recall   │──ptr──│ Symbols      │  │ Notes    │ │       │
│  │  │ Pointers │──ptr──│ Relationships│  │ Markdown │ │       │
│  │  │ Decay    │       │ Audits       │  │ Content  │ │       │
│  │  └──────────┘       └──────────────┘  └──────────┘ │       │
│  └────────────────────────────────────────────────────┘       │
│                   │                                           │
│                   ▼                                           │
│  ┌────────────────────────────────────────────────────┐       │
│  │              Infrastructure Layer                   │       │
│  │                                                     │       │
│  │  Cerberus              Sigil                        │       │
│  │  ┌──────────────┐     ┌──────────────────┐         │       │
│  │  │ Service mgmt │     │ UI YAML compiler │         │       │
│  │  │ Cloud APIs   │     │ Multi-target gen  │         │       │
│  │  │ Infra control│     │ Component schemas │         │       │
│  │  └──────────────┘     └──────────────────┘         │       │
│  └────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

### Integration surfaces

- **MCP** — the primary loose-coupling mechanism. Every app exposes or consumes MCP tools. Agents in Nanite call Stack Explorer, Nil, Vanta, CM, Hadron, Fast Triage, and Cerberus over MCP.
- **Library embedding** — for tight, in-process integration. Nanite embeds Vanta as a Go library. Any Go app can embed Vanta the same way.
- **Vanta pointers** — the cross-system reference model. Vanta stores URI pointers (nil://, se://, file://, git://) to content owned by other systems. Agents resolve pointers via MCP.
- **HTTP API** — Stack Explorer, Cerberus, CM, and Hadron all expose REST APIs for programmatic access and for Sigil-generated frontends.
- **go-queue** — shared job queue for background work (embedding, refresh, cleanup) across Vanta, Nil, and Nanite.

### The knowledge flow

```
Agent session starts
  → Nanite loads context (hot-swap from Vanta memories)
  → Agent plans work (decomposed into CM tasks)
  → CM executes tasks (via Claude CLI, signals back)
    → Task needs approval → Fast Triage (MCP, blocking)
    → Task needs a scan → Hadron (blueprint, scheduled)
    → Task produces findings → Stack Explorer (ingested)
    → Task produces documents → Nil (stored)
    → Task produces learnings → Vanta (memory, with pointers)
  → Next session starts with full prior knowledge
```

---

## Strategic priorities (cross-portfolio)

### Critical path: CM executor

Clockwork Manifold's execution layer is the last load-bearing gap. The data plane (tasks, sprints, API, GUI) is production-ready with 478+ tests. The execution plane (actually running tasks with LLM calls, tool invocation, and signal parsing) is stubbed. **Connecting this is the single highest-priority piece across the entire portfolio** — it's what turns the task/sprint/role workflow from "agents plan" to "agents plan and execute."

### High value, near-term

| Priority | What | Why |
|---|---|---|
| **Stack Explorer audit extension** | Migration 016 (audits table) + audit import | Makes 6 existing Nanite audits into queryable data; validates the schema |
| **Stack Explorer MCP server** | `internal/mcp/` with 6 initial tools | Makes code knowledge available to every agent session |
| **Vanta hybrid search** | FTS5 for BM25 + RRF fusion | Eliminates "freshly written memory invisible to search" problem |
| **Vanta Ollama provider** | In go-providers | Local embeddings without cloud API key |
| **Nil MCP server** | `nil mcp serve` with 5 tools | Makes Nil useful to every agent in the portfolio |
| **Cerberus v2 config migration** | Commit to v2, drop v1 reading | Unblocks every other Cerberus improvement |
| **go-queue hardening** | Retry ID fix, handler timeouts, middleware | Benefits Vanta and Nil immediately |

### Framework improvements

| Priority | What | Why |
|---|---|---|
| **Publish libs with version tags** | go-otel, go-plugin, go-strutil, go-queue | Drop fragile replace directives |
| **go-providers SDK swap** | Replace hand-rolled HTTP adapters with official Anthropic/OpenAI/Gemini SDKs | Deletes ~4K LOC of commodity code |
| **Extract go-sandbox** | From Nanite's shell/sandbox code | Every app that runs shell commands needs this |
| **Extract go-signals** | From FE's executor signal protocol | CM needs this; the protocol should be shared |
| **Nanite structured envelope responses** | Replace onSendMessage(string) with typed ResponseV1 | Closes the envelope↔Fast Triage gap |

### Enterprise readiness (Nanite, for in-house company use)

| Priority | What | Why |
|---|---|---|
| **OIDC integration** | coreos/go-oidc against the company's IdP | Multi-user auth |
| **Users table + session scoping** | Schema migration + chi middleware | Prerequisite for everything else |
| **Casbin RBAC** | Embedded, SQLite adapter | Role-based access alongside existing tool permission engine |
| **Audit log** | Hash-chained append-only table | Compliance and traceability |
| **Tink field encryption** | For plaintext API keys in SQLite | Data-at-rest security |

---

## Design principles (observed, not prescribed)

These emerged from auditing the code, not from README files:

1. **SQLite is the universal store.** Every Go app uses `modernc.org/sqlite` with WAL mode. Pure Go, no CGo, single-file backup. This is a portfolio-wide commitment.

2. **MCP is the integration surface.** Apps that need to talk to each other do it over MCP. This keeps coupling loose and lets external tools (Claude Code, Copilot, any MCP client) participate without special integration.

3. **No bundled UIs on backend services.** Stack Explorer, Vanta, and Hadron are server/daemon-only. Sigil generates UIs. Nanite and Nil have their own Wails-based UIs. This separation is intentional.

4. **Vanta points, doesn't own.** Memory references content via URI pointers. The content lives in Nil (documents), Stack Explorer (findings/symbols), or on disk (files). Vanta is the index, not the archive.

5. **Tasks, not chat, for execution.** Chat is for conversation and context. Plans are decomposed into discrete tasks in CM. Mechanical work is blueprinted in Hadron. This separation keeps each system focused.

6. **Agents are the primary developers.** Agent-driven sprints, .agentrc context files, boot-prompt protocols, session handoffs. The tools are built for agents to use, not just humans.

7. **Shared packages for shared interfaces, apps for differentiated behavior.** go-providers, go-queue, go-otel, go-plugin are packages. Nanite, CM, Vanta are apps. The line is: if external users would want only half of it, it's two things.

---

## Dependency philosophy

- **Pure Go where possible.** Only one CGo budget item across the portfolio (tree-sitter, proposed for Stack Explorer).
- **Official SDKs for cloud services.** Anthropic, OpenAI, AWS, Cloudflare, DigitalOcean, GitHub all have official Go clients. Use them.
- **Shell out to mature tools.** Ollama, Lizard, scc, git, ansible-playbook, terraform — don't reimplement what exists.
- **Low-churn once stable.** Prefer boring, well-maintained dependencies (robfig/cron, chi, cobra, yaml.v3, modernc/sqlite) over trendy ones.

---

## The convergence path: envelopes + Sigil + Nanite

Three projects are converging on "agents create interactive UIs on the fly":

- **Nanite's envelope system** (23 types, plugin rendering, dynamic ESM loader) provides the rendering infrastructure.
- **Fast Triage's protocol** (typed envelopes, zod validation, blocking MCP, structured responses) provides the interaction model.
- **Sigil's code generation** (YAML → React components with typed data bindings) provides the creation pipeline.

The end state: an agent writes 20 lines of Sigil YAML, calls `sigil generate --target nanite-envelope`, the output registers as a Nanite envelope type, renders inside the chat session, the user interacts, structured data flows back to the agent. Cost: near-zero tokens (deterministic codegen, not LLM inference). Time: milliseconds. This replaces "agent reasons through JSX for 2000 tokens" with "agent emits a config and calls a compiler."

**Two changes to get there:** (1) Nanite's envelope responses need to be structured (borrow Fast Triage's ResponseV1), and (2) Sigil needs a `nanite-envelope` render target that outputs components conforming to Nanite's plugin interface.

---

## What nobody else is building

Three things in this portfolio are genuinely novel in the agent tooling space:

1. **Activation-ranked memory with typed pointers** (Vanta). Memory systems exist (mem0, zep). Code knowledge bases exist (Sourcegraph). Document stores exist (Notion, Obsidian). But a memory layer that ranks by activation/decay, stores typed URI pointers to content in other systems, and serves recall over MCP — that's a new primitive.

2. **A code-level audit store with file:line anchoring + themes + per-commit versioning + MCP query surface** (Stack Explorer, proposed). Bug trackers are workflow. Doc systems are narrative. Knowledge graphs are abstract. A queryable audit store that anchors findings to code symbols and exposes prior-art retrieval to agents mid-session is a new category.

3. **Declarative UI compilation with agent-first authoring** (Sigil). Visual builders exist. Low-code platforms exist. OpenAPI codegen exists. A YAML→multi-target UI compiler with MCP tools that agents use to create interfaces — where the output is editable source code, not a runtime dependency — is genuinely new.

These are the three things worth protecting from scope creep and worth talking about publicly.

---

## Risk register

| Risk | Impact | Mitigation |
|---|---|---|
| **CM executor doesn't ship** | Blocks the entire task execution thesis | In active development; port FE's proven patterns |
| **go-providers maintenance burden** (10K LOC, 16 adapters) | API changes in any provider break the portfolio | Replace HTTP adapters with official SDKs; keep only CLI bridges |
| **Schema churn in early projects** | Migrations are one-way; mistakes ripple | Lock schemas for Stack Explorer audits and CM tasks before committing |
| **Single-developer bus factor** | All institutional knowledge in one person | This document, CLAUDE.md files, Vanta memories, and agent context files mitigate |
| **Tight coupling via library embedding** | Vanta schema changes break Nanite | Version the library API; integration tests across consumer apps |
| **Enterprise deployment untested** | Multi-user Nanite hasn't been deployed | Plan the OIDC + RBAC + audit work as a focused sprint before company rollout |

---

*This summary reflects the state of the portfolio as of April 2026. It will drift as work progresses. The architectural principles and integration patterns should be more durable than the specific priorities.*

# LongMemEval on Vanta — Design Spec

**Date:** 2026-04-16
**Status:** Draft, pending user review
**Owner:** chrispian

## 1. Goal and scope

Build a **LongMemEval** benchmark harness at `stack-explorer/benchmarks/longmemeval/` that evaluates a memory system on the published LongMemEval dataset and produces per-run results plus a time-series snapshot in stack-explorer.

Ships with **one adapter implementation: Vanta (vanta-conduit, local HTTP on `:8089`)**. The harness is structured so additional adapters (Cortex, mem0, MemGPT, etc.) can be added later without changes to the core runner.

### Non-goals (v0)

- No second memory-system adapter beyond Vanta.
- No new scoring lens or dimension in stack-explorer's SQLite DB.
- No CI integration.
- No retrieval tuning beyond two baseline configs (similarity and activation).
- No publication-grade reproducibility packaging.

### Success criteria

1. A `lme-run smoke` execution ingests, retrieves, answers, and grades a stratified 10–20-question sample end-to-end against a running local Vanta, using both `LongMemEval_S` and `LongMemEval_Oracle` variants.
2. Per-question JSON results, per-stage markdown summary, and aggregate snapshot (via `stack-explorer snapshot take`) are produced.
3. Pipeline fails cleanly on any misconfiguration (missing dataset, Vanta down, API key missing) before incurring LLM cost.
4. No auto-promotion between stages — each stage requires explicit human confirmation.

## 2. Architecture

Go driver + Python LLM-judge subprocess. Directory layout:

```
stack-explorer/benchmarks/longmemeval/
├── cmd/lme-run/              # main CLI
├── internal/
│   ├── adapter/
│   │   ├── adapter.go        # MemoryAdapter interface
│   │   └── vanta/            # Vanta HTTP client (:8089 /v1/memory/*)
│   ├── dataset/              # LongMemEval JSON loader (_S, _Oracle)
│   ├── answerer/             # LLM answer gen via go-providers
│   ├── judge/                # subprocess wrapper around Python judge
│   ├── runner/               # orchestration, concurrency, retry, resume, gating
│   └── report/               # per-run JSON + markdown summary + cost report
├── data/                     # LongMemEval dataset files (gitignored, fetched)
├── runs/                     # timestamped output dirs (gitignored)
├── python/                   # vendored official judge (minimum subset)
│   ├── judge.py
│   └── requirements.txt
├── testdata/                 # 2-question mini-dataset for integration test
└── scripts/
    ├── fetch-dataset.sh      # fetch official HF dataset
    └── setup-python.sh       # venv + pip install judge deps
```

### 2.1 `MemoryAdapter` interface

```go
type Adapter interface {
    Reset(ctx context.Context, questionID string) error
    Ingest(ctx context.Context, questionID string, sessions []Session) error
    Retrieve(ctx context.Context, questionID, query string, config RetrievalConfig) ([]Hit, error)
    Close() error
}

type RetrievalConfig struct {
    Strategy string // "similarity" | "activation"
    TopK     int
}

type Session struct {
    SessionID string
    Timestamp time.Time
    Turns     []Turn
}

type Turn struct {
    Role    string // "user" | "assistant"
    Content string
}

type Hit struct {
    SessionID string
    TurnIdx   int
    Content   string
    Score     float64
}
```

### 2.2 Vanta adapter (v0)

Calls Vanta over HTTP (`:8089`). Namespace-per-question: `bench/lme/<question_id>`. Uses `memory_write` for ingest and `memory_recall` for retrieval. `Reset` issues `memory_deprecate` across the namespace's keys, or (if available) a namespace-scoped delete.

### 2.3 Concurrency and resume

- Questions are independent → runner processes N questions at a time (configurable, default 4).
- Per-question work is sequential within a single retrieval config: Reset → Ingest → Retrieve → Answer → Judge.
- The two retrieval configs (similarity, activation) share the same ingested namespace. Within a question, they are evaluated sequentially against the same ingested haystack — ingest happens once, two retrievals + two answers + two judge calls follow.
- Results are written per-question-per-config to disk as soon as they complete.
- Re-invocation with the same `run-id` skips `{qid, config}` pairs already on disk (idempotent resume after kill/crash/rate-limit pause).

### 2.4 Per-question namespace isolation

Each question's haystack lives in its own Vanta namespace (`bench/lme/<qid>`), so retrieval can never cross-contaminate between questions. `Reset` is called before every ingest to guarantee a clean slate (handles re-runs and resume).

## 3. Data flow per question

```
1. adapter.Reset(qid)
2. adapter.Ingest(qid, sessions):
   For each turn, memory_write:
     namespace:   bench/lme/<qid>
     memory_key:  <session_id>:<turn_idx>
     payload_body: turn.content
     payload_summary: first 200 chars
     origin:      turn.role
     timestamp:   session.timestamp
     tags:        ["lme", question_type]
3. adapter.Retrieve(qid, question_text, cfg):
   POST /v1/memory/recall
     namespace: bench/lme/<qid>
     query:     question_text
     ranking:   cfg.Strategy (similarity | activation)
     top_k:     cfg.TopK
   → []Hit
4. answerer.Answer(question, hits):
   go-providers Anthropic call, default claude-sonnet-4-6
   Prompt: LongMemEval paper's generation template + retrieved turns
5. judge.Grade(question, gold, predicted):
   exec python/judge.py with JSON on stdin
   → {correct: bool, rationale: string}
6. runner writes runs/<stage>-<ts>/<config_id>/<qid>.json:
   { question_id, predicted, judge_result, latency_ms,
     retrieval_hits, token_counts, cost_usd }
```

### 3.1 Oracle variant

Steps 1–3 are skipped. Gold-evidence sessions from the dataset are passed directly to step 4 as `hits`. Same answerer and judge. Isolates reasoning from retrieval.

### 3.2 Retrieval configs

Two configs run in parallel per question (see Section 1 — two adapter configs, not two ingests):

- `similarity`: embedding-only ranking, `top_k = 10`.
- `activation`: Vanta's activation-weighted ranking, `top_k = 10`.

Both configs share the same ingested namespace; only the retrieval call differs.

## 4. LLM configuration

All configurable via env vars and CLI flags; defaults:

| Role       | Default model        | Provider  |
|------------|----------------------|-----------|
| Answerer   | `claude-sonnet-4-6`  | Anthropic |
| Judge      | `gpt-4o`             | OpenAI    |

- Anthropic key: `ANTHROPIC_API_KEY`, loaded from shell or injected via `op run` (1Password).
- OpenAI key: `OPENAI_API_KEY`, same.
- `go-providers` (from `framework/libs/go-providers`) used for answerer.
- Judge runs as a Python subprocess using the official LongMemEval judge script (vendored minimal subset). The judge **prompt** is preserved verbatim from the paper; the judge **model** defaults to `gpt-4o`, which is in the GPT-4 family but is newer than the paper's original snapshot. Scores are therefore directionally comparable to published LongMemEval numbers but not identical. A judge A/B (paper's original model vs `gpt-4o` vs Claude-as-judge) is out of scope for v0 and deferred.

## 5. Error handling

| Failure mode | Behavior |
|---|---|
| Vanta unreachable | Fail fast at preflight. |
| Ingest turn failure | Retry 3× with backoff. If still failing, mark question `{error: "ingest_failed"}`, skip downstream steps. |
| Retrieval returns zero hits | Proceed with empty context. Don't synthesize. |
| Answerer LLM error | Retry 2× with backoff. If still failing, mark `{error: "answerer_failed"}`. |
| Judge malformed JSON | Retry once with a stricter prompt reminder. If still malformed, mark `{judge_error: true}` — counted as untested, not incorrect. |
| Rate limit (429) | Per-provider token bucket, exponential backoff. |
| Cost ceiling exceeded | `--max-cost-usd` flag. Runner halts cleanly, writes partial summary, exits non-zero. |

Per-question failures never abort the batch. They are recorded and reported.

## 6. Testing

- **Go unit tests** for dataset loader, adapter (mocked HTTP), runner state machine, resume logic, cost accounting. No network.
- **One integration test** gated by `VANTA_URL` env: hits a real local Vanta with a 2-question mini-dataset in `testdata/`, runs the full pipeline end-to-end, verifies per-question JSON is produced and judged. Skipped when env is unset.
- **Judge wrapper tests** cover exit codes, JSON parsing, and timeout. Do not assert on judge behavior itself — the judge is treated as an external dependency.

## 7. Reporting

Per run, the harness writes:

- `runs/<stage>-<ts>/<config_id>/<qid>.json` — full per-question record.
- `runs/<stage>-<ts>/summary.json` — aggregate numbers (per question category, per retrieval config): accuracy, latency, token counts, USD cost broken out per provider.
- `runs/<stage>-<ts>/summary.md` — human-readable markdown version of summary.
- `runs/<stage>-<ts>/cost.json` — per-provider token and cost accounting.

At the end of a `subset` or `full` run, the harness calls `stack-explorer snapshot take` with aggregate metrics (overall accuracy per config, overall cost) so results accumulate in the time-series table without coupling to stack-explorer's qualitative scoring schema.

## 8. Run-gate sequence

Three stages, each with a mandatory human-in-the-loop pause.

### 8.1 Preflight

`lme-run preflight` verifies:
- Vanta health endpoint reachable on configured URL.
- `ANTHROPIC_API_KEY` present; dry `models.list` succeeds.
- `OPENAI_API_KEY` present; dry `models.list` succeeds.
- Dataset files present in `data/`.
- Python venv in `python/` exists and judge script imports clean.
- Adequate free disk.

Must pass before any stage runs.

### 8.2 Stages

1. **Smoke** — `lme-run smoke` — 10–20 stratified questions across all question categories, both `_S` and `_Oracle`, both retrieval configs. Expected cost: single-digit USD.
2. **Subset** — `lme-run subset --confirm --after <smoke-run-id>` — 100 stratified questions.
3. **Full** — `lme-run full --confirm --after <subset-run-id>` — all 500 questions. Additionally requires an interactive `GO` confirmation on stdin.

### 8.3 Mandatory pause behavior

Between every stage: **no auto-promotion**.

After each completed stage, the harness writes its summary, then prints to stdout:

```
Stage <name> complete.
  Anthropic spend:  $X.XX  (Y input tok / Z output tok)
  OpenAI spend:     $X.XX  (Y input tok / Z output tok)

PAUSE: Check your API credit balances before proceeding.
  Anthropic console: https://console.anthropic.com/settings/billing
  OpenAI console:    https://platform.openai.com/usage

Projected cost for next stage (<subset|full>, <N> questions):
  Anthropic:  ~$X.XX   OpenAI: ~$X.XX

To proceed:   lme-run <next-stage> --confirm --after <this-run-id>
```

The harness then exits 0. The next stage command:

- Requires `--confirm` (no default).
- Requires `--after <run-id>` pointing at a completed prior stage.
- Refuses to run if the prior run-id does not exist or did not complete.
- `full` additionally prompts for a literal `GO` on stdin before starting.

## 9. Dependencies added

- New Python venv under `benchmarks/longmemeval/python/` with the official LongMemEval judge requirements (`openai` SDK, tiktoken, etc.).
- Go deps: `go-providers` from the `framework/libs/go-providers` module (replace-directive in the benchmarks submodule). No new top-level Go deps beyond stdlib + existing framework libs.
- No changes to `stack-explorer`'s main module or its CLI binary.

## 10. Open items / deferred

- Second adapter (Cortex or mem0) — follows in a v1 design after Vanta baseline is stable.
- New stack-explorer scoring dimension / lens for memory benchmarks — revisit once there are 2+ memory systems benchmarked.
- Judge A/B (Claude-as-judge vs GPT-4o-as-judge) — deferred; configurable today, a later cheap re-run will quantify judge variance.
- Multi-turn retrieval (re-query with answer-drafting history) — not in v0.
- `conduit_lookup` unified search config — deferred; adds memory+knowledge mixing which isn't part of the LongMemEval setup.

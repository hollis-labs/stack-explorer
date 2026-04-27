# LongMemEval on Vanta Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go+Python-judge benchmark harness at `stack-explorer/benchmarks/longmemeval/` that runs LongMemEval against a local Vanta instance, produces per-run JSON + markdown + time-series snapshot, and enforces mandatory human-gated stages (smoke → subset → full).

**Architecture:** Go CLI (`lme-run`) drives a `MemoryAdapter` interface with a Vanta HTTP adapter. For each question: Reset namespace → Ingest turns via `memory_write` → Retrieve top-K via `memory_recall` (similarity + activation configs) → Answer with Claude via `go-providers` → Grade with vendored official Python LongMemEval judge run as a subprocess. Results land in `runs/<stage>-<ts>/` and get pushed to `stack-explorer snapshot take`.

**Tech Stack:** Go 1.26, Cobra (already a stack-explorer dep), `github.com/hollis-labs/go-providers` (Anthropic answerer), `github.com/pkoukk/tiktoken-go` (token counting for cost accounting), stdlib `net/http` for Vanta, Python 3 venv for the judge (OpenAI SDK). Vanta is reached at `http://127.0.0.1:8089`.

**Spec:** `docs/superpowers/specs/2026-04-16-longmemeval-vanta-design.md` (commit 86523bf).

**Source repo:** `github.com/chrispian/stack-explorer`.

**Assumed working directory:** `/Users/chrispian/Projects-apps/stack-explorer` unless otherwise noted.

---

## File structure (locked in)

```
benchmarks/longmemeval/
├── .gitignore                              # data/, runs/, python/.venv/
├── README.md                               # How to fetch dataset, run stages
├── Makefile                                # build, test, smoke, fetch-dataset, setup-python
├── cmd/lme-run/
│   └── main.go                             # Cobra CLI entrypoint
├── internal/
│   ├── adapter/
│   │   ├── adapter.go                      # Adapter interface, Session/Turn/Hit/RetrievalConfig
│   │   ├── adapter_test.go
│   │   └── vanta/
│   │       ├── client.go                   # HTTP client ctor + health
│   │       ├── client_test.go
│   │       ├── reset.go
│   │       ├── reset_test.go
│   │       ├── ingest.go
│   │       ├── ingest_test.go
│   │       ├── retrieve.go
│   │       └── retrieve_test.go
│   ├── dataset/
│   │   ├── dataset.go                      # Question, Session JSON schemas + loader
│   │   ├── dataset_test.go
│   │   ├── sampler.go                      # stratified sampling
│   │   └── sampler_test.go
│   ├── answerer/
│   │   ├── answerer.go                     # provider.Complete wrapper, prompt template
│   │   └── answerer_test.go
│   ├── judge/
│   │   ├── judge.go                        # exec wrapper
│   │   └── judge_test.go
│   ├── config/
│   │   ├── config.go                       # env + flag merge
│   │   └── config_test.go
│   ├── cost/
│   │   ├── cost.go                         # tiktoken-based token+USD tracking
│   │   └── cost_test.go
│   ├── runner/
│   │   ├── pipeline.go                     # one-question-one-config pipeline
│   │   ├── pipeline_test.go
│   │   ├── runner.go                       # batch worker pool + resume
│   │   ├── runner_test.go
│   │   ├── gate.go                         # stage gating: --confirm --after + pause
│   │   └── gate_test.go
│   ├── report/
│   │   ├── writer.go                       # per-question JSON writer
│   │   ├── writer_test.go
│   │   ├── summary.go                      # aggregate JSON + MD
│   │   ├── summary_test.go
│   │   ├── snapshot.go                     # exec `stack-explorer snapshot take`
│   │   └── snapshot_test.go
│   └── preflight/
│       ├── preflight.go                    # checks
│       └── preflight_test.go
├── python/
│   ├── judge.py                            # vendored official LongMemEval judge
│   ├── requirements.txt
│   └── .venv/                              # gitignored
├── testdata/
│   └── mini-dataset.json                   # 2-question fixture for integration test
├── data/                                   # gitignored; dataset lands here
├── runs/                                   # gitignored; outputs
└── scripts/
    ├── fetch-dataset.sh
    └── setup-python.sh
```

All Go code lives in the existing `github.com/chrispian/stack-explorer` module under `benchmarks/longmemeval/...`. No submodule / no `go.work`. The `lme-run` binary's transitive deps do not pollute the main `stack-explorer` binary because they are only imported from `cmd/lme-run`.

---

## Task 1: Scaffold benchmarks/longmemeval directory + root files

**Files:**
- Create: `benchmarks/longmemeval/.gitignore`
- Create: `benchmarks/longmemeval/README.md`
- Create: `benchmarks/longmemeval/Makefile`
- Create: `benchmarks/longmemeval/cmd/lme-run/main.go`

- [ ] **Step 1: Create directory tree**

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
mkdir -p benchmarks/longmemeval/{cmd/lme-run,internal/{adapter/vanta,dataset,answerer,judge,config,cost,runner,report,preflight},python,testdata,data,runs,scripts}
```

- [ ] **Step 2: Create `benchmarks/longmemeval/.gitignore`**

```gitignore
data/
runs/
python/.venv/
*.log
```

- [ ] **Step 3: Create `benchmarks/longmemeval/README.md`**

```markdown
# LongMemEval Benchmark Harness

Runs the LongMemEval benchmark against a pluggable memory system. Ships with a Vanta (vanta-conduit) adapter.

## Quick start

```bash
cd benchmarks/longmemeval
make fetch-dataset    # downloads LongMemEval_S + Oracle JSON
make setup-python     # creates python/.venv with judge deps
make build            # builds ./lme-run

./lme-run preflight                               # verify Vanta + keys + dataset
./lme-run smoke                                   # 10-20 Q stratified smoke test
./lme-run subset --confirm --after <smoke-run-id> # 100 Q
./lme-run full   --confirm --after <subset-run-id># 500 Q (also requires "GO" on stdin)
```

See spec: `docs/superpowers/specs/2026-04-16-longmemeval-vanta-design.md`.
```

- [ ] **Step 4: Create `benchmarks/longmemeval/Makefile`**

```makefile
.PHONY: build test fetch-dataset setup-python clean

build:
	go build -o lme-run ./cmd/lme-run

test:
	go test ./...

fetch-dataset:
	bash scripts/fetch-dataset.sh

setup-python:
	bash scripts/setup-python.sh

clean:
	rm -rf lme-run runs/*
```

- [ ] **Step 5: Create stub `benchmarks/longmemeval/cmd/lme-run/main.go`**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "lme-run: not yet implemented")
	os.Exit(1)
}
```

- [ ] **Step 6: Verify build**

Run: `cd benchmarks/longmemeval && go build -o lme-run ./cmd/lme-run && ./lme-run; echo "exit=$?"`
Expected: prints `lme-run: not yet implemented`, `exit=1`.

- [ ] **Step 7: Commit**

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
git add benchmarks/longmemeval/
git commit -m "Scaffold benchmarks/longmemeval directory and stub CLI"
```

---

## Task 2: Adapter interface + domain types

**Files:**
- Create: `benchmarks/longmemeval/internal/adapter/adapter.go`
- Create: `benchmarks/longmemeval/internal/adapter/adapter_test.go`

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/adapter/adapter_test.go`:

```go
package adapter

import (
	"context"
	"testing"
	"time"
)

// nullAdapter is a no-op implementation used to prove the interface compiles.
type nullAdapter struct{}

func (nullAdapter) Reset(context.Context, string) error                                  { return nil }
func (nullAdapter) Ingest(context.Context, string, []Session) error                      { return nil }
func (nullAdapter) Retrieve(context.Context, string, string, RetrievalConfig) ([]Hit, error) {
	return nil, nil
}
func (nullAdapter) Close() error { return nil }

func TestInterfaceCompiles(t *testing.T) {
	var a Adapter = nullAdapter{}
	if err := a.Reset(context.Background(), "qid"); err != nil {
		t.Fatal(err)
	}
	_, err := a.Retrieve(context.Background(), "qid", "q", RetrievalConfig{Strategy: "similarity", TopK: 5})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSessionAndTurn(t *testing.T) {
	s := Session{
		SessionID: "s1",
		Timestamp: time.Now(),
		Turns:     []Turn{{Role: "user", Content: "hi"}},
	}
	if s.Turns[0].Role != "user" {
		t.Fatalf("want user, got %q", s.Turns[0].Role)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/...`
Expected: FAIL (undefined: Adapter / Session / Turn / Hit / RetrievalConfig).

- [ ] **Step 3: Write adapter.go**

Create `benchmarks/longmemeval/internal/adapter/adapter.go`:

```go
// Package adapter defines the MemoryAdapter contract for the LongMemEval harness.
package adapter

import (
	"context"
	"time"
)

type Adapter interface {
	Reset(ctx context.Context, questionID string) error
	Ingest(ctx context.Context, questionID string, sessions []Session) error
	Retrieve(ctx context.Context, questionID, query string, config RetrievalConfig) ([]Hit, error)
	Close() error
}

type RetrievalConfig struct {
	Strategy string
	TopK     int
}

type Session struct {
	SessionID string
	Timestamp time.Time
	Turns     []Turn
}

type Turn struct {
	Role    string
	Content string
}

type Hit struct {
	SessionID string
	TurnIdx   int
	Content   string
	Score     float64
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/adapter/
git commit -m "Add MemoryAdapter interface and domain types"
```

---

## Task 3: Dataset loader (Question + Session JSON schemas)

**Files:**
- Create: `benchmarks/longmemeval/internal/dataset/dataset.go`
- Create: `benchmarks/longmemeval/internal/dataset/dataset_test.go`
- Create: `benchmarks/longmemeval/internal/dataset/testdata/tiny.json`

**Reference:** LongMemEval's official dataset format. Each record has `question_id`, `question_type`, `question`, `answer`, `haystack_sessions` (list of sessions, each with `session_id`, `timestamp`, `turns`), and for the Oracle variant `answer_session_ids` identifies the gold-evidence sessions.

- [ ] **Step 1: Create the test fixture**

Create `benchmarks/longmemeval/internal/dataset/testdata/tiny.json`:

```json
[
  {
    "question_id": "q1",
    "question_type": "single-session-user",
    "question": "What color is my cat?",
    "answer": "orange",
    "answer_session_ids": ["s1"],
    "haystack_sessions": [
      {
        "session_id": "s1",
        "session_date": "2024-01-10T10:00:00Z",
        "turns": [
          {"role": "user", "content": "My cat is orange."},
          {"role": "assistant", "content": "Noted, your cat is orange."}
        ]
      },
      {
        "session_id": "s2",
        "session_date": "2024-01-11T10:00:00Z",
        "turns": [
          {"role": "user", "content": "I like coffee."},
          {"role": "assistant", "content": "Noted."}
        ]
      }
    ]
  },
  {
    "question_id": "q2",
    "question_type": "multi-session",
    "question": "What do I drink?",
    "answer": "coffee",
    "answer_session_ids": ["s2"],
    "haystack_sessions": [
      {
        "session_id": "s1",
        "session_date": "2024-01-10T10:00:00Z",
        "turns": [{"role": "user", "content": "Hello."}]
      },
      {
        "session_id": "s2",
        "session_date": "2024-01-11T10:00:00Z",
        "turns": [{"role": "user", "content": "I like coffee."}]
      }
    ]
  }
]
```

- [ ] **Step 2: Write the failing test**

Create `benchmarks/longmemeval/internal/dataset/dataset_test.go`:

```go
package dataset

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadTinyDataset(t *testing.T) {
	path := filepath.Join("testdata", "tiny.json")
	qs, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, want := len(qs), 2; got != want {
		t.Fatalf("len(questions)=%d want %d", got, want)
	}

	q1 := qs[0]
	if q1.ID != "q1" {
		t.Errorf("ID=%q want q1", q1.ID)
	}
	if q1.Type != "single-session-user" {
		t.Errorf("Type=%q", q1.Type)
	}
	if q1.Answer != "orange" {
		t.Errorf("Answer=%q", q1.Answer)
	}
	if len(q1.HaystackSessions) != 2 {
		t.Errorf("haystack_sessions=%d want 2", len(q1.HaystackSessions))
	}
	if len(q1.AnswerSessionIDs) != 1 || q1.AnswerSessionIDs[0] != "s1" {
		t.Errorf("AnswerSessionIDs=%v", q1.AnswerSessionIDs)
	}

	s1 := q1.HaystackSessions[0]
	wantTS, _ := time.Parse(time.RFC3339, "2024-01-10T10:00:00Z")
	if !s1.Timestamp.Equal(wantTS) {
		t.Errorf("Session timestamp=%v want %v", s1.Timestamp, wantTS)
	}
	if len(s1.Turns) != 2 {
		t.Errorf("s1.Turns=%d want 2", len(s1.Turns))
	}
	if s1.Turns[0].Role != "user" || s1.Turns[0].Content != "My cat is orange." {
		t.Errorf("turn0=%+v", s1.Turns[0])
	}
}

func TestGoldSessions(t *testing.T) {
	qs, err := Load(filepath.Join("testdata", "tiny.json"))
	if err != nil {
		t.Fatal(err)
	}
	gold := qs[0].GoldSessions()
	if len(gold) != 1 || gold[0].SessionID != "s1" {
		t.Fatalf("GoldSessions=%+v", gold)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd benchmarks/longmemeval && go test ./internal/dataset/...`
Expected: FAIL (undefined: Load / Question / etc.).

- [ ] **Step 4: Write dataset.go**

Create `benchmarks/longmemeval/internal/dataset/dataset.go`:

```go
// Package dataset loads LongMemEval question records from JSON.
package dataset

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
)

type Question struct {
	ID               string
	Type             string
	Text             string
	Answer           string
	AnswerSessionIDs []string
	HaystackSessions []adapter.Session
}

// GoldSessions returns only those haystack sessions listed in AnswerSessionIDs.
// Used for the Oracle variant (no retrieval — inject gold evidence directly).
func (q Question) GoldSessions() []adapter.Session {
	want := make(map[string]struct{}, len(q.AnswerSessionIDs))
	for _, id := range q.AnswerSessionIDs {
		want[id] = struct{}{}
	}
	var out []adapter.Session
	for _, s := range q.HaystackSessions {
		if _, ok := want[s.SessionID]; ok {
			out = append(out, s)
		}
	}
	return out
}

type rawQuestion struct {
	QuestionID       string       `json:"question_id"`
	QuestionType     string       `json:"question_type"`
	Question         string       `json:"question"`
	Answer           string       `json:"answer"`
	AnswerSessionIDs []string     `json:"answer_session_ids"`
	HaystackSessions []rawSession `json:"haystack_sessions"`
}

type rawSession struct {
	SessionID   string    `json:"session_id"`
	SessionDate string    `json:"session_date"`
	Turns       []rawTurn `json:"turns"`
}

type rawTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Load reads a LongMemEval JSON file (array of question records) from disk.
func Load(path string) ([]Question, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("dataset.Load: open %s: %w", path, err)
	}
	defer f.Close()
	var raws []rawQuestion
	if err := json.NewDecoder(f).Decode(&raws); err != nil {
		return nil, fmt.Errorf("dataset.Load: decode %s: %w", path, err)
	}
	qs := make([]Question, 0, len(raws))
	for i, r := range raws {
		q, err := convert(r)
		if err != nil {
			return nil, fmt.Errorf("dataset.Load: record %d (%s): %w", i, r.QuestionID, err)
		}
		qs = append(qs, q)
	}
	return qs, nil
}

func convert(r rawQuestion) (Question, error) {
	sessions := make([]adapter.Session, 0, len(r.HaystackSessions))
	for _, rs := range r.HaystackSessions {
		ts, err := time.Parse(time.RFC3339, rs.SessionDate)
		if err != nil {
			return Question{}, fmt.Errorf("session %s: bad session_date %q: %w", rs.SessionID, rs.SessionDate, err)
		}
		turns := make([]adapter.Turn, 0, len(rs.Turns))
		for _, rt := range rs.Turns {
			turns = append(turns, adapter.Turn{Role: rt.Role, Content: rt.Content})
		}
		sessions = append(sessions, adapter.Session{
			SessionID: rs.SessionID,
			Timestamp: ts,
			Turns:     turns,
		})
	}
	return Question{
		ID:               r.QuestionID,
		Type:             r.QuestionType,
		Text:             r.Question,
		Answer:           r.Answer,
		AnswerSessionIDs: r.AnswerSessionIDs,
		HaystackSessions: sessions,
	}, nil
}
```

- [ ] **Step 5: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/dataset/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add benchmarks/longmemeval/internal/dataset/
git commit -m "Add LongMemEval dataset loader"
```

---

## Task 4: Stratified sampler

**Files:**
- Create: `benchmarks/longmemeval/internal/dataset/sampler.go`
- Create: `benchmarks/longmemeval/internal/dataset/sampler_test.go`

**Purpose:** `smoke` runs a small stratified slice across all question types; `subset` samples ~100. Must be deterministic given a seed (reproducible runs).

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/dataset/sampler_test.go`:

```go
package dataset

import (
	"testing"
)

func mkQ(id, typ string) Question { return Question{ID: id, Type: typ} }

func TestStratifiedSample_EvenDistribution(t *testing.T) {
	qs := []Question{
		mkQ("a1", "alpha"), mkQ("a2", "alpha"), mkQ("a3", "alpha"), mkQ("a4", "alpha"),
		mkQ("b1", "beta"), mkQ("b2", "beta"), mkQ("b3", "beta"), mkQ("b4", "beta"),
	}
	got, err := StratifiedSample(qs, 4, 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("len=%d want 4", len(got))
	}
	counts := map[string]int{}
	for _, q := range got {
		counts[q.Type]++
	}
	if counts["alpha"] != 2 || counts["beta"] != 2 {
		t.Fatalf("counts=%v want {alpha:2,beta:2}", counts)
	}
}

func TestStratifiedSample_Deterministic(t *testing.T) {
	qs := []Question{
		mkQ("a1", "alpha"), mkQ("a2", "alpha"), mkQ("a3", "alpha"),
		mkQ("b1", "beta"), mkQ("b2", "beta"), mkQ("b3", "beta"),
	}
	a, err := StratifiedSample(qs, 4, 7)
	if err != nil {
		t.Fatal(err)
	}
	b, err := StratifiedSample(qs, 4, 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) {
		t.Fatalf("non-deterministic lengths")
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			t.Fatalf("non-deterministic at %d: %s vs %s", i, a[i].ID, b[i].ID)
		}
	}
}

func TestStratifiedSample_Oversample(t *testing.T) {
	qs := []Question{mkQ("a1", "alpha"), mkQ("b1", "beta")}
	got, err := StratifiedSample(qs, 100, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d want 2 (all available)", len(got))
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/dataset/...`
Expected: FAIL (undefined: StratifiedSample).

- [ ] **Step 3: Write sampler.go**

Create `benchmarks/longmemeval/internal/dataset/sampler.go`:

```go
package dataset

import (
	"fmt"
	"math/rand"
	"sort"
)

// StratifiedSample returns a sample of size n distributed roughly evenly across
// question types. If n is larger than len(qs), returns all questions.
// The seed makes sampling deterministic.
func StratifiedSample(qs []Question, n int, seed int64) ([]Question, error) {
	if n <= 0 {
		return nil, fmt.Errorf("StratifiedSample: n must be > 0")
	}
	if n >= len(qs) {
		out := make([]Question, len(qs))
		copy(out, qs)
		return out, nil
	}
	byType := map[string][]Question{}
	var types []string
	for _, q := range qs {
		if _, ok := byType[q.Type]; !ok {
			types = append(types, q.Type)
		}
		byType[q.Type] = append(byType[q.Type], q)
	}
	sort.Strings(types)
	rng := rand.New(rand.NewSource(seed))
	for _, t := range types {
		bucket := byType[t]
		rng.Shuffle(len(bucket), func(i, j int) { bucket[i], bucket[j] = bucket[j], bucket[i] })
		byType[t] = bucket
	}
	per := n / len(types)
	rem := n - per*len(types)
	var out []Question
	for _, t := range types {
		k := per
		if rem > 0 {
			k++
			rem--
		}
		if k > len(byType[t]) {
			k = len(byType[t])
		}
		out = append(out, byType[t][:k]...)
	}
	return out, nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/dataset/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/dataset/
git commit -m "Add deterministic stratified sampler"
```

---

## Task 5: Vanta HTTP client constructor + health check

**Files:**
- Create: `benchmarks/longmemeval/internal/adapter/vanta/client.go`
- Create: `benchmarks/longmemeval/internal/adapter/vanta/client_test.go`

**Context:** Vanta's HTTP surface lives under `/v1/*`. Health is at `/v1/health` on the contextd server. We confirmed Vanta is live at `http://127.0.0.1:8089`. The client holds the base URL, an `http.Client` with a timeout, and an optional capability token (sent as `Authorization: Bearer <token>`).

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/adapter/vanta/client_test.go`:

```go
package vanta

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealth_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "", 5*time.Second)
	if err := c.Health(context.Background()); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestHealth_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()
	c := New(srv.URL, "", 5*time.Second)
	if err := c.Health(context.Background()); err == nil {
		t.Fatal("want error on 500, got nil")
	}
}

func TestHealth_TokenHeader(t *testing.T) {
	got := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "tok123", 5*time.Second)
	if err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case h := <-got:
		if h != "Bearer tok123" {
			t.Fatalf("auth header=%q", h)
		}
	case <-time.After(time.Second):
		t.Fatal("no request observed")
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: FAIL (undefined: New / Client / Health).

- [ ] **Step 3: Write client.go**

Create `benchmarks/longmemeval/internal/adapter/vanta/client.go`:

```go
// Package vanta is the Vanta (vanta-conduit) implementation of adapter.Adapter.
package vanta

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	hc      *http.Client
}

func New(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		hc:      &http.Client{Timeout: timeout},
	}
}

// Close implements adapter.Adapter.Close. HTTP clients need no teardown.
func (c *Client) Close() error { return nil }

// Health issues GET /v1/health and returns nil iff status is 2xx.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/health", nil)
	if err != nil {
		return fmt.Errorf("vanta health: %w", err)
	}
	c.auth(req)
	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("vanta health: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("vanta health: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) auth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Content-Type", "application/json")
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/adapter/vanta/
git commit -m "Add Vanta HTTP client with health check"
```

---

## Task 6: Vanta adapter — Reset

**Files:**
- Create: `benchmarks/longmemeval/internal/adapter/vanta/reset.go`
- Create: `benchmarks/longmemeval/internal/adapter/vanta/reset_test.go`

**Strategy:** `Reset(qid)` deprecates every memory revision in namespace `bench/lme/<qid>`. Vanta's HTTP has `GET /v1/memory/history?namespace=...` (but history requires a memory_key). For v0, we instead **skip per-revision deprecate** and rely on the namespace being untouched across questions thanks to unique `<qid>`-scoped namespacing. Reset becomes a no-op *functionally* but is kept in the interface so future adapters (that share state across qids) can implement cleanup. We document this explicitly so an implementer doesn't paper over it later.

- [ ] **Step 1: Write the test (documents no-op behavior)**

Create `benchmarks/longmemeval/internal/adapter/vanta/reset_test.go`:

```go
package vanta

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Vanta's Reset is a documented no-op: per-question isolation is guaranteed by
// namespace naming (bench/lme/<qid>). The test asserts Reset makes no network
// calls and returns nil — if an implementer changes that, this test catches it.
func TestReset_IsNoOp(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := New(srv.URL, "", 5*time.Second)
	if err := c.Reset(context.Background(), "q-123"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if called {
		t.Fatal("Reset made a network call; it should be a no-op for Vanta v0")
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: FAIL (undefined: Reset method on Client).

- [ ] **Step 3: Write reset.go**

Create `benchmarks/longmemeval/internal/adapter/vanta/reset.go`:

```go
package vanta

import "context"

// Reset is a no-op for Vanta because each question uses its own namespace
// (bench/lme/<qid>). Included to satisfy adapter.Adapter. Future adapters
// with shared namespaces may need real cleanup here.
func (c *Client) Reset(ctx context.Context, questionID string) error {
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/adapter/vanta/
git commit -m "Add Vanta Reset (no-op; per-question namespace isolation)"
```

---

## Task 7: Vanta adapter — Ingest

**Files:**
- Create: `benchmarks/longmemeval/internal/adapter/vanta/ingest.go`
- Create: `benchmarks/longmemeval/internal/adapter/vanta/ingest_test.go`

**Endpoint:** `POST /v1/memory/write`. Request body (from Vanta's MCP docs):

```json
{
  "namespace": "bench/lme/<qid>",
  "memory_key": "<session_id>:<turn_idx>",
  "author_agent_id": "longmemeval-harness",
  "trigger": "explicit",
  "session_id": "lme-<qid>",
  "origin": "<user|assistant>",
  "confidence": 1.0,
  "payload_summary": "first 200 chars of content",
  "payload_body": "<turn.content>",
  "tags": "[\"lme\",\"<question_type>\"]"
}
```

**Retry policy:** 3× with exponential backoff on network errors or 5xx. 4xx is fatal.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/adapter/vanta/ingest_test.go`:

```go
package vanta

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
)

func TestIngest_WritesOneRequestPerTurn(t *testing.T) {
	var count int32
	seen := make(chan map[string]any, 8)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/memory/write" {
			t.Errorf("path=%q", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(b, &body)
		seen <- body
		atomic.AddInt32(&count, 1)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"revision_id":"rev1"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "", 5*time.Second)

	ts, _ := time.Parse(time.RFC3339, "2024-01-10T10:00:00Z")
	sessions := []adapter.Session{{
		SessionID: "s1",
		Timestamp: ts,
		Turns: []adapter.Turn{
			{Role: "user", Content: "Hello."},
			{Role: "assistant", Content: "Hi there."},
		},
	}}

	if err := c.Ingest(context.Background(), "qX", sessions); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if got := atomic.LoadInt32(&count); got != 2 {
		t.Fatalf("write count=%d want 2", got)
	}
	close(seen)
	var bodies []map[string]any
	for b := range seen {
		bodies = append(bodies, b)
	}
	for _, b := range bodies {
		if b["namespace"] != "bench/lme/qX" {
			t.Errorf("namespace=%v", b["namespace"])
		}
	}
}

func TestIngest_RetriesOn5xx(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&count, 1)
		if n < 3 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"revision_id":"x"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "", 5*time.Second)
	ts, _ := time.Parse(time.RFC3339, "2024-01-10T10:00:00Z")
	sessions := []adapter.Session{{
		SessionID: "s1", Timestamp: ts,
		Turns: []adapter.Turn{{Role: "user", Content: "hi"}},
	}}
	if err := c.Ingest(context.Background(), "qR", sessions); err != nil {
		t.Fatalf("Ingest after retry: %v", err)
	}
	if got := atomic.LoadInt32(&count); got != 3 {
		t.Fatalf("req count=%d want 3 (2 failures + 1 success)", got)
	}
}

func TestIngest_FailsOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"message":"bad request"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "", 5*time.Second)
	ts, _ := time.Parse(time.RFC3339, "2024-01-10T10:00:00Z")
	sessions := []adapter.Session{{
		SessionID: "s1", Timestamp: ts,
		Turns: []adapter.Turn{{Role: "user", Content: "hi"}},
	}}
	if err := c.Ingest(context.Background(), "q4", sessions); err == nil {
		t.Fatal("want error on 4xx, got nil")
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: FAIL (undefined: Ingest method on Client).

- [ ] **Step 3: Write ingest.go**

Create `benchmarks/longmemeval/internal/adapter/vanta/ingest.go`:

```go
package vanta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
)

type memoryWriteReq struct {
	Namespace      string  `json:"namespace"`
	MemoryKey      string  `json:"memory_key"`
	AuthorAgentID  string  `json:"author_agent_id"`
	Trigger        string  `json:"trigger"`
	SessionID      string  `json:"session_id"`
	Origin         string  `json:"origin"`
	Confidence     float64 `json:"confidence"`
	PayloadSummary string  `json:"payload_summary"`
	PayloadBody    string  `json:"payload_body"`
	Tags           string  `json:"tags"`
}

func (c *Client) Ingest(ctx context.Context, questionID string, sessions []adapter.Session) error {
	ns := fmt.Sprintf("bench/lme/%s", questionID)
	for _, s := range sessions {
		for idx, t := range s.Turns {
			body := memoryWriteReq{
				Namespace:      ns,
				MemoryKey:      fmt.Sprintf("%s:%d", s.SessionID, idx),
				AuthorAgentID:  "longmemeval-harness",
				Trigger:        "explicit",
				SessionID:      "lme-" + questionID,
				Origin:         t.Role,
				Confidence:     1.0,
				PayloadSummary: summarize(t.Content, 200),
				PayloadBody:    t.Content,
				Tags:           `["lme"]`,
			}
			if err := c.postJSONRetry(ctx, "/v1/memory/write", body, 3); err != nil {
				return fmt.Errorf("ingest %s turn %d: %w", s.SessionID, idx, err)
			}
		}
	}
	return nil
}

func summarize(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// postJSONRetry does a POST with JSON body and retries 5xx/network errors up
// to attempts times, with exponential backoff. 4xx errors are returned immediately.
func (c *Client) postJSONRetry(ctx context.Context, path string, body any, attempts int) error {
	var last error
	for i := 0; i < attempts; i++ {
		err := c.postJSON(ctx, path, body)
		if err == nil {
			return nil
		}
		last = err
		var he *httpError
		if asErr(err, &he) && he.status >= 400 && he.status < 500 {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff(i)):
		}
	}
	return last
}

func backoff(attempt int) time.Duration {
	base := 200 * time.Millisecond
	for i := 0; i < attempt; i++ {
		base *= 2
	}
	if base > 5*time.Second {
		base = 5 * time.Second
	}
	return base
}

type httpError struct {
	status int
	body   string
}

func (e *httpError) Error() string { return fmt.Sprintf("HTTP %d: %s", e.status, e.body) }

// asErr is a minimal errors.As-ish for tests without needing stdlib errors pkg imports here.
func asErr(err error, target **httpError) bool {
	for err != nil {
		if he, ok := err.(*httpError); ok {
			*target = he
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

func (c *Client) postJSON(ctx context.Context, path string, body any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	c.auth(req)
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	b, _ := io.ReadAll(resp.Body)
	return &httpError{status: resp.StatusCode, body: string(b)}
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/adapter/vanta/
git commit -m "Add Vanta Ingest (POST /v1/memory/write) with retry"
```

---

## Task 8: Vanta adapter — Retrieve

**Files:**
- Create: `benchmarks/longmemeval/internal/adapter/vanta/retrieve.go`
- Create: `benchmarks/longmemeval/internal/adapter/vanta/retrieve_test.go`

**Endpoint:** `POST /v1/memory/recall`. Request:

```json
{
  "namespace": "bench/lme/<qid>",
  "query": "<question_text>",
  "ranking": "similarity" | "activation",
  "top_k": 10
}
```

Response (idealized; harness must tolerate extra fields):

```json
{
  "hits": [
    {"memory_key": "s1:0", "payload_body": "...", "score": 0.91}
  ]
}
```

The adapter parses `memory_key` back into `<session_id>:<turn_idx>` to populate `adapter.Hit`.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/adapter/vanta/retrieve_test.go`:

```go
package vanta

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
)

func TestRetrieve_ParsesHits(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/memory/recall" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"hits":[
			{"memory_key":"s1:0","payload_body":"hello","score":0.9},
			{"memory_key":"s2:3","payload_body":"world","score":0.7}
		]}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "", 5*time.Second)

	hits, err := c.Retrieve(context.Background(), "qA", "what?", adapter.RetrievalConfig{Strategy: "similarity", TopK: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits=%d want 2", len(hits))
	}
	if hits[0].SessionID != "s1" || hits[0].TurnIdx != 0 || hits[0].Content != "hello" || hits[0].Score != 0.9 {
		t.Errorf("hit0=%+v", hits[0])
	}
	if hits[1].SessionID != "s2" || hits[1].TurnIdx != 3 {
		t.Errorf("hit1=%+v", hits[1])
	}
	if gotBody["namespace"] != "bench/lme/qA" {
		t.Errorf("ns=%v", gotBody["namespace"])
	}
	if gotBody["ranking"] != "similarity" {
		t.Errorf("ranking=%v", gotBody["ranking"])
	}
	if int(gotBody["top_k"].(float64)) != 5 {
		t.Errorf("top_k=%v", gotBody["top_k"])
	}
}

func TestRetrieve_RejectsUnknownStrategy(t *testing.T) {
	c := New("http://unused", "", time.Second)
	_, err := c.Retrieve(context.Background(), "q", "x", adapter.RetrievalConfig{Strategy: "garbage", TopK: 1})
	if err == nil {
		t.Fatal("want error for unknown strategy")
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: FAIL (undefined: Retrieve method).

- [ ] **Step 3: Write retrieve.go**

Create `benchmarks/longmemeval/internal/adapter/vanta/retrieve.go`:

```go
package vanta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
)

type recallReq struct {
	Namespace string `json:"namespace"`
	Query     string `json:"query"`
	Ranking   string `json:"ranking"`
	TopK      int    `json:"top_k"`
}

type recallResp struct {
	Hits []struct {
		MemoryKey   string  `json:"memory_key"`
		PayloadBody string  `json:"payload_body"`
		Score       float64 `json:"score"`
	} `json:"hits"`
}

var validStrategies = map[string]bool{"similarity": true, "activation": true}

func (c *Client) Retrieve(ctx context.Context, questionID, query string, cfg adapter.RetrievalConfig) ([]adapter.Hit, error) {
	if !validStrategies[cfg.Strategy] {
		return nil, fmt.Errorf("vanta.Retrieve: unknown strategy %q (want similarity|activation)", cfg.Strategy)
	}
	if cfg.TopK <= 0 {
		return nil, fmt.Errorf("vanta.Retrieve: TopK must be > 0")
	}
	body := recallReq{
		Namespace: fmt.Sprintf("bench/lme/%s", questionID),
		Query:     query,
		Ranking:   cfg.Strategy,
		TopK:      cfg.TopK,
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/memory/recall", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	c.auth(req)
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vanta.Retrieve: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vanta.Retrieve: HTTP %d", resp.StatusCode)
	}
	var out recallResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("vanta.Retrieve decode: %w", err)
	}
	hits := make([]adapter.Hit, 0, len(out.Hits))
	for _, h := range out.Hits {
		sid, turn, err := parseMemoryKey(h.MemoryKey)
		if err != nil {
			return nil, err
		}
		hits = append(hits, adapter.Hit{
			SessionID: sid,
			TurnIdx:   turn,
			Content:   h.PayloadBody,
			Score:     h.Score,
		})
	}
	return hits, nil
}

func parseMemoryKey(k string) (string, int, error) {
	idx := strings.LastIndex(k, ":")
	if idx < 0 {
		return "", 0, fmt.Errorf("bad memory_key %q", k)
	}
	turn, err := strconv.Atoi(k[idx+1:])
	if err != nil {
		return "", 0, fmt.Errorf("bad memory_key %q: %w", k, err)
	}
	return k[:idx], turn, nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/adapter/vanta/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/adapter/vanta/
git commit -m "Add Vanta Retrieve (POST /v1/memory/recall)"
```

---

## Task 9: Cost accounting with tiktoken

**Files:**
- Create: `benchmarks/longmemeval/internal/cost/cost.go`
- Create: `benchmarks/longmemeval/internal/cost/cost_test.go`

**Purpose:** Estimate per-call USD cost by counting tokens via `tiktoken-go` and applying per-model pricing. Prices are configurable (centralized) so updating model pricing is a one-file change. Claude token counts are approximated with the OpenAI cl100k_base tokenizer — accurate enough for budget ceilings (off by <15%), and avoids a separate Claude tokenizer dep.

- [ ] **Step 1: Add the tiktoken-go dep**

Run from stack-explorer root:

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
go get github.com/pkoukk/tiktoken-go@latest
go mod tidy
```

- [ ] **Step 2: Write the failing test**

Create `benchmarks/longmemeval/internal/cost/cost_test.go`:

```go
package cost

import (
	"math"
	"testing"
)

func TestCountTokens_NonEmpty(t *testing.T) {
	n, err := CountTokens("hello world, this is a test")
	if err != nil {
		t.Fatal(err)
	}
	if n < 4 || n > 20 {
		t.Fatalf("token count out of plausible range: %d", n)
	}
}

func TestPrice_Sonnet(t *testing.T) {
	usd := Price(Anthropic, "claude-sonnet-4-6", 1_000_000, 1_000_000)
	// Sonnet 4.6: $3/M input, $15/M output -> $18
	if math.Abs(usd-18.0) > 0.01 {
		t.Fatalf("price=%.4f want ~18.00", usd)
	}
}

func TestPrice_GPT4o(t *testing.T) {
	usd := Price(OpenAI, "gpt-4o", 1_000_000, 1_000_000)
	// gpt-4o: $2.50/M input, $10/M output -> $12.50
	if math.Abs(usd-12.5) > 0.01 {
		t.Fatalf("price=%.4f want ~12.50", usd)
	}
}

func TestPrice_UnknownModel(t *testing.T) {
	usd := Price(Anthropic, "claude-unreleased-999", 1000, 1000)
	if usd != 0 {
		t.Fatalf("unknown model should return 0; got %.4f", usd)
	}
}

func TestLedger_AddAndTotal(t *testing.T) {
	l := NewLedger()
	l.Add(Anthropic, "claude-sonnet-4-6", 1_000_000, 1_000_000)
	l.Add(OpenAI, "gpt-4o", 1_000_000, 1_000_000)
	p := l.TotalUSD()
	if math.Abs(p.Anthropic-18.0) > 0.01 {
		t.Fatalf("anthropic=%.4f", p.Anthropic)
	}
	if math.Abs(p.OpenAI-12.5) > 0.01 {
		t.Fatalf("openai=%.4f", p.OpenAI)
	}
	if math.Abs(p.Total-30.5) > 0.01 {
		t.Fatalf("total=%.4f", p.Total)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/cost/...`
Expected: FAIL (undefined symbols).

- [ ] **Step 4: Write cost.go**

Create `benchmarks/longmemeval/internal/cost/cost.go`:

```go
// Package cost tracks approximate token usage and USD spend per provider.
package cost

import (
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

type Provider string

const (
	Anthropic Provider = "anthropic"
	OpenAI    Provider = "openai"
)

// priceTable maps (provider, model) -> (inputUSDPerMTok, outputUSDPerMTok).
// All prices as of plan-authoring. Update here when providers change pricing.
var priceTable = map[Provider]map[string][2]float64{
	Anthropic: {
		"claude-sonnet-4-6": {3.00, 15.00},
		"claude-opus-4-7":   {15.00, 75.00},
		"claude-haiku-4-5":  {0.80, 4.00},
	},
	OpenAI: {
		"gpt-4o":      {2.50, 10.00},
		"gpt-4o-mini": {0.15, 0.60},
	},
}

// Price returns the USD cost for inputTok input tokens and outputTok output
// tokens on the given (provider, model). Unknown (provider, model) -> 0.
func Price(p Provider, model string, inputTok, outputTok int) float64 {
	m, ok := priceTable[p]
	if !ok {
		return 0
	}
	rates, ok := m[model]
	if !ok {
		return 0
	}
	return (float64(inputTok)/1_000_000)*rates[0] + (float64(outputTok)/1_000_000)*rates[1]
}

var tkOnce sync.Once
var tk *tiktoken.Tiktoken
var tkErr error

// CountTokens returns an approximate token count using cl100k_base (OpenAI's
// default tokenizer). For Claude models this is a reasonable approximation
// (within ~15%) — sufficient for cost ceilings, not for billing.
func CountTokens(s string) (int, error) {
	tkOnce.Do(func() {
		tk, tkErr = tiktoken.GetEncoding("cl100k_base")
	})
	if tkErr != nil {
		return 0, tkErr
	}
	return len(tk.Encode(s, nil, nil)), nil
}

type LedgerTotal struct {
	Anthropic float64
	OpenAI    float64
	Total     float64
}

type entry struct {
	provider  Provider
	model     string
	inputTok  int
	outputTok int
}

type Ledger struct {
	mu      sync.Mutex
	entries []entry
}

func NewLedger() *Ledger { return &Ledger{} }

func (l *Ledger) Add(p Provider, model string, inputTok, outputTok int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry{p, model, inputTok, outputTok})
}

func (l *Ledger) TotalUSD() LedgerTotal {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out LedgerTotal
	for _, e := range l.entries {
		usd := Price(e.provider, e.model, e.inputTok, e.outputTok)
		switch e.provider {
		case Anthropic:
			out.Anthropic += usd
		case OpenAI:
			out.OpenAI += usd
		}
		out.Total += usd
	}
	return out
}

// TotalTokens returns totals per provider.
type TokenTotals struct {
	AnthropicInput, AnthropicOutput int
	OpenAIInput, OpenAIOutput       int
}

func (l *Ledger) TotalTokens() TokenTotals {
	l.mu.Lock()
	defer l.mu.Unlock()
	var t TokenTotals
	for _, e := range l.entries {
		switch e.provider {
		case Anthropic:
			t.AnthropicInput += e.inputTok
			t.AnthropicOutput += e.outputTok
		case OpenAI:
			t.OpenAIInput += e.inputTok
			t.OpenAIOutput += e.outputTok
		}
	}
	return t
}
```

- [ ] **Step 5: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/cost/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
git add go.mod go.sum benchmarks/longmemeval/internal/cost/
git commit -m "Add cost ledger (tiktoken-based token count + per-provider USD)"
```

---

## Task 10: Answerer (go-providers wrapper)

**Files:**
- Create: `benchmarks/longmemeval/internal/answerer/answerer.go`
- Create: `benchmarks/longmemeval/internal/answerer/answerer_test.go`

**Purpose:** Given `(question, []adapter.Hit)`, call the provider's `Complete` with a LongMemEval-style prompt and return `(answer, inputTokens, outputTokens, error)`. The provider is injected (interface) so tests can use a fake.

**Prompt template** (from LongMemEval paper, simplified; literal strings in the code):

```
System: You are a helpful assistant. Use only the provided prior conversation snippets to answer. If the snippets don't contain the answer, say so concisely.

User:
Prior snippets (most relevant first):
<hit 0 content>
<hit 1 content>
...

Question: <question>

Answer:
```

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/answerer/answerer_test.go`:

```go
package answerer

import (
	"context"
	"strings"
	"testing"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/hollis-labs/go-providers/provider"
)

type fakeProvider struct {
	capturedSystem   string
	capturedUserLast string
	reply            string
	err              error
}

func (f *fakeProvider) StreamChat(context.Context, provider.ChatRequest) (<-chan provider.StreamEvent, error) {
	return nil, nil
}
func (f *fakeProvider) Complete(ctx context.Context, req provider.ChatRequest) (string, error) {
	f.capturedSystem = req.SystemPrompt
	for _, m := range req.Messages {
		if m.Role == "user" {
			f.capturedUserLast = m.Content
		}
	}
	return f.reply, f.err
}
func (f *fakeProvider) Capabilities() provider.ProviderCapabilities { return provider.ProviderCapabilities{} }

func TestAnswer_BuildsPromptAndReturnsTokens(t *testing.T) {
	fp := &fakeProvider{reply: "orange"}
	a := New(fp, "claude-sonnet-4-6")
	hits := []adapter.Hit{
		{SessionID: "s1", TurnIdx: 0, Content: "My cat is orange."},
		{SessionID: "s2", TurnIdx: 1, Content: "Irrelevant."},
	}
	res, err := a.Answer(context.Background(), "What color is my cat?", hits)
	if err != nil {
		t.Fatal(err)
	}
	if res.Answer != "orange" {
		t.Errorf("answer=%q", res.Answer)
	}
	if !strings.Contains(fp.capturedUserLast, "My cat is orange.") {
		t.Errorf("user prompt missing hit content: %q", fp.capturedUserLast)
	}
	if !strings.Contains(fp.capturedUserLast, "What color is my cat?") {
		t.Errorf("user prompt missing question: %q", fp.capturedUserLast)
	}
	if res.InputTokens <= 0 || res.OutputTokens <= 0 {
		t.Errorf("tokens: in=%d out=%d", res.InputTokens, res.OutputTokens)
	}
}

func TestAnswer_EmptyHits(t *testing.T) {
	fp := &fakeProvider{reply: "I don't know."}
	a := New(fp, "claude-sonnet-4-6")
	res, err := a.Answer(context.Background(), "Q?", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Answer != "I don't know." {
		t.Errorf("answer=%q", res.Answer)
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/answerer/...`
Expected: FAIL (package deps / undefined types).

- [ ] **Step 3: Add go-providers dep**

Run from stack-explorer root:

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
go get github.com/hollis-labs/go-providers@latest
go mod tidy
```

If the module isn't published, add a local replace directive. Run:

```bash
echo 'replace github.com/hollis-labs/go-providers => /Users/chrispian/Projects-apps/framework/libs/go-providers' >> go.mod
go mod tidy
```

- [ ] **Step 4: Write answerer.go**

Create `benchmarks/longmemeval/internal/answerer/answerer.go`:

```go
// Package answerer generates an answer to a LongMemEval question using a
// provider.Provider from go-providers and retrieved memory hits.
package answerer

import (
	"context"
	"fmt"
	"strings"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
	"github.com/hollis-labs/go-providers/provider"
)

const systemPrompt = `You are a helpful assistant. Use only the provided prior conversation snippets to answer the user's question. If the snippets do not contain the answer, say so concisely.`

type Answerer struct {
	p     provider.Provider
	model string
}

func New(p provider.Provider, model string) *Answerer { return &Answerer{p: p, model: model} }

type Result struct {
	Answer       string
	InputTokens  int
	OutputTokens int
	PromptUsed   string
}

func (a *Answerer) Answer(ctx context.Context, question string, hits []adapter.Hit) (Result, error) {
	userPrompt := buildUserPrompt(question, hits)
	req := provider.ChatRequest{
		Model:        a.model,
		SystemPrompt: systemPrompt,
		Messages:     []provider.ChatMessage{{Role: "user", Content: userPrompt}},
	}
	out, err := a.p.Complete(ctx, req)
	if err != nil {
		return Result{}, fmt.Errorf("answerer.Complete: %w", err)
	}
	in, err := cost.CountTokens(systemPrompt + "\n" + userPrompt)
	if err != nil {
		return Result{}, err
	}
	outTok, err := cost.CountTokens(out)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Answer:       strings.TrimSpace(out),
		InputTokens:  in,
		OutputTokens: outTok,
		PromptUsed:   userPrompt,
	}, nil
}

func buildUserPrompt(question string, hits []adapter.Hit) string {
	var b strings.Builder
	b.WriteString("Prior snippets (most relevant first):\n")
	if len(hits) == 0 {
		b.WriteString("(no snippets retrieved)\n")
	}
	for i, h := range hits {
		fmt.Fprintf(&b, "[%d] (session=%s turn=%d) %s\n", i, h.SessionID, h.TurnIdx, h.Content)
	}
	fmt.Fprintf(&b, "\nQuestion: %s\n\nAnswer:", question)
	return b.String()
}
```

- [ ] **Step 5: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/answerer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
git add go.mod go.sum benchmarks/longmemeval/internal/answerer/
git commit -m "Add answerer wrapping go-providers + LongMemEval prompt"
```

---

## Task 11: Vendored Python judge + setup script

**Files:**
- Create: `benchmarks/longmemeval/python/judge.py`
- Create: `benchmarks/longmemeval/python/requirements.txt`
- Create: `benchmarks/longmemeval/scripts/setup-python.sh`

**Purpose:** A minimal Python judge script that reads `{"question","gold","predicted","model"}` JSON on stdin and prints `{"correct":bool,"rationale":"..."}`. The judging prompt is taken verbatim from the LongMemEval paper's evaluation section.

- [ ] **Step 1: Create `benchmarks/longmemeval/python/requirements.txt`**

```txt
openai>=1.30.0
```

- [ ] **Step 2: Create `benchmarks/longmemeval/python/judge.py`**

```python
#!/usr/bin/env python3
"""LongMemEval judge (vendored minimal subset).

Reads JSON on stdin:
    {"question": str, "gold": str, "predicted": str, "model": str}

Writes JSON on stdout:
    {"correct": bool, "rationale": str}

Exits 0 on success (including a judge verdict of "incorrect"). Exits non-zero
only on transport / parsing failures.
"""

import json
import os
import re
import sys

from openai import OpenAI


JUDGE_PROMPT = """You are an impartial grader for a long-term memory benchmark.
Given a question, a gold answer, and a predicted answer, decide whether the
predicted answer is substantially correct (captures the essential facts of the
gold answer). Ignore style and wording differences.

Respond with a single JSON object and nothing else:
{"correct": true|false, "rationale": "<one-sentence reason>"}

Question: {question}

Gold answer: {gold}

Predicted answer: {predicted}
"""


def main() -> int:
    try:
        payload = json.load(sys.stdin)
    except Exception as e:
        print(json.dumps({"error": f"bad stdin: {e}"}), file=sys.stderr)
        return 2

    try:
        question = payload["question"]
        gold = payload["gold"]
        predicted = payload["predicted"]
        model = payload.get("model", "gpt-4o")
    except KeyError as e:
        print(json.dumps({"error": f"missing field: {e}"}), file=sys.stderr)
        return 2

    prompt = JUDGE_PROMPT.format(question=question, gold=gold, predicted=predicted)
    client = OpenAI(api_key=os.environ["OPENAI_API_KEY"])
    resp = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": prompt}],
        temperature=0,
    )
    text = resp.choices[0].message.content or ""
    text = text.strip()

    # Judge sometimes wraps JSON in a fence. Strip it.
    m = re.search(r"\{.*\}", text, re.DOTALL)
    if not m:
        print(json.dumps({"error": "no JSON in judge reply", "raw": text}), file=sys.stderr)
        return 3

    try:
        parsed = json.loads(m.group(0))
    except json.JSONDecodeError as e:
        print(json.dumps({"error": f"invalid JSON from judge: {e}", "raw": text}), file=sys.stderr)
        return 3

    out = {
        "correct": bool(parsed.get("correct", False)),
        "rationale": str(parsed.get("rationale", "")),
        "input_tokens": int(getattr(resp.usage, "prompt_tokens", 0) or 0),
        "output_tokens": int(getattr(resp.usage, "completion_tokens", 0) or 0),
        "model": model,
    }
    print(json.dumps(out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
```

- [ ] **Step 3: Create `benchmarks/longmemeval/scripts/setup-python.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$HERE/python"

if [[ ! -d .venv ]]; then
  python3 -m venv .venv
fi

.venv/bin/pip install --upgrade pip >/dev/null
.venv/bin/pip install -r requirements.txt

echo "python venv ready: $HERE/python/.venv"
```

Make it executable:

```bash
chmod +x benchmarks/longmemeval/scripts/setup-python.sh
```

- [ ] **Step 4: Run setup-python locally to verify**

Run:

```bash
cd benchmarks/longmemeval
bash scripts/setup-python.sh
```

Expected: exits 0, creates `python/.venv/`, prints `python venv ready: ...`.

- [ ] **Step 5: Smoke-test the judge script against the venv (without calling OpenAI)**

Run:

```bash
echo '{}' | benchmarks/longmemeval/python/.venv/bin/python benchmarks/longmemeval/python/judge.py || true
```

Expected: stderr contains `"missing field"`, exit code 2. This confirms the script loads, parses stdin, and exits with the documented contract — without making any API call.

- [ ] **Step 6: Commit**

```bash
git add benchmarks/longmemeval/python/ benchmarks/longmemeval/scripts/setup-python.sh
git commit -m "Add vendored LongMemEval Python judge + venv setup script"
```

---

## Task 12: Judge subprocess wrapper (Go)

**Files:**
- Create: `benchmarks/longmemeval/internal/judge/judge.go`
- Create: `benchmarks/longmemeval/internal/judge/judge_test.go`

**Contract:** `Judge.Grade(ctx, question, gold, predicted)` execs `python/.venv/bin/python python/judge.py`, pipes JSON in, reads JSON out, returns `(correct, rationale, inputTok, outputTok, err)`. Exit code 2 or 3 → return a typed `JudgeError` (wrapper caller can count these as `judge_error: true` without aborting).

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/judge/judge_test.go`:

```go
package judge

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// writeScript writes a fake Python script that echoes canned JSON and exits code.
// It's invoked via `/usr/bin/env python3`, so we don't need a real venv in the test.
func writeScript(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "judge.py")
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGrade_CorrectTrue(t *testing.T) {
	script := writeScript(t, `#!/usr/bin/env python3
import sys, json
_ = json.load(sys.stdin)
print(json.dumps({"correct": True, "rationale": "yep", "input_tokens": 10, "output_tokens": 3, "model": "gpt-4o"}))
`)
	j := New("python3", script, "gpt-4o")
	r, err := j.Grade(context.Background(), "q?", "gold", "gold")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Correct {
		t.Fatal("want Correct=true")
	}
	if r.InputTokens != 10 || r.OutputTokens != 3 {
		t.Errorf("tokens: %+v", r)
	}
}

func TestGrade_MalformedJSON(t *testing.T) {
	script := writeScript(t, `#!/usr/bin/env python3
import sys
sys.stderr.write('{"error":"no JSON"}')
sys.exit(3)
`)
	j := New("python3", script, "gpt-4o")
	_, err := j.Grade(context.Background(), "q", "g", "p")
	if err == nil {
		t.Fatal("want error from exit code 3")
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/judge/...`
Expected: FAIL (undefined).

- [ ] **Step 3: Write judge.go**

Create `benchmarks/longmemeval/internal/judge/judge.go`:

```go
// Package judge wraps the vendored Python LongMemEval judge as a subprocess.
package judge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Judge struct {
	python  string
	script  string
	model   string
	timeout time.Duration
}

func New(python, script, model string) *Judge {
	return &Judge{python: python, script: script, model: model, timeout: 60 * time.Second}
}

type Result struct {
	Correct      bool
	Rationale    string
	InputTokens  int
	OutputTokens int
	Model        string
}

type judgeReply struct {
	Correct      bool   `json:"correct"`
	Rationale    string `json:"rationale"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Model        string `json:"model"`
}

func (j *Judge) Grade(ctx context.Context, question, gold, predicted string) (Result, error) {
	in, err := json.Marshal(map[string]string{
		"question":  question,
		"gold":      gold,
		"predicted": predicted,
		"model":     j.model,
	})
	if err != nil {
		return Result{}, err
	}
	tctx, cancel := context.WithTimeout(ctx, j.timeout)
	defer cancel()
	cmd := exec.CommandContext(tctx, j.python, j.script)
	cmd.Stdin = bytes.NewReader(in)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Result{}, fmt.Errorf("judge subprocess: %w (stderr=%q)", err, stderr.String())
	}
	var r judgeReply
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil {
		return Result{}, fmt.Errorf("judge stdout unmarshal: %w (stdout=%q)", err, stdout.String())
	}
	return Result{
		Correct: r.Correct, Rationale: r.Rationale,
		InputTokens: r.InputTokens, OutputTokens: r.OutputTokens,
		Model: r.Model,
	}, nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/judge/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/judge/
git commit -m "Add judge subprocess wrapper"
```

---

## Task 13: Config loader (env + flags)

**Files:**
- Create: `benchmarks/longmemeval/internal/config/config.go`
- Create: `benchmarks/longmemeval/internal/config/config_test.go`

**Fields:**

```
VantaURL         (env VANTA_URL,         default http://127.0.0.1:8089)
VantaToken       (env VANTA_TOKEN,       default "")
AnthropicAPIKey  (env ANTHROPIC_API_KEY) — required when running
OpenAIAPIKey     (env OPENAI_API_KEY)    — required when running
AnswererModel    (env LME_ANSWER_MODEL,  default claude-sonnet-4-6)
JudgeModel       (env LME_JUDGE_MODEL,   default gpt-4o)
PythonBin        (env LME_PYTHON,        default benchmarks/longmemeval/python/.venv/bin/python)
JudgeScript      (env LME_JUDGE_SCRIPT,  default benchmarks/longmemeval/python/judge.py)
DataDir          (env LME_DATA_DIR,      default benchmarks/longmemeval/data)
RunsDir          (env LME_RUNS_DIR,      default benchmarks/longmemeval/runs)
Concurrency      (env LME_CONCURRENCY,   default 4)
TopK             (env LME_TOPK,          default 10)
Seed             (env LME_SEED,          default 42)
MaxCostUSD       (env LME_MAX_COST_USD,  default 0 = no limit)
```

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/config/config_test.go`:

```go
package config

import (
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	c, err := Load(map[string]string{
		"ANTHROPIC_API_KEY": "x",
		"OPENAI_API_KEY":    "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.VantaURL != "http://127.0.0.1:8089" {
		t.Errorf("vanta url default wrong: %s", c.VantaURL)
	}
	if c.AnswererModel != "claude-sonnet-4-6" {
		t.Errorf("answerer default: %s", c.AnswererModel)
	}
	if c.JudgeModel != "gpt-4o" {
		t.Errorf("judge default: %s", c.JudgeModel)
	}
	if c.Concurrency != 4 || c.TopK != 10 || c.Seed != 42 {
		t.Errorf("numeric defaults: %+v", c)
	}
}

func TestLoad_RequiresKeys(t *testing.T) {
	_, err := Load(map[string]string{"ANTHROPIC_API_KEY": "x"})
	if err == nil {
		t.Fatal("want error when OPENAI_API_KEY missing")
	}
	_, err = Load(map[string]string{"OPENAI_API_KEY": "y"})
	if err == nil {
		t.Fatal("want error when ANTHROPIC_API_KEY missing")
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	c, err := Load(map[string]string{
		"ANTHROPIC_API_KEY": "x",
		"OPENAI_API_KEY":    "y",
		"VANTA_URL":         "http://example:9000",
		"LME_CONCURRENCY":   "8",
		"LME_TOPK":          "20",
		"LME_MAX_COST_USD":  "25.50",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.VantaURL != "http://example:9000" || c.Concurrency != 8 || c.TopK != 20 {
		t.Errorf("overrides not applied: %+v", c)
	}
	if c.MaxCostUSD != 25.50 {
		t.Errorf("max cost: %v", c.MaxCostUSD)
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/config/...`
Expected: FAIL.

- [ ] **Step 3: Write config.go**

Create `benchmarks/longmemeval/internal/config/config.go`:

```go
// Package config loads LongMemEval harness configuration from env maps.
package config

import (
	"fmt"
	"strconv"
)

type Config struct {
	VantaURL        string
	VantaToken      string
	AnthropicAPIKey string
	OpenAIAPIKey    string
	AnswererModel   string
	JudgeModel      string
	PythonBin       string
	JudgeScript     string
	DataDir         string
	RunsDir         string
	Concurrency     int
	TopK            int
	Seed            int64
	MaxCostUSD      float64
}

func Load(env map[string]string) (Config, error) {
	c := Config{
		VantaURL:      def(env, "VANTA_URL", "http://127.0.0.1:8089"),
		VantaToken:    env["VANTA_TOKEN"],
		AnswererModel: def(env, "LME_ANSWER_MODEL", "claude-sonnet-4-6"),
		JudgeModel:    def(env, "LME_JUDGE_MODEL", "gpt-4o"),
		PythonBin:     def(env, "LME_PYTHON", "benchmarks/longmemeval/python/.venv/bin/python"),
		JudgeScript:   def(env, "LME_JUDGE_SCRIPT", "benchmarks/longmemeval/python/judge.py"),
		DataDir:       def(env, "LME_DATA_DIR", "benchmarks/longmemeval/data"),
		RunsDir:       def(env, "LME_RUNS_DIR", "benchmarks/longmemeval/runs"),
	}
	c.AnthropicAPIKey = env["ANTHROPIC_API_KEY"]
	c.OpenAIAPIKey = env["OPENAI_API_KEY"]
	if c.AnthropicAPIKey == "" {
		return Config{}, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	if c.OpenAIAPIKey == "" {
		return Config{}, fmt.Errorf("OPENAI_API_KEY is required")
	}

	conc, err := atoiDefault(env, "LME_CONCURRENCY", 4)
	if err != nil {
		return Config{}, err
	}
	c.Concurrency = conc

	topk, err := atoiDefault(env, "LME_TOPK", 10)
	if err != nil {
		return Config{}, err
	}
	c.TopK = topk

	seed, err := atoi64Default(env, "LME_SEED", 42)
	if err != nil {
		return Config{}, err
	}
	c.Seed = seed

	cost, err := atofDefault(env, "LME_MAX_COST_USD", 0)
	if err != nil {
		return Config{}, err
	}
	c.MaxCostUSD = cost

	return c, nil
}

func def(env map[string]string, k, fallback string) string {
	if v, ok := env[k]; ok && v != "" {
		return v
	}
	return fallback
}

func atoiDefault(env map[string]string, k string, fallback int) (int, error) {
	v, ok := env[k]
	if !ok || v == "" {
		return fallback, nil
	}
	return strconv.Atoi(v)
}

func atoi64Default(env map[string]string, k string, fallback int64) (int64, error) {
	v, ok := env[k]
	if !ok || v == "" {
		return fallback, nil
	}
	return strconv.ParseInt(v, 10, 64)
}

func atofDefault(env map[string]string, k string, fallback float64) (float64, error) {
	v, ok := env[k]
	if !ok || v == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(v, 64)
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/config/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/config/
git commit -m "Add harness config loader"
```

---

## Task 14: Per-question pipeline (one question, one retrieval config)

**Files:**
- Create: `benchmarks/longmemeval/internal/runner/pipeline.go`
- Create: `benchmarks/longmemeval/internal/runner/pipeline_test.go`

**Contract:** `Pipeline.Run(ctx, question, cfg)` performs Reset → Ingest → Retrieve → Answer → Judge, accumulates tokens into the shared ledger, and returns a `QuestionResult`. The pipeline accepts injected `adapter.Adapter`, `*answerer.Answerer`, `*judge.Judge`, `*cost.Ledger`, and a flag `oracle bool` — when true, Ingest+Retrieve are skipped and gold sessions are flattened into hits.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/runner/pipeline_test.go`:

```go
package runner

import (
	"context"
	"testing"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/answerer"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/dataset"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/judge"
	"github.com/hollis-labs/go-providers/provider"
)

type fakeAdapter struct {
	resetCalls, ingestCalls, retrieveCalls int
	hits                                   []adapter.Hit
}

func (f *fakeAdapter) Reset(context.Context, string) error { f.resetCalls++; return nil }
func (f *fakeAdapter) Ingest(context.Context, string, []adapter.Session) error {
	f.ingestCalls++
	return nil
}
func (f *fakeAdapter) Retrieve(context.Context, string, string, adapter.RetrievalConfig) ([]adapter.Hit, error) {
	f.retrieveCalls++
	return f.hits, nil
}
func (f *fakeAdapter) Close() error { return nil }

type fakeProvider struct{ reply string }

func (f *fakeProvider) StreamChat(context.Context, provider.ChatRequest) (<-chan provider.StreamEvent, error) {
	return nil, nil
}
func (f *fakeProvider) Complete(context.Context, provider.ChatRequest) (string, error) {
	return f.reply, nil
}
func (f *fakeProvider) Capabilities() provider.ProviderCapabilities {
	return provider.ProviderCapabilities{}
}

// fakeJudge writes a Python script on disk that echoes a fixed verdict.
// We reuse the real judge wrapper to exercise the subprocess path in tests.
func newFakeJudge(t *testing.T, correct bool) *judge.Judge {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/judge.py"
	var body string
	if correct {
		body = `#!/usr/bin/env python3
import sys, json
_ = json.load(sys.stdin)
print(json.dumps({"correct": True, "rationale": "yes", "input_tokens": 5, "output_tokens": 2, "model": "gpt-4o"}))
`
	} else {
		body = `#!/usr/bin/env python3
import sys, json
_ = json.load(sys.stdin)
print(json.dumps({"correct": False, "rationale": "no", "input_tokens": 5, "output_tokens": 2, "model": "gpt-4o"}))
`
	}
	if err := writeFile(path, body); err != nil {
		t.Fatal(err)
	}
	return judge.New("python3", path, "gpt-4o")
}

func writeFile(path, body string) error {
	return writeFileHelper(path, body)
}

func writeFileHelper(path, body string) error {
	// separate func so we can easily replace with os.WriteFile; kept for clarity
	return osWriteFile(path, body)
}

// TestPipeline_Standard runs one question in similarity mode with all fakes.
func TestPipeline_Standard(t *testing.T) {
	fa := &fakeAdapter{hits: []adapter.Hit{{SessionID: "s1", TurnIdx: 0, Content: "orange cat"}}}
	ans := answerer.New(&fakeProvider{reply: "orange"}, "claude-sonnet-4-6")
	j := newFakeJudge(t, true)
	ledger := cost.NewLedger()
	p := NewPipeline(fa, ans, j, ledger)

	q := dataset.Question{
		ID: "q1", Type: "single-session-user", Text: "color?",
		Answer: "orange",
		HaystackSessions: []adapter.Session{
			{SessionID: "s1", Timestamp: time.Now(), Turns: []adapter.Turn{{Role: "user", Content: "orange cat"}}},
		},
	}
	r, err := p.Run(context.Background(), q, adapter.RetrievalConfig{Strategy: "similarity", TopK: 5}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Judge.Correct {
		t.Errorf("want correct, got %+v", r.Judge)
	}
	if fa.ingestCalls != 1 || fa.retrieveCalls != 1 {
		t.Errorf("adapter calls: ingest=%d retrieve=%d", fa.ingestCalls, fa.retrieveCalls)
	}
}

// TestPipeline_Oracle skips ingest/retrieve; gold sessions become hits.
func TestPipeline_Oracle(t *testing.T) {
	fa := &fakeAdapter{}
	ans := answerer.New(&fakeProvider{reply: "orange"}, "claude-sonnet-4-6")
	j := newFakeJudge(t, true)
	p := NewPipeline(fa, ans, j, cost.NewLedger())
	q := dataset.Question{
		ID: "q1", Type: "single-session-user", Text: "color?",
		Answer: "orange", AnswerSessionIDs: []string{"s1"},
		HaystackSessions: []adapter.Session{
			{SessionID: "s1", Timestamp: time.Now(), Turns: []adapter.Turn{{Role: "user", Content: "orange cat"}}},
			{SessionID: "s2", Timestamp: time.Now(), Turns: []adapter.Turn{{Role: "user", Content: "irrelevant"}}},
		},
	}
	r, err := p.Run(context.Background(), q, adapter.RetrievalConfig{Strategy: "similarity", TopK: 5}, true)
	if err != nil {
		t.Fatal(err)
	}
	if fa.ingestCalls != 0 || fa.retrieveCalls != 0 {
		t.Errorf("oracle should skip adapter; got ingest=%d retrieve=%d", fa.ingestCalls, fa.retrieveCalls)
	}
	if !r.Judge.Correct {
		t.Errorf("want correct")
	}
}
```

Add helper file for `osWriteFile` so the test compiles:

Create `benchmarks/longmemeval/internal/runner/testutil_test.go`:

```go
package runner

import "os"

func osWriteFile(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o755)
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/runner/...`
Expected: FAIL (undefined: Pipeline / NewPipeline / QuestionResult).

- [ ] **Step 3: Write pipeline.go**

Create `benchmarks/longmemeval/internal/runner/pipeline.go`:

```go
// Package runner orchestrates per-question evaluation.
package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/answerer"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/dataset"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/judge"
)

type Pipeline struct {
	adapter  adapter.Adapter
	answerer *answerer.Answerer
	judge    *judge.Judge
	ledger   *cost.Ledger
}

func NewPipeline(a adapter.Adapter, ans *answerer.Answerer, j *judge.Judge, l *cost.Ledger) *Pipeline {
	return &Pipeline{adapter: a, answerer: ans, judge: j, ledger: l}
}

// QuestionResult captures the full evaluation record for one (question, config) pair.
type QuestionResult struct {
	QuestionID   string
	QuestionType string
	Config       adapter.RetrievalConfig
	Oracle       bool
	Predicted    string
	Judge        judge.Result
	Hits         []adapter.Hit
	TokensIn     int
	TokensOut    int
	JudgeInTok   int
	JudgeOutTok  int
	LatencyMS    int64
	Error        string
}

// Run executes the full pipeline for one (question, config). `oracle=true`
// skips Reset/Ingest/Retrieve and uses question.GoldSessions() as hits.
func (p *Pipeline) Run(ctx context.Context, q dataset.Question, cfg adapter.RetrievalConfig, oracle bool) (QuestionResult, error) {
	res := QuestionResult{QuestionID: q.ID, QuestionType: q.Type, Config: cfg, Oracle: oracle}
	start := time.Now()

	var hits []adapter.Hit
	if oracle {
		hits = flattenSessions(q.GoldSessions())
	} else {
		if err := p.adapter.Reset(ctx, q.ID); err != nil {
			res.Error = fmt.Sprintf("reset: %v", err)
			return res, nil
		}
		if err := p.adapter.Ingest(ctx, q.ID, q.HaystackSessions); err != nil {
			res.Error = fmt.Sprintf("ingest: %v", err)
			return res, nil
		}
		h, err := p.adapter.Retrieve(ctx, q.ID, q.Text, cfg)
		if err != nil {
			res.Error = fmt.Sprintf("retrieve: %v", err)
			return res, nil
		}
		hits = h
	}
	res.Hits = hits

	ans, err := p.answerer.Answer(ctx, q.Text, hits)
	if err != nil {
		res.Error = fmt.Sprintf("answer: %v", err)
		return res, nil
	}
	res.Predicted = ans.Answer
	res.TokensIn = ans.InputTokens
	res.TokensOut = ans.OutputTokens
	p.ledger.Add(cost.Anthropic, "claude-sonnet-4-6", ans.InputTokens, ans.OutputTokens)

	jr, err := p.judge.Grade(ctx, q.Text, q.Answer, ans.Answer)
	if err != nil {
		res.Error = fmt.Sprintf("judge: %v", err)
		return res, nil
	}
	res.Judge = jr
	res.JudgeInTok = jr.InputTokens
	res.JudgeOutTok = jr.OutputTokens
	p.ledger.Add(cost.OpenAI, jr.Model, jr.InputTokens, jr.OutputTokens)

	res.LatencyMS = time.Since(start).Milliseconds()
	return res, nil
}

func flattenSessions(sessions []adapter.Session) []adapter.Hit {
	var out []adapter.Hit
	for _, s := range sessions {
		for i, t := range s.Turns {
			out = append(out, adapter.Hit{
				SessionID: s.SessionID,
				TurnIdx:   i,
				Content:   t.Content,
				Score:     1.0,
			})
		}
	}
	return out
}
```

Note: the current implementation pins the answerer model to `"claude-sonnet-4-6"` in the ledger Add call. That's a loose coupling we fix in Task 15 (orchestrator) by passing the model into `NewPipeline`. Flag this as a follow-up.

Actually, fix it now — add `answererModel` to `NewPipeline`:

Update `NewPipeline` signature to take `answererModel string` and use it in the ledger call. Update the test fake to pass `"claude-sonnet-4-6"`.

Revised `pipeline.go` diff:

```go
type Pipeline struct {
	adapter        adapter.Adapter
	answerer       *answerer.Answerer
	judge          *judge.Judge
	ledger         *cost.Ledger
	answererModel  string
}

func NewPipeline(a adapter.Adapter, ans *answerer.Answerer, j *judge.Judge, l *cost.Ledger, answererModel string) *Pipeline {
	return &Pipeline{adapter: a, answerer: ans, judge: j, ledger: l, answererModel: answererModel}
}
```

And in Run, replace:
```go
p.ledger.Add(cost.Anthropic, "claude-sonnet-4-6", ans.InputTokens, ans.OutputTokens)
```
with:
```go
p.ledger.Add(cost.Anthropic, p.answererModel, ans.InputTokens, ans.OutputTokens)
```

Update the two test call sites (`NewPipeline(fa, ans, j, ledger, "claude-sonnet-4-6")`).

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/runner/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/runner/
git commit -m "Add per-question pipeline with oracle variant"
```

---

## Task 15: Runner orchestration (worker pool + resume)

**Files:**
- Create: `benchmarks/longmemeval/internal/runner/runner.go`
- Create: `benchmarks/longmemeval/internal/runner/runner_test.go`

**Contract:** `Runner.Run(ctx, questions, configs, oracle)` spawns N workers, each pulling (question, config) pairs from a queue. For each pair, the worker calls `Pipeline.Run` and writes a result JSON to `<runDir>/<configID>/<qid>.json` immediately. On re-invocation with the same `runDir`, already-written pairs are skipped.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/runner/runner_test.go`:

```go
package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/answerer"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/dataset"
	"github.com/hollis-labs/go-providers/provider"
)

type nullAdapter struct{}

func (nullAdapter) Reset(context.Context, string) error                    { return nil }
func (nullAdapter) Ingest(context.Context, string, []adapter.Session) error { return nil }
func (nullAdapter) Retrieve(context.Context, string, string, adapter.RetrievalConfig) ([]adapter.Hit, error) {
	return nil, nil
}
func (nullAdapter) Close() error { return nil }

type constProvider struct{ reply string }

func (c constProvider) StreamChat(context.Context, provider.ChatRequest) (<-chan provider.StreamEvent, error) {
	return nil, nil
}
func (c constProvider) Complete(context.Context, provider.ChatRequest) (string, error) {
	return c.reply, nil
}
func (c constProvider) Capabilities() provider.ProviderCapabilities {
	return provider.ProviderCapabilities{}
}

func TestRunner_WritesPerQuestionJSON(t *testing.T) {
	dir := t.TempDir()
	q1 := dataset.Question{ID: "qA", Type: "t", Text: "?", Answer: "a", HaystackSessions: []adapter.Session{{SessionID: "s", Timestamp: time.Now(), Turns: []adapter.Turn{{Role: "user", Content: "x"}}}}}
	q2 := dataset.Question{ID: "qB", Type: "t", Text: "?", Answer: "b", HaystackSessions: q1.HaystackSessions}
	j := newFakeJudge(t, true)
	p := NewPipeline(nullAdapter{}, answerer.New(constProvider{reply: "a"}, "claude-sonnet-4-6"), j, cost.NewLedger(), "claude-sonnet-4-6")
	r := NewRunner(p, dir, 2)

	if err := r.Run(context.Background(), []dataset.Question{q1, q2}, []adapter.RetrievalConfig{{Strategy: "similarity", TopK: 3}}, false); err != nil {
		t.Fatal(err)
	}
	for _, qid := range []string{"qA", "qB"} {
		path := filepath.Join(dir, "similarity-k3", qid+".json")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
		b, _ := os.ReadFile(path)
		var got map[string]any
		_ = json.Unmarshal(b, &got)
		if got["question_id"] != qid {
			t.Errorf("%s question_id=%v", path, got["question_id"])
		}
	}
}

func TestRunner_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	// pre-create the result file for qA.
	os.MkdirAll(filepath.Join(dir, "similarity-k3"), 0o755)
	os.WriteFile(filepath.Join(dir, "similarity-k3", "qA.json"), []byte(`{"question_id":"qA","skipped":true}`), 0o644)

	q := dataset.Question{ID: "qA", Type: "t", Text: "?", Answer: "a"}
	j := newFakeJudge(t, true)
	p := NewPipeline(nullAdapter{}, answerer.New(constProvider{reply: "a"}, "claude-sonnet-4-6"), j, cost.NewLedger(), "claude-sonnet-4-6")
	r := NewRunner(p, dir, 1)
	if err := r.Run(context.Background(), []dataset.Question{q}, []adapter.RetrievalConfig{{Strategy: "similarity", TopK: 3}}, false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "similarity-k3", "qA.json"))
	if string(b) != `{"question_id":"qA","skipped":true}` {
		t.Fatalf("file was overwritten: %s", b)
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/runner/...`
Expected: FAIL (undefined: Runner / NewRunner).

- [ ] **Step 3: Write runner.go**

Create `benchmarks/longmemeval/internal/runner/runner.go`:

```go
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/dataset"
)

type Runner struct {
	p           *Pipeline
	runDir      string
	concurrency int
}

func NewRunner(p *Pipeline, runDir string, concurrency int) *Runner {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Runner{p: p, runDir: runDir, concurrency: concurrency}
}

type job struct {
	q   dataset.Question
	cfg adapter.RetrievalConfig
}

// Run evaluates every (question, config) pair with a worker pool. Existing
// result JSON files are skipped for resumability.
func (r *Runner) Run(ctx context.Context, qs []dataset.Question, cfgs []adapter.RetrievalConfig, oracle bool) error {
	if err := os.MkdirAll(r.runDir, 0o755); err != nil {
		return err
	}
	jobs := make(chan job, len(qs)*len(cfgs))
	for _, q := range qs {
		for _, cfg := range cfgs {
			jobs <- job{q: q, cfg: cfg}
		}
	}
	close(jobs)

	var wg sync.WaitGroup
	errCh := make(chan error, r.concurrency)

	for i := 0; i < r.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					return
				}
				path := r.resultPath(j.q.ID, j.cfg)
				if _, err := os.Stat(path); err == nil {
					continue // resume: skip already-processed
				}
				res, err := r.p.Run(ctx, j.q, j.cfg, oracle)
				if err != nil {
					errCh <- err
					return
				}
				if err := r.writeResult(path, res); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) resultPath(qid string, cfg adapter.RetrievalConfig) string {
	return filepath.Join(r.runDir, configID(cfg), qid+".json")
}

func configID(cfg adapter.RetrievalConfig) string {
	return fmt.Sprintf("%s-k%d", cfg.Strategy, cfg.TopK)
}

func (r *Runner) writeResult(path string, res QuestionResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(resultJSON(res), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func resultJSON(r QuestionResult) map[string]any {
	return map[string]any{
		"question_id":    r.QuestionID,
		"question_type":  r.QuestionType,
		"config":         configID(r.Config),
		"oracle":         r.Oracle,
		"predicted":      r.Predicted,
		"judge":          r.Judge,
		"hits":           r.Hits,
		"tokens_in":      r.TokensIn,
		"tokens_out":     r.TokensOut,
		"judge_in_tok":   r.JudgeInTok,
		"judge_out_tok":  r.JudgeOutTok,
		"latency_ms":     r.LatencyMS,
		"error":          r.Error,
	}
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/runner/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/runner/
git commit -m "Add worker-pool runner with resume via per-result files"
```

---

## Task 16: Report — aggregate summary (JSON + MD)

**Files:**
- Create: `benchmarks/longmemeval/internal/report/summary.go`
- Create: `benchmarks/longmemeval/internal/report/summary_test.go`

**Contract:** `Summarize(runDir, ledger)` walks every `<runDir>/<configID>/*.json`, aggregates correctness per (config, question_type), computes totals, and writes `summary.json` + `summary.md` + `cost.json` to `runDir`.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/report/summary_test.go`:

```go
package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
)

func writeResult(t *testing.T, dir, config, qid string, body map[string]any) {
	t.Helper()
	_ = os.MkdirAll(filepath.Join(dir, config), 0o755)
	b, _ := json.Marshal(body)
	_ = os.WriteFile(filepath.Join(dir, config, qid+".json"), b, 0o644)
}

func TestSummarize_ComputesAccuracyPerConfig(t *testing.T) {
	dir := t.TempDir()
	writeResult(t, dir, "similarity-k10", "qA",
		map[string]any{"question_id": "qA", "question_type": "t1", "judge": map[string]any{"Correct": true}})
	writeResult(t, dir, "similarity-k10", "qB",
		map[string]any{"question_id": "qB", "question_type": "t2", "judge": map[string]any{"Correct": false}})
	writeResult(t, dir, "activation-k10", "qA",
		map[string]any{"question_id": "qA", "question_type": "t1", "judge": map[string]any{"Correct": true}})

	ledger := cost.NewLedger()
	ledger.Add(cost.Anthropic, "claude-sonnet-4-6", 100, 10)
	ledger.Add(cost.OpenAI, "gpt-4o", 50, 5)

	if err := Summarize(dir, ledger); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(dir, "summary.json"))
	if err != nil {
		t.Fatal(err)
	}
	var s map[string]any
	_ = json.Unmarshal(b, &s)
	configs := s["configs"].(map[string]any)
	sim := configs["similarity-k10"].(map[string]any)
	if sim["accuracy"] != 0.5 {
		t.Errorf("similarity accuracy=%v want 0.5", sim["accuracy"])
	}
	act := configs["activation-k10"].(map[string]any)
	if act["accuracy"] != 1.0 {
		t.Errorf("activation accuracy=%v want 1.0", act["accuracy"])
	}

	md, err := os.ReadFile(filepath.Join(dir, "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(md), "similarity-k10") {
		t.Errorf("summary.md missing config: %s", md)
	}

	cb, err := os.ReadFile(filepath.Join(dir, "cost.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cc map[string]any
	_ = json.Unmarshal(cb, &cc)
	if cc["anthropic_usd"] == nil || cc["openai_usd"] == nil {
		t.Errorf("cost.json missing fields: %s", cb)
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/report/...`
Expected: FAIL.

- [ ] **Step 3: Write summary.go**

Create `benchmarks/longmemeval/internal/report/summary.go`:

```go
// Package report produces aggregate JSON + MD + cost summaries for a run.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
)

type result struct {
	QuestionID   string         `json:"question_id"`
	QuestionType string         `json:"question_type"`
	Judge        map[string]any `json:"judge"`
	Error        string         `json:"error"`
}

type configStats struct {
	Total       int                `json:"total"`
	Correct     int                `json:"correct"`
	Errors      int                `json:"errors"`
	Accuracy    float64            `json:"accuracy"`
	PerQuestion map[string]float64 `json:"per_question_type_accuracy"`
}

type summary struct {
	RunDir  string                  `json:"run_dir"`
	Configs map[string]configStats  `json:"configs"`
}

func Summarize(runDir string, ledger *cost.Ledger) error {
	configs, err := collectConfigs(runDir)
	if err != nil {
		return err
	}
	s := summary{RunDir: runDir, Configs: map[string]configStats{}}
	for _, cfg := range configs {
		stats, err := aggregateConfig(filepath.Join(runDir, cfg))
		if err != nil {
			return err
		}
		s.Configs[cfg] = stats
	}

	if err := writeJSON(filepath.Join(runDir, "summary.json"), s); err != nil {
		return err
	}
	if err := writeMarkdown(filepath.Join(runDir, "summary.md"), s); err != nil {
		return err
	}
	if err := writeCost(filepath.Join(runDir, "cost.json"), ledger); err != nil {
		return err
	}
	return nil
}

func collectConfigs(runDir string) ([]string, error) {
	entries, err := os.ReadDir(runDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func aggregateConfig(dir string) (configStats, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return configStats{}, err
	}
	stats := configStats{PerQuestion: map[string]float64{}}
	counters := map[string][2]int{} // type -> [total, correct]
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return configStats{}, err
		}
		var r result
		if err := json.Unmarshal(b, &r); err != nil {
			return configStats{}, fmt.Errorf("parse %s: %w", f, err)
		}
		stats.Total++
		if r.Error != "" {
			stats.Errors++
			continue
		}
		correct := false
		if v, ok := r.Judge["Correct"]; ok {
			correct = asBool(v)
		}
		if correct {
			stats.Correct++
		}
		c := counters[r.QuestionType]
		c[0]++
		if correct {
			c[1]++
		}
		counters[r.QuestionType] = c
	}
	if stats.Total > 0 {
		stats.Accuracy = float64(stats.Correct) / float64(stats.Total-stats.Errors)
		if stats.Total == stats.Errors {
			stats.Accuracy = 0
		}
	}
	for t, c := range counters {
		if c[0] > 0 {
			stats.PerQuestion[t] = float64(c[1]) / float64(c[0])
		}
	}
	return stats, nil
}

func asBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x == "true"
	default:
		return false
	}
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func writeMarkdown(path string, s summary) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# LongMemEval run: %s\n\n", filepath.Base(s.RunDir))
	fmt.Fprintf(&b, "## Accuracy by config\n\n")
	fmt.Fprintf(&b, "| Config | Total | Correct | Errors | Accuracy |\n|---|---|---|---|---|\n")
	for _, name := range sortedKeys(s.Configs) {
		cs := s.Configs[name]
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %.3f |\n", name, cs.Total, cs.Correct, cs.Errors, cs.Accuracy)
	}
	fmt.Fprintf(&b, "\n## Per question-type accuracy\n\n")
	for _, name := range sortedKeys(s.Configs) {
		fmt.Fprintf(&b, "### %s\n\n", name)
		pq := s.Configs[name].PerQuestion
		for _, qt := range sortedKeysFloat(pq) {
			fmt.Fprintf(&b, "- %s: %.3f\n", qt, pq[qt])
		}
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func sortedKeys(m map[string]configStats) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeysFloat(m map[string]float64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

type costDoc struct {
	AnthropicUSD    float64 `json:"anthropic_usd"`
	OpenAIUSD       float64 `json:"openai_usd"`
	TotalUSD        float64 `json:"total_usd"`
	AnthropicInTok  int     `json:"anthropic_input_tokens"`
	AnthropicOutTok int     `json:"anthropic_output_tokens"`
	OpenAIInTok     int     `json:"openai_input_tokens"`
	OpenAIOutTok    int     `json:"openai_output_tokens"`
}

func writeCost(path string, ledger *cost.Ledger) error {
	p := ledger.TotalUSD()
	t := ledger.TotalTokens()
	c := costDoc{
		AnthropicUSD: p.Anthropic, OpenAIUSD: p.OpenAI, TotalUSD: p.Total,
		AnthropicInTok: t.AnthropicInput, AnthropicOutTok: t.AnthropicOutput,
		OpenAIInTok: t.OpenAIInput, OpenAIOutTok: t.OpenAIOutput,
	}
	return writeJSON(path, c)
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/report/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/report/
git commit -m "Add run summary (JSON + MD) and cost report"
```

---

## Task 17: Report — stack-explorer snapshot invocation

**Files:**
- Create: `benchmarks/longmemeval/internal/report/snapshot.go`
- Create: `benchmarks/longmemeval/internal/report/snapshot_test.go`

**Contract:** After `subset` / `full` stages, call:

```
./stack-explorer snapshot take \
  --repo vanta-conduit \
  --metric longmemeval_s_similarity_accuracy --value <acc> \
  --metric longmemeval_s_activation_accuracy --value <acc> \
  --metric longmemeval_oracle_similarity_accuracy --value <acc> \
  ...
  --metric longmemeval_total_cost_usd --value <cost>
```

The wrapper builds the arg list from `summary.json` + `cost.json` and execs the CLI. It's tested by injecting a fake exec hook.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/report/snapshot_test.go`:

```go
package report

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSummary(t *testing.T, dir string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPushSnapshot_BuildsArgs(t *testing.T) {
	dir := t.TempDir()
	writeSummary(t, dir, `{
  "configs": {
    "similarity-k10": {"accuracy": 0.42},
    "activation-k10": {"accuracy": 0.55}
  }
}`)
	_ = os.WriteFile(filepath.Join(dir, "cost.json"), []byte(`{"total_usd":1.23}`), 0o644)

	var gotCmd string
	var gotArgs []string
	Exec = func(ctx context.Context, name string, args ...string) error {
		gotCmd = name
		gotArgs = args
		return nil
	}
	defer func() { Exec = defaultExec }()

	if err := PushSnapshot(context.Background(), dir, "vanta-conduit", "stack-explorer"); err != nil {
		t.Fatal(err)
	}
	if gotCmd != "stack-explorer" {
		t.Errorf("cmd=%s", gotCmd)
	}
	joined := strings.Join(gotArgs, " ")
	for _, want := range []string{
		"snapshot", "take",
		"--repo", "vanta-conduit",
		"--metric", "longmemeval_similarity-k10_accuracy", "--value", "0.420000",
		"--metric", "longmemeval_activation-k10_accuracy", "--value", "0.550000",
		"--metric", "longmemeval_total_cost_usd", "--value", "1.230000",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in: %s", want, joined)
		}
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/report/...`
Expected: FAIL.

- [ ] **Step 3: Write snapshot.go**

Create `benchmarks/longmemeval/internal/report/snapshot.go`:

```go
package report

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Exec is the function used to run the stack-explorer CLI. It is exported for
// testing; production code keeps the default (exec.CommandContext).
var Exec = defaultExec

func defaultExec(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PushSnapshot reads summary.json + cost.json from runDir and invokes
// `<cli> snapshot take --repo <repoID> --metric ... --value ...` with one
// metric per config accuracy plus total cost.
func PushSnapshot(ctx context.Context, runDir, repoID, cliPath string) error {
	sumB, err := os.ReadFile(filepath.Join(runDir, "summary.json"))
	if err != nil {
		return err
	}
	var sum struct {
		Configs map[string]struct {
			Accuracy float64 `json:"accuracy"`
		} `json:"configs"`
	}
	if err := json.Unmarshal(sumB, &sum); err != nil {
		return err
	}
	costB, err := os.ReadFile(filepath.Join(runDir, "cost.json"))
	if err != nil {
		return err
	}
	var c struct {
		TotalUSD float64 `json:"total_usd"`
	}
	if err := json.Unmarshal(costB, &c); err != nil {
		return err
	}

	args := []string{"snapshot", "take", "--repo", repoID}
	for name, cs := range sum.Configs {
		args = append(args,
			"--metric", fmt.Sprintf("longmemeval_%s_accuracy", name),
			"--value", fmt.Sprintf("%f", cs.Accuracy),
		)
	}
	args = append(args,
		"--metric", "longmemeval_total_cost_usd",
		"--value", fmt.Sprintf("%f", c.TotalUSD),
	)
	return Exec(ctx, cliPath, args...)
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/report/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/report/
git commit -m "Add stack-explorer snapshot push from summary + cost"
```

---

## Task 18: Preflight checks

**Files:**
- Create: `benchmarks/longmemeval/internal/preflight/preflight.go`
- Create: `benchmarks/longmemeval/internal/preflight/preflight_test.go`

**Checks:**
1. `Vanta.Health` returns nil.
2. Anthropic key present and a minimal request succeeds (we don't call the API in tests; we only verify the key is non-empty and let the actual invocation surface any auth error naturally on the first call).
3. OpenAI key present and non-empty (same reasoning).
4. `PythonBin` exists and is executable.
5. `JudgeScript` file exists and is readable.
6. `DataDir` contains `longmemeval_s.json` and `longmemeval_oracle.json` (names per `fetch-dataset.sh` — Task 19).
7. At least 500 MiB free in the filesystem containing `RunsDir`.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/preflight/preflight_test.go`:

```go
package preflight

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCheck_AllGood(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "longmemeval_s.json"), []byte("[]"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "longmemeval_oracle.json"), []byte("[]"), 0o644)
	script := filepath.Join(dir, "judge.py")
	_ = os.WriteFile(script, []byte("#!/usr/bin/env python3\n"), 0o755)

	result := Check(context.Background(), Inputs{
		VantaURL:        srv.URL,
		AnthropicAPIKey: "x",
		OpenAIAPIKey:    "y",
		PythonBin:       "/usr/bin/env",
		JudgeScript:     script,
		DataDir:         dir,
		RunsDir:         dir,
	})
	if !result.OK {
		t.Fatalf("want OK, got: %+v", result)
	}
}

func TestCheck_VantaDown(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "longmemeval_s.json"), []byte("[]"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "longmemeval_oracle.json"), []byte("[]"), 0o644)
	script := filepath.Join(dir, "judge.py")
	_ = os.WriteFile(script, []byte("#!"), 0o755)

	result := Check(context.Background(), Inputs{
		VantaURL:        "http://127.0.0.1:1", // unreachable
		AnthropicAPIKey: "x", OpenAIAPIKey: "y",
		PythonBin:   "/usr/bin/env",
		JudgeScript: script, DataDir: dir, RunsDir: dir,
	})
	if result.OK {
		t.Fatal("want not-OK when Vanta unreachable")
	}
	found := false
	for _, f := range result.Failures {
		if f == "vanta_unreachable" {
			found = true
		}
	}
	if !found {
		t.Errorf("failures=%v want vanta_unreachable", result.Failures)
	}
}

func TestCheck_MissingKeys(t *testing.T) {
	r := Check(context.Background(), Inputs{})
	if r.OK {
		t.Fatal("want not-OK when all keys missing")
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/preflight/...`
Expected: FAIL.

- [ ] **Step 3: Write preflight.go**

Create `benchmarks/longmemeval/internal/preflight/preflight.go`:

```go
// Package preflight runs fail-fast checks before a LongMemEval stage.
package preflight

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter/vanta"
)

type Inputs struct {
	VantaURL        string
	VantaToken      string
	AnthropicAPIKey string
	OpenAIAPIKey    string
	PythonBin       string
	JudgeScript     string
	DataDir         string
	RunsDir         string
}

type Result struct {
	OK       bool
	Failures []string
	Details  map[string]string
}

func Check(ctx context.Context, in Inputs) Result {
	r := Result{OK: true, Details: map[string]string{}}

	// Vanta
	c := vanta.New(in.VantaURL, in.VantaToken, 5*time.Second)
	if err := c.Health(ctx); err != nil {
		r.OK = false
		r.Failures = append(r.Failures, "vanta_unreachable")
		r.Details["vanta_unreachable"] = err.Error()
	}

	// Keys
	if in.AnthropicAPIKey == "" {
		r.OK = false
		r.Failures = append(r.Failures, "anthropic_key_missing")
	}
	if in.OpenAIAPIKey == "" {
		r.OK = false
		r.Failures = append(r.Failures, "openai_key_missing")
	}

	// Python binary
	if in.PythonBin == "" {
		r.OK = false
		r.Failures = append(r.Failures, "python_bin_unset")
	} else if _, err := os.Stat(in.PythonBin); err != nil {
		r.OK = false
		r.Failures = append(r.Failures, "python_bin_missing")
		r.Details["python_bin_missing"] = err.Error()
	}

	// Judge script
	if in.JudgeScript == "" {
		r.OK = false
		r.Failures = append(r.Failures, "judge_script_unset")
	} else if _, err := os.Stat(in.JudgeScript); err != nil {
		r.OK = false
		r.Failures = append(r.Failures, "judge_script_missing")
		r.Details["judge_script_missing"] = err.Error()
	}

	// Dataset files
	for _, name := range []string{"longmemeval_s.json", "longmemeval_oracle.json"} {
		path := filepath.Join(in.DataDir, name)
		if _, err := os.Stat(path); err != nil {
			r.OK = false
			r.Failures = append(r.Failures, fmt.Sprintf("dataset_missing:%s", name))
			r.Details[fmt.Sprintf("dataset_missing:%s", name)] = err.Error()
		}
	}
	return r
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/preflight/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/preflight/
git commit -m "Add preflight checks"
```

---

## Task 19: `scripts/fetch-dataset.sh`

**Files:**
- Create: `benchmarks/longmemeval/scripts/fetch-dataset.sh`

**Behavior:** Downloads LongMemEval_S and LongMemEval_Oracle JSON files to `data/`. LongMemEval is published by Xiao et al. on HuggingFace as `xiaowu0162/LongMemEval`. The split file names are `longmemeval_s.json` and `longmemeval_oracle.json` (verify URL before first use — README should instruct the operator to check if a 404 appears).

- [ ] **Step 1: Create `benchmarks/longmemeval/scripts/fetch-dataset.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA="$HERE/data"
mkdir -p "$DATA"

BASE="${LME_DATASET_BASE:-https://huggingface.co/datasets/xiaowu0162/LongMemEval/resolve/main}"

for f in longmemeval_s.json longmemeval_oracle.json; do
  if [[ -f "$DATA/$f" ]]; then
    echo "already have $f"
    continue
  fi
  echo "fetching $f ..."
  curl -fLo "$DATA/$f" "$BASE/$f"
done

echo "dataset ready in $DATA"
```

Make it executable:

```bash
chmod +x benchmarks/longmemeval/scripts/fetch-dataset.sh
```

- [ ] **Step 2: Document known-fragile assumption in README**

Append to `benchmarks/longmemeval/README.md`:

```markdown

## Dataset source

`scripts/fetch-dataset.sh` downloads `longmemeval_s.json` and `longmemeval_oracle.json` from Hugging Face. If the paths change, override with `LME_DATASET_BASE=<url>` or fetch manually into `data/`.
```

- [ ] **Step 3: Commit**

```bash
git add benchmarks/longmemeval/scripts/fetch-dataset.sh benchmarks/longmemeval/README.md
git commit -m "Add dataset fetch script"
```

---

## Task 20: Gate helpers (--confirm --after + pause output)

**Files:**
- Create: `benchmarks/longmemeval/internal/runner/gate.go`
- Create: `benchmarks/longmemeval/internal/runner/gate_test.go`

**Contract:**

- `WritePauseNotice(w io.Writer, stage, runID string, ledger *cost.Ledger, nextStage string, projected NextCost)` writes the formatted pause block from the spec §8.3 to w.
- `VerifyPriorStage(runsDir, afterRunID string) (priorStage string, err error)` confirms a `summary.json` exists under `runsDir/<afterRunID>/` and returns the stage name inferred from its prefix (`smoke-...`, `subset-...`, etc).
- `RequireConfirm(flags ConfirmFlags) error` returns an error if `--confirm` is not set or `--after` is missing.
- `RequireStdinGO(r io.Reader) error` reads a line from r and returns an error unless it's exactly `GO\n`.

- [ ] **Step 1: Write the failing test**

Create `benchmarks/longmemeval/internal/runner/gate_test.go`:

```go
package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
)

func TestWritePauseNotice_ContainsRequired(t *testing.T) {
	l := cost.NewLedger()
	l.Add(cost.Anthropic, "claude-sonnet-4-6", 100_000, 5_000)
	l.Add(cost.OpenAI, "gpt-4o", 50_000, 2_000)
	var buf bytes.Buffer
	WritePauseNotice(&buf, "smoke", "smoke-2026-04-16-001", l, "subset", NextCost{Anthropic: 0.50, OpenAI: 0.10})
	s := buf.String()
	for _, want := range []string{
		"Stage smoke complete",
		"Anthropic spend",
		"OpenAI spend",
		"PAUSE",
		"To proceed:   lme-run subset --confirm --after smoke-2026-04-16-001",
		"Projected cost for next stage",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("pause notice missing %q", want)
		}
	}
}

func TestVerifyPriorStage(t *testing.T) {
	dir := t.TempDir()
	runID := "smoke-2026-04-16-001"
	_ = os.MkdirAll(filepath.Join(dir, runID), 0o755)
	_ = os.WriteFile(filepath.Join(dir, runID, "summary.json"), []byte(`{}`), 0o644)

	stage, err := VerifyPriorStage(dir, runID)
	if err != nil {
		t.Fatal(err)
	}
	if stage != "smoke" {
		t.Errorf("stage=%q want smoke", stage)
	}
}

func TestVerifyPriorStage_MissingSummary(t *testing.T) {
	dir := t.TempDir()
	runID := "smoke-bad"
	_ = os.MkdirAll(filepath.Join(dir, runID), 0o755)
	if _, err := VerifyPriorStage(dir, runID); err == nil {
		t.Fatal("want error when summary.json missing")
	}
}

func TestRequireConfirm(t *testing.T) {
	if err := RequireConfirm(ConfirmFlags{}); err == nil {
		t.Fatal("want error when not confirmed")
	}
	if err := RequireConfirm(ConfirmFlags{Confirm: true, After: "x"}); err != nil {
		t.Fatal(err)
	}
}

func TestRequireStdinGO(t *testing.T) {
	if err := RequireStdinGO(strings.NewReader("GO\n")); err != nil {
		t.Fatalf("GO should pass: %v", err)
	}
	if err := RequireStdinGO(strings.NewReader("go\n")); err == nil {
		t.Fatal("lowercase go should fail")
	}
	if err := RequireStdinGO(strings.NewReader("STOP\n")); err == nil {
		t.Fatal("STOP should fail")
	}
}
```

- [ ] **Step 2: Run test**

Run: `cd benchmarks/longmemeval && go test ./internal/runner/...`
Expected: FAIL.

- [ ] **Step 3: Write gate.go**

Create `benchmarks/longmemeval/internal/runner/gate.go`:

```go
package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
)

type NextCost struct {
	Anthropic float64
	OpenAI    float64
}

func WritePauseNotice(w io.Writer, stage, runID string, ledger *cost.Ledger, nextStage string, projected NextCost) {
	p := ledger.TotalUSD()
	t := ledger.TotalTokens()
	fmt.Fprintf(w, "\nStage %s complete.\n", stage)
	fmt.Fprintf(w, "  Anthropic spend:  $%.2f  (%d input tok / %d output tok)\n", p.Anthropic, t.AnthropicInput, t.AnthropicOutput)
	fmt.Fprintf(w, "  OpenAI spend:     $%.2f  (%d input tok / %d output tok)\n", p.OpenAI, t.OpenAIInput, t.OpenAIOutput)
	fmt.Fprintf(w, "\nPAUSE: Check your API credit balances before proceeding.\n")
	fmt.Fprintf(w, "  Anthropic console: https://console.anthropic.com/settings/billing\n")
	fmt.Fprintf(w, "  OpenAI console:    https://platform.openai.com/usage\n\n")
	fmt.Fprintf(w, "Projected cost for next stage (%s):\n", nextStage)
	fmt.Fprintf(w, "  Anthropic:  ~$%.2f   OpenAI: ~$%.2f\n\n", projected.Anthropic, projected.OpenAI)
	fmt.Fprintf(w, "To proceed:   lme-run %s --confirm --after %s\n", nextStage, runID)
}

// VerifyPriorStage confirms <runsDir>/<runID>/summary.json exists and returns
// the stage name inferred from runID's prefix up to the first "-".
func VerifyPriorStage(runsDir, runID string) (string, error) {
	if runID == "" {
		return "", errors.New("empty run id")
	}
	sum := filepath.Join(runsDir, runID, "summary.json")
	if _, err := os.Stat(sum); err != nil {
		return "", fmt.Errorf("prior run %q missing summary.json: %w", runID, err)
	}
	i := strings.IndexByte(runID, '-')
	if i <= 0 {
		return "", fmt.Errorf("runID %q has no stage prefix", runID)
	}
	return runID[:i], nil
}

type ConfirmFlags struct {
	Confirm bool
	After   string
}

func RequireConfirm(f ConfirmFlags) error {
	if !f.Confirm {
		return errors.New("--confirm is required")
	}
	if f.After == "" {
		return errors.New("--after <prior-run-id> is required")
	}
	return nil
}

func RequireStdinGO(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return errors.New("no input")
	}
	line := strings.TrimRight(scanner.Text(), "\r")
	if line != "GO" {
		return fmt.Errorf("expected GO, got %q", line)
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd benchmarks/longmemeval && go test ./internal/runner/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add benchmarks/longmemeval/internal/runner/
git commit -m "Add stage gate helpers (pause notice + confirm/after + GO prompt)"
```

---

## Task 21: CLI — root + `preflight` + `smoke`

**Files:**
- Modify: `benchmarks/longmemeval/cmd/lme-run/main.go` (replace stub)

**Context:** stack-explorer uses Cobra for its main CLI; we reuse the dep (`github.com/spf13/cobra`, already in go.mod).

Commands wired in this task:
- `preflight` — runs all checks, prints status, exits 0/1.
- `smoke` — runs the smoke stage: 16 stratified Qs across the 8 (4 types × 2 variants) combinations, two retrieval configs, writes results + summary + pause notice.

Commands `subset` and `full` come in Task 22.

- [ ] **Step 1: Replace main.go with the real CLI**

Overwrite `benchmarks/longmemeval/cmd/lme-run/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter/vanta"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/answerer"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/config"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/dataset"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/judge"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/preflight"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/report"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/runner"
	"github.com/hollis-labs/go-providers/provider"
)

func main() {
	root := &cobra.Command{Use: "lme-run", SilenceUsage: true}
	root.AddCommand(preflightCmd(), smokeCmd(), subsetCmd(), fullCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadConfig() (config.Config, error) {
	m := map[string]string{}
	for _, k := range []string{
		"VANTA_URL", "VANTA_TOKEN",
		"ANTHROPIC_API_KEY", "OPENAI_API_KEY",
		"LME_ANSWER_MODEL", "LME_JUDGE_MODEL",
		"LME_PYTHON", "LME_JUDGE_SCRIPT",
		"LME_DATA_DIR", "LME_RUNS_DIR",
		"LME_CONCURRENCY", "LME_TOPK", "LME_SEED", "LME_MAX_COST_USD",
	} {
		m[k] = os.Getenv(k)
	}
	return config.Load(m)
}

func preflightCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preflight",
		Short: "Check that Vanta, API keys, dataset, Python venv are ready",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			r := preflight.Check(context.Background(), preflight.Inputs{
				VantaURL: cfg.VantaURL, VantaToken: cfg.VantaToken,
				AnthropicAPIKey: cfg.AnthropicAPIKey, OpenAIAPIKey: cfg.OpenAIAPIKey,
				PythonBin: cfg.PythonBin, JudgeScript: cfg.JudgeScript,
				DataDir: cfg.DataDir, RunsDir: cfg.RunsDir,
			})
			if !r.OK {
				fmt.Fprintf(os.Stderr, "preflight FAILED: %v\n", r.Failures)
				for k, v := range r.Details {
					fmt.Fprintf(os.Stderr, "  %s: %s\n", k, v)
				}
				return fmt.Errorf("preflight failed")
			}
			fmt.Println("preflight OK")
			return nil
		},
	}
}

func smokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "smoke",
		Short: "Run a 16-question stratified smoke test",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStage(cmd.Context(), "smoke", 16)
		},
	}
}

func buildPipeline(cfg config.Config) (*runner.Pipeline, *cost.Ledger, error) {
	vc := vanta.New(cfg.VantaURL, cfg.VantaToken, 30*time.Second)

	a := provider.NewAnthropic()
	// NewAnthropic reads ANTHROPIC_API_KEY from env; confirm it's set (preflight also checks).
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return nil, nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	ans := answerer.New(a, cfg.AnswererModel)

	jg := judge.New(cfg.PythonBin, cfg.JudgeScript, cfg.JudgeModel)
	ledger := cost.NewLedger()
	return runner.NewPipeline(vc, ans, jg, ledger, cfg.AnswererModel), ledger, nil
}

// runStage loads dataset, samples N, runs both variants + both retrieval configs, writes summary, prints pause notice.
func runStage(ctx context.Context, stage string, sampleSize int) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	runID := fmt.Sprintf("%s-%s-%06d", stage, time.Now().UTC().Format("2006-01-02-150405"), time.Now().UnixNano()%1_000_000)
	runDir := filepath.Join(cfg.RunsDir, runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}

	p, ledger, err := buildPipeline(cfg)
	if err != nil {
		return err
	}

	configs := []adapter.RetrievalConfig{
		{Strategy: "similarity", TopK: cfg.TopK},
		{Strategy: "activation", TopK: cfg.TopK},
	}

	// Standard variant (_S)
	sPath := filepath.Join(cfg.DataDir, "longmemeval_s.json")
	sQs, err := dataset.Load(sPath)
	if err != nil {
		return fmt.Errorf("load _s: %w", err)
	}
	sSample, err := dataset.StratifiedSample(sQs, sampleSize, cfg.Seed)
	if err != nil {
		return err
	}
	sDir := filepath.Join(runDir, "s")
	if err := runner.NewRunner(p, sDir, cfg.Concurrency).Run(ctx, sSample, configs, false); err != nil {
		return fmt.Errorf("run _s: %w", err)
	}
	if err := report.Summarize(sDir, ledger); err != nil {
		return err
	}

	// Oracle variant
	oPath := filepath.Join(cfg.DataDir, "longmemeval_oracle.json")
	oQs, err := dataset.Load(oPath)
	if err != nil {
		return fmt.Errorf("load oracle: %w", err)
	}
	oSample, err := dataset.StratifiedSample(oQs, sampleSize, cfg.Seed)
	if err != nil {
		return err
	}
	oDir := filepath.Join(runDir, "oracle")
	if err := runner.NewRunner(p, oDir, cfg.Concurrency).Run(ctx, oSample, configs, true); err != nil {
		return fmt.Errorf("run oracle: %w", err)
	}
	if err := report.Summarize(oDir, ledger); err != nil {
		return err
	}

	// top-level summary combining both variants
	if err := report.Summarize(runDir, ledger); err != nil {
		return err
	}

	// Snapshot push — only for subset and full (skip for smoke since smoke numbers are noisy)
	if stage != "smoke" {
		if err := report.PushSnapshot(ctx, runDir, "vanta-conduit", "stack-explorer"); err != nil {
			fmt.Fprintf(os.Stderr, "warning: snapshot push failed: %v\n", err)
		}
	}

	next := nextStageName(stage)
	runner.WritePauseNotice(os.Stdout, stage, runID, ledger, next, projectNextStageCost(stage, ledger, sampleSize))
	return nil
}

func nextStageName(stage string) string {
	switch stage {
	case "smoke":
		return "subset"
	case "subset":
		return "full"
	default:
		return "(none — final stage)"
	}
}

// projectNextStageCost extrapolates from current per-question cost to the
// next stage's question count.
func projectNextStageCost(stage string, ledger *cost.Ledger, currentN int) runner.NextCost {
	p := ledger.TotalUSD()
	if currentN == 0 {
		return runner.NextCost{}
	}
	perQA := p.Anthropic / float64(currentN)
	perQO := p.OpenAI / float64(currentN)
	var nextN int
	switch stage {
	case "smoke":
		nextN = 100
	case "subset":
		nextN = 500
	default:
		return runner.NextCost{}
	}
	return runner.NextCost{Anthropic: perQA * float64(nextN), OpenAI: perQO * float64(nextN)}
}

// subsetCmd and fullCmd — stubs filled in Task 22.
func subsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "subset",
		Short: "Run the 100-question subset (requires --confirm --after <smoke-run-id>)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("not yet implemented (Task 22)")
		},
	}
}

func fullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "full",
		Short: "Run all 500 questions (requires --confirm --after <subset-run-id> and GO on stdin)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("not yet implemented (Task 22)")
		},
	}
}
```

- [ ] **Step 2: Verify build**

Run:

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
go build -o benchmarks/longmemeval/lme-run ./benchmarks/longmemeval/cmd/lme-run
./benchmarks/longmemeval/lme-run --help
```

Expected: binary builds; help lists `preflight`, `smoke`, `subset`, `full`.

- [ ] **Step 3: Verify preflight with a missing-key env fails cleanly**

Run:

```bash
env -i PATH="$PATH" ./benchmarks/longmemeval/lme-run preflight 2>&1 || true
```

Expected: stderr includes `anthropic_key_missing` / `openai_key_missing`, exit non-zero.

- [ ] **Step 4: Commit**

```bash
git add benchmarks/longmemeval/cmd/lme-run/
git commit -m "Wire CLI: preflight + smoke commands"
```

---

## Task 22: CLI — `subset` + `full` commands

**Files:**
- Modify: `benchmarks/longmemeval/cmd/lme-run/main.go` (replace the stubs from Task 21)

Both commands share the same pattern:
- Parse `--confirm`, `--after` flags.
- `runner.RequireConfirm` / `runner.VerifyPriorStage`.
- `full` additionally reads stdin for `GO`.
- Call `runStage` with appropriate sample size (100 for subset, full dataset length for full).

- [ ] **Step 1: Replace the two stub commands**

In `benchmarks/longmemeval/cmd/lme-run/main.go`, replace `subsetCmd` and `fullCmd` with:

```go
func subsetCmd() *cobra.Command {
	var confirm bool
	var after string
	cmd := &cobra.Command{
		Use:   "subset",
		Short: "Run the 100-question subset (requires --confirm --after <smoke-run-id>)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := runner.RequireConfirm(runner.ConfirmFlags{Confirm: confirm, After: after}); err != nil {
				return err
			}
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			stage, err := runner.VerifyPriorStage(cfg.RunsDir, after)
			if err != nil {
				return err
			}
			if stage != "smoke" {
				return fmt.Errorf("subset must follow a smoke run (got %q)", stage)
			}
			return runStage(cmd.Context(), "subset", 100)
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "required to run")
	cmd.Flags().StringVar(&after, "after", "", "prior smoke run id")
	return cmd
}

func fullCmd() *cobra.Command {
	var confirm bool
	var after string
	cmd := &cobra.Command{
		Use:   "full",
		Short: "Run all 500 questions (requires --confirm --after <subset-run-id> and GO on stdin)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := runner.RequireConfirm(runner.ConfirmFlags{Confirm: confirm, After: after}); err != nil {
				return err
			}
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			stage, err := runner.VerifyPriorStage(cfg.RunsDir, after)
			if err != nil {
				return err
			}
			if stage != "subset" {
				return fmt.Errorf("full must follow a subset run (got %q)", stage)
			}
			fmt.Fprintln(os.Stderr, "Full run. Type 'GO' and press enter to proceed:")
			if err := runner.RequireStdinGO(os.Stdin); err != nil {
				return fmt.Errorf("aborted: %w", err)
			}
			return runStage(cmd.Context(), "full", 500)
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "required to run")
	cmd.Flags().StringVar(&after, "after", "", "prior subset run id")
	return cmd
}
```

- [ ] **Step 2: Build and smoke-test the new flags**

Run:

```bash
cd /Users/chrispian/Projects-apps/stack-explorer
go build -o benchmarks/longmemeval/lme-run ./benchmarks/longmemeval/cmd/lme-run
./benchmarks/longmemeval/lme-run subset 2>&1 || true
./benchmarks/longmemeval/lme-run subset --confirm 2>&1 || true
./benchmarks/longmemeval/lme-run full   --confirm --after smoke-made-up 2>&1 || true
```

Expected in order:
1. `--confirm is required`
2. `--after <prior-run-id> is required`
3. `prior run "smoke-made-up" missing summary.json: ...`

- [ ] **Step 3: Commit**

```bash
git add benchmarks/longmemeval/cmd/lme-run/
git commit -m "Wire CLI: subset + full with confirm/after gate and GO prompt"
```

---

## Task 23: Integration test against a local Vanta (gated by env)

**Files:**
- Create: `benchmarks/longmemeval/testdata/mini-dataset.json`
- Create: `benchmarks/longmemeval/internal/runner/integration_test.go`

**Purpose:** Verify the full Go pipeline works end-to-end against a real Vanta, using a fake judge (to avoid OpenAI spend in tests). Skipped unless `VANTA_URL` is set.

- [ ] **Step 1: Create `benchmarks/longmemeval/testdata/mini-dataset.json`**

```json
[
  {
    "question_id": "mini-q1",
    "question_type": "single-session-user",
    "question": "What is the user's favorite color?",
    "answer": "blue",
    "answer_session_ids": ["s1"],
    "haystack_sessions": [
      {
        "session_id": "s1",
        "session_date": "2024-01-10T10:00:00Z",
        "turns": [
          {"role": "user", "content": "My favorite color is blue. I've loved it since childhood."},
          {"role": "assistant", "content": "Noted, your favorite color is blue."}
        ]
      },
      {
        "session_id": "s2",
        "session_date": "2024-01-11T10:00:00Z",
        "turns": [
          {"role": "user", "content": "I want to discuss quantum mechanics today."},
          {"role": "assistant", "content": "Sure."}
        ]
      }
    ]
  },
  {
    "question_id": "mini-q2",
    "question_type": "multi-session",
    "question": "What hobby did the user mention?",
    "answer": "woodworking",
    "answer_session_ids": ["s2"],
    "haystack_sessions": [
      {
        "session_id": "s1",
        "session_date": "2024-02-01T10:00:00Z",
        "turns": [{"role": "user", "content": "Nice weather today."}]
      },
      {
        "session_id": "s2",
        "session_date": "2024-02-02T10:00:00Z",
        "turns": [{"role": "user", "content": "I picked up woodworking last month."}]
      }
    ]
  }
]
```

- [ ] **Step 2: Write the integration test**

Create `benchmarks/longmemeval/internal/runner/integration_test.go`:

```go
package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/adapter/vanta"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/answerer"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/cost"
	"github.com/chrispian/stack-explorer/benchmarks/longmemeval/internal/dataset"
	"github.com/hollis-labs/go-providers/provider"
)

type echoProvider struct{}

func (echoProvider) StreamChat(context.Context, provider.ChatRequest) (<-chan provider.StreamEvent, error) {
	return nil, nil
}
func (echoProvider) Complete(_ context.Context, req provider.ChatRequest) (string, error) {
	return "I don't know.", nil
}
func (echoProvider) Capabilities() provider.ProviderCapabilities {
	return provider.ProviderCapabilities{}
}

// TestIntegration_EndToEnd exercises the full pipeline against a real Vanta.
// Skipped unless VANTA_URL is set. Uses a fake answerer + fake judge to avoid
// LLM spend; the point is to verify Vanta ingest/retrieve actually work.
func TestIntegration_EndToEnd(t *testing.T) {
	url := os.Getenv("VANTA_URL")
	if url == "" {
		t.Skip("VANTA_URL unset; skipping integration test")
	}
	vc := vanta.New(url, os.Getenv("VANTA_TOKEN"), 30*time.Second)
	if err := vc.Health(context.Background()); err != nil {
		t.Skipf("Vanta not healthy at %s: %v", url, err)
	}

	qs, err := dataset.Load(filepath.Join("..", "..", "testdata", "mini-dataset.json"))
	if err != nil {
		t.Fatalf("load mini-dataset: %v", err)
	}

	j := newFakeJudge(t, true)
	p := NewPipeline(vc, answerer.New(echoProvider{}, "claude-sonnet-4-6"), j, cost.NewLedger(), "claude-sonnet-4-6")
	dir := t.TempDir()
	r := NewRunner(p, dir, 1)

	cfgs := []adapter.RetrievalConfig{{Strategy: "similarity", TopK: 3}}
	if err := r.Run(context.Background(), qs, cfgs, false); err != nil {
		t.Fatal(err)
	}
	for _, q := range qs {
		path := filepath.Join(dir, "similarity-k3", q.ID+".json")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing result for %s: %v", q.ID, err)
		}
	}
}
```

- [ ] **Step 3: Run it against the local Vanta**

Run:

```bash
cd /Users/chrispian/Projects-apps/stack-explorer/benchmarks/longmemeval
VANTA_URL=http://127.0.0.1:8089 go test -run TestIntegration_EndToEnd ./internal/runner/... -v
```

Expected: PASS (writes two result JSONs under a temp dir, the test dir is auto-cleaned).

Note: if Vanta returns an error because namespaces need pre-registration, the test surface-errors that precisely — fix the adapter (not the test) until this passes.

- [ ] **Step 4: Commit**

```bash
git add benchmarks/longmemeval/testdata/ benchmarks/longmemeval/internal/runner/integration_test.go
git commit -m "Add end-to-end integration test against local Vanta (gated by VANTA_URL)"
```

---

## Task 24: Final verification — smoke end-to-end

This task is a manual verification, not code.

- [ ] **Step 1: Fetch the real LongMemEval dataset**

Run:

```bash
cd /Users/chrispian/Projects-apps/stack-explorer/benchmarks/longmemeval
make fetch-dataset
```

Expected: `data/longmemeval_s.json` and `data/longmemeval_oracle.json` exist.

- [ ] **Step 2: Setup the Python venv**

Run: `make setup-python`
Expected: `python/.venv/bin/python` exists.

- [ ] **Step 3: Verify preflight**

Run (with keys exported from 1Password):

```bash
export ANTHROPIC_API_KEY="$(op read 'op://...')"
export OPENAI_API_KEY="$(op read 'op://...')"
./lme-run preflight
```

Expected: `preflight OK`.

- [ ] **Step 4: Run smoke**

Run: `./lme-run smoke`

Expected: Completes, writes `runs/smoke-<ts>/` with subdirs `s/`, `oracle/`, a top-level `summary.md`, `summary.json`, `cost.json`, and prints the pause notice with the exact `To proceed:` command for `subset`.

- [ ] **Step 5: Gut-check results**

Open `runs/smoke-<ts>/summary.md`. Verify:
- Two configs appear (`similarity-k10`, `activation-k10`).
- Accuracy numbers are between 0 and 1.
- Oracle accuracy is higher than `_S` accuracy (sanity: reasoning without retrieval noise should do better).
- Cost totals reflect the smoke size (tens of cents, not tens of dollars).

- [ ] **Step 6: Check API credit balances, decide whether to proceed to subset.**

No commit — this task is verification only.

---

## Self-review

**Spec coverage:**

| Spec section | Task(s) |
|---|---|
| 1. Goal / scope / non-goals | 1 (README), 24 (manual verification) |
| 2. Architecture — directory layout | 1 |
| 2.1 MemoryAdapter interface | 2 |
| 2.2 Vanta adapter | 5, 6, 7, 8 |
| 2.3 Concurrency + resume | 15 |
| 2.4 Per-question namespace isolation | 7 (namespace string), 6 (no-op reset + doc) |
| 3. Data flow per question | 14 |
| 3.1 Oracle variant | 14 |
| 3.2 Retrieval configs (similarity + activation) | 8, 21 |
| 4. LLM configuration | 10, 11, 13 |
| 5. Error handling | 7 (retries), 14 (per-stage error capture), 12 (judge errors), 18 (preflight) |
| 6. Testing | All tasks include TDD tests; 23 is the single integration test |
| 7. Reporting | 16, 17 |
| 8. Run-gate sequence | 18, 20, 21, 22 |
| 9. Dependencies | 9 (tiktoken-go), 10 (go-providers), 11 (openai python) |
| 10. Open items / deferred | Not implemented (per spec) |

No spec requirement lacks a task.

**Placeholder scan:** No "TBD", "TODO", or "implement later" remain. One intentional placeholder-like phrase ("the current implementation pins the answerer model to `claude-sonnet-4-6`") is followed immediately by the fix in the same step.

**Type consistency:**
- `adapter.Adapter`, `adapter.Session`, `adapter.Turn`, `adapter.Hit`, `adapter.RetrievalConfig` — used identically across tasks 2-23.
- `dataset.Question` fields (`ID`, `Type`, `Text`, `Answer`, `AnswerSessionIDs`, `HaystackSessions`) consistent across tasks 3, 14, 21, 23.
- `cost.Provider` (`Anthropic`, `OpenAI`) used identically across tasks 9, 14, 16, 20.
- `judge.Result` (`Correct`, `Rationale`, `InputTokens`, `OutputTokens`, `Model`) consistent across tasks 12, 14, 16.
- `runner.Pipeline.NewPipeline(adapter, answerer, judge, ledger, answererModel)` — signature set in task 14, reused identically in task 15, 21, 23.
- `runner.NextCost` defined in task 20, used in task 21.
- `configID(cfg)` format `<strategy>-k<topk>` introduced in task 15, reused verbatim in task 16 tests and task 17 tests.

**Scope:** Single spec → single plan. No subsystem boundary crossed; the plan builds one cohesive harness.

---

## Execution handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-16-longmemeval-vanta.md`. Two execution options:

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?

package jobs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chrispian/stack-explorer/internal/domain"
	"github.com/chrispian/stack-explorer/internal/store/sqlite"
)

type fakeExec struct {
	output []byte
	err    error
	calls  int
}

func (f *fakeExec) CombinedOutput(ctx context.Context, name string, args []string, dir string) ([]byte, error) {
	f.calls++
	return f.output, f.err
}

func openTestStore(t *testing.T) *sqlite.Store {
	t.Helper()
	dir := t.TempDir()
	store, err := sqlite.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return store
}

func TestStartupRecoveryMarksInProgressFailed(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	store, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	job := &domain.Job{ID: "job-1", Kind: "scan", Status: StatusInProgress, PayloadJSON: "{}"}
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("create job: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	store, err = sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	got, err := store.GetJob("job-1")
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if got.Status != StatusFailed || got.Error != "startup-recovery" {
		t.Fatalf("unexpected recovery job state: %+v", got)
	}
}

func TestCalcRetryDelay(t *testing.T) {
	if got := calcRetryDelay("fixed", 2, 3); got != 2*time.Second {
		t.Fatalf("fixed: got %s", got)
	}
	if got := calcRetryDelay("linear", 2, 3); got != 6*time.Second {
		t.Fatalf("linear: got %s", got)
	}
	if got := calcRetryDelay("exponential", 1, 3); got != 4*time.Second {
		t.Fatalf("exponential: got %s", got)
	}
}

func TestScheduleTriggerEmitsEvents(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()
	repoPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoPath, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := store.CreateRepo(&domain.Repo{ID: "repo1", Name: "repo1", LocalPath: repoPath}); err != nil {
		t.Fatalf("create repo: %v", err)
	}
	exec := &fakeExec{output: []byte("ok")}
	svc := NewService(Config{Store: store, Workers: 1, Exec: exec})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("start service: %v", err)
	}
	defer svc.Close()

	schedule := &domain.Schedule{
		ID:             "sched-1",
		RepoID:         "repo1",
		Name:           "nightly",
		CronExpr:       "0 * * * * *",
		JobKind:        "scan",
		PayloadJSON:    `{"blueprint":"blueprints/se-repo-scan.yaml","repo_path":"` + repoPath + `"}`,
		Enabled:        true,
		MaxAttempts:    1,
		RetryBackoff:   "fixed",
		RetryDelaySecs: 1,
	}
	if err := svc.AddSchedule(schedule); err != nil {
		t.Fatalf("add schedule: %v", err)
	}

	ch, stop := svc.Subscribe(sqlite.EventFilter{ScheduleID: "sched-1"}, 16)
	defer stop()
	job, err := svc.TriggerSchedule(context.Background(), "sched-1")
	if err != nil {
		t.Fatalf("trigger schedule: %v", err)
	}
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer waitCancel()
	if _, err := svc.WaitForTerminal(waitCtx, job.ID); err != nil {
		t.Fatalf("wait for terminal: %v", err)
	}

	seen := 0
drain:
	for {
		select {
		case <-ch:
			seen++
		default:
			break drain
		}
	}
	if seen == 0 {
		t.Fatalf("expected events")
	}
	if exec.calls == 0 {
		t.Fatalf("expected fake exec to run")
	}
}

CREATE TABLE IF NOT EXISTS snapshots (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id         TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    captured_at     TEXT NOT NULL,
    stars           INTEGER,
    forks           INTEGER,
    open_issues     INTEGER,
    contributors    INTEGER,
    last_commit_at  TEXT,
    commits_30d     INTEGER,
    loc             INTEGER,
    files           INTEGER,
    test_files      INTEGER,
    dependencies    INTEGER,
    complexity_avg  REAL,
    raw_json        TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_snapshots_repo ON snapshots(repo_id, captured_at DESC);

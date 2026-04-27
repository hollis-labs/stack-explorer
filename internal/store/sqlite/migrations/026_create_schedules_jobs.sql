CREATE TABLE IF NOT EXISTS schedules (
    id               TEXT PRIMARY KEY,
    repo_id          TEXT REFERENCES repos(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    cron_expr        TEXT NOT NULL,
    job_kind         TEXT NOT NULL,
    payload_json     TEXT NOT NULL DEFAULT '{}',
    enabled          INTEGER NOT NULL DEFAULT 1,
    max_attempts     INTEGER NOT NULL DEFAULT 1,
    retry_backoff    TEXT NOT NULL DEFAULT 'fixed',
    retry_delay_secs INTEGER NOT NULL DEFAULT 1,
    last_run_at      TEXT,
    next_run_at      TEXT,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_schedules_repo ON schedules(repo_id);
CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_schedules_job_kind ON schedules(job_kind);

CREATE TABLE IF NOT EXISTS jobs (
    id               TEXT PRIMARY KEY,
    schedule_id      TEXT REFERENCES schedules(id) ON DELETE SET NULL,
    repo_id          TEXT REFERENCES repos(id) ON DELETE CASCADE,
    kind             TEXT NOT NULL,
    status           TEXT NOT NULL,
    attempt_count    INTEGER NOT NULL DEFAULT 0,
    max_attempts     INTEGER NOT NULL DEFAULT 1,
    retry_backoff    TEXT NOT NULL DEFAULT 'fixed',
    retry_delay_secs INTEGER NOT NULL DEFAULT 1,
    payload_json     TEXT NOT NULL DEFAULT '{}',
    output_json      TEXT NOT NULL DEFAULT '{}',
    error            TEXT NOT NULL DEFAULT '',
    started_at       TEXT,
    finished_at      TEXT,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_schedule ON jobs(schedule_id);
CREATE INDEX IF NOT EXISTS idx_jobs_repo ON jobs(repo_id);
CREATE INDEX IF NOT EXISTS idx_jobs_kind ON jobs(kind);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_created ON jobs(created_at DESC);

CREATE TABLE IF NOT EXISTS job_events (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id       TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    schedule_id  TEXT REFERENCES schedules(id) ON DELETE SET NULL,
    repo_id      TEXT REFERENCES repos(id) ON DELETE CASCADE,
    job_kind     TEXT NOT NULL,
    event_type   TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT '',
    message      TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL DEFAULT '{}',
    created_at   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_job_events_job ON job_events(job_id, id);
CREATE INDEX IF NOT EXISTS idx_job_events_schedule ON job_events(schedule_id, id);
CREATE INDEX IF NOT EXISTS idx_job_events_repo ON job_events(repo_id, id);
CREATE INDEX IF NOT EXISTS idx_job_events_kind ON job_events(job_kind, id);

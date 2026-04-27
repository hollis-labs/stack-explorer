CREATE TABLE IF NOT EXISTS audits (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id          TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    scope            TEXT NOT NULL,
    scope_paths      TEXT NOT NULL DEFAULT '[]',
    audit_type       TEXT NOT NULL,
    auditor          TEXT NOT NULL DEFAULT '',
    audited_at_ref   TEXT,
    summary_markdown TEXT NOT NULL DEFAULT '',
    verdict          TEXT,
    status           TEXT NOT NULL DEFAULT 'in_progress',
    supersedes_id    INTEGER REFERENCES audits(id),
    actor_kind       TEXT NOT NULL,
    actor_id         TEXT NOT NULL,
    session_id       TEXT,
    tool_name        TEXT,
    model_name       TEXT,
    started_at       TEXT NOT NULL,
    finished_at      TEXT,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_audits_repo ON audits(repo_id, status);
CREATE INDEX IF NOT EXISTS idx_audits_audited_at ON audits(repo_id, audited_at_ref);

ALTER TABLE findings ADD COLUMN audit_id INTEGER REFERENCES audits(id) ON DELETE SET NULL;
ALTER TABLE findings ADD COLUMN body_markdown TEXT NOT NULL DEFAULT '';
ALTER TABLE findings ADD COLUMN audited_at_ref TEXT;
ALTER TABLE findings ADD COLUMN is_out_of_scope INTEGER NOT NULL DEFAULT 0;
ALTER TABLE findings ADD COLUMN symbol_id INTEGER;
ALTER TABLE findings ADD COLUMN actor_kind TEXT NOT NULL DEFAULT 'unknown';
ALTER TABLE findings ADD COLUMN actor_id TEXT NOT NULL DEFAULT 'unknown';
ALTER TABLE findings ADD COLUMN session_id TEXT;
ALTER TABLE findings ADD COLUMN tool_name TEXT;
ALTER TABLE findings ADD COLUMN model_name TEXT;

CREATE INDEX IF NOT EXISTS idx_findings_audit ON findings(audit_id);

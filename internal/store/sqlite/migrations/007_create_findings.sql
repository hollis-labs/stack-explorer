CREATE TABLE IF NOT EXISTS findings (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id     TEXT REFERENCES repos(id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    category    TEXT NOT NULL DEFAULT '',
    severity    TEXT NOT NULL DEFAULT 'info',
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'open',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_findings_repo ON findings(repo_id);
CREATE INDEX IF NOT EXISTS idx_findings_category ON findings(category);
CREATE INDEX IF NOT EXISTS idx_findings_status ON findings(status);

CREATE TABLE IF NOT EXISTS scans (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id       TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    blueprint     TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    result_json   TEXT NOT NULL DEFAULT '{}',
    error_message TEXT NOT NULL DEFAULT '',
    started_at    TEXT,
    finished_at   TEXT,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_scans_repo ON scans(repo_id);
CREATE INDEX IF NOT EXISTS idx_scans_status ON scans(status);
CREATE INDEX IF NOT EXISTS idx_scans_blueprint ON scans(blueprint);
CREATE INDEX IF NOT EXISTS idx_scans_created ON scans(created_at DESC);

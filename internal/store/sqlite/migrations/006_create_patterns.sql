CREATE TABLE IF NOT EXISTS architecture_patterns (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL DEFAULT 'pattern',
    category    TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS repo_patterns (
    repo_id    TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    pattern_id TEXT NOT NULL REFERENCES architecture_patterns(id) ON DELETE CASCADE,
    quality    TEXT NOT NULL DEFAULT 'present',
    notes      TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (repo_id, pattern_id)
);

CREATE INDEX IF NOT EXISTS idx_repo_patterns_pattern ON repo_patterns(pattern_id);

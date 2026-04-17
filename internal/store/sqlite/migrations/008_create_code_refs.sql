CREATE TABLE IF NOT EXISTS code_references (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id     TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    file_path   TEXT NOT NULL,
    line_start  INTEGER,
    line_end    INTEGER,
    description TEXT NOT NULL DEFAULT '',
    ref_type    TEXT NOT NULL DEFAULT '',
    pattern_id  TEXT REFERENCES architecture_patterns(id) ON DELETE SET NULL,
    finding_id  INTEGER REFERENCES findings(id) ON DELETE SET NULL,
    created_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_code_refs_repo ON code_references(repo_id);
CREATE INDEX IF NOT EXISTS idx_code_refs_pattern ON code_references(pattern_id);
CREATE INDEX IF NOT EXISTS idx_code_refs_finding ON code_references(finding_id);

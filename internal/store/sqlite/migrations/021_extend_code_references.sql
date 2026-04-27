CREATE TABLE code_references_new (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id             TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    file_path           TEXT NOT NULL,
    line_start          INTEGER,
    line_end            INTEGER,
    description         TEXT NOT NULL DEFAULT '',
    ref_type            TEXT NOT NULL DEFAULT '',
    pattern_id          TEXT REFERENCES architecture_patterns(id) ON DELETE SET NULL,
    finding_id          INTEGER REFERENCES findings(id) ON DELETE SET NULL,
    symbol_id           INTEGER REFERENCES symbols(id) ON DELETE SET NULL,
    anchor_content_hash TEXT NOT NULL DEFAULT '',
    stale_since_commit  TEXT,
    created_at          TEXT NOT NULL
);

INSERT INTO code_references_new (
    id, repo_id, file_path, line_start, line_end, description, ref_type, pattern_id, finding_id, created_at
)
SELECT
    id, repo_id, file_path, line_start, line_end, description, ref_type, pattern_id, finding_id, created_at
FROM code_references;

DROP TABLE code_references;
ALTER TABLE code_references_new RENAME TO code_references;

CREATE INDEX IF NOT EXISTS idx_code_refs_repo ON code_references(repo_id);
CREATE INDEX IF NOT EXISTS idx_code_refs_pattern ON code_references(pattern_id);
CREATE INDEX IF NOT EXISTS idx_code_refs_finding ON code_references(finding_id);
CREATE INDEX IF NOT EXISTS idx_code_refs_symbol ON code_references(symbol_id);
CREATE INDEX IF NOT EXISTS idx_code_refs_anchor_content_hash ON code_references(anchor_content_hash);

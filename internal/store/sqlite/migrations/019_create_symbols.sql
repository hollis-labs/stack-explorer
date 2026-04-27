CREATE TABLE IF NOT EXISTS symbols (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id          TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    kind             TEXT NOT NULL,
    name             TEXT NOT NULL,
    qualified_name   TEXT NOT NULL,
    file_path        TEXT NOT NULL,
    line_start       INTEGER,
    line_end         INTEGER,
    content_hash     TEXT NOT NULL,
    signature_hash   TEXT NOT NULL,
    parent_symbol_id INTEGER REFERENCES symbols(id) ON DELETE CASCADE,
    language         TEXT NOT NULL,
    visibility       TEXT,
    docstring        TEXT NOT NULL DEFAULT '',
    stale_since_commit TEXT,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_symbols_repo_qualified_name ON symbols(repo_id, qualified_name);
CREATE INDEX IF NOT EXISTS idx_symbols_repo_content_hash ON symbols(repo_id, content_hash);
CREATE INDEX IF NOT EXISTS idx_symbols_repo_kind ON symbols(repo_id, kind);
CREATE INDEX IF NOT EXISTS idx_symbols_repo_language ON symbols(repo_id, language);
CREATE INDEX IF NOT EXISTS idx_symbols_parent ON symbols(parent_symbol_id);

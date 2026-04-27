CREATE TABLE IF NOT EXISTS relationships (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id       TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    src_symbol_id INTEGER NOT NULL REFERENCES symbols(id) ON DELETE CASCADE,
    dst_symbol_id INTEGER NOT NULL REFERENCES symbols(id) ON DELETE CASCADE,
    kind          TEXT NOT NULL,
    weight        REAL NOT NULL DEFAULT 1.0,
    source        TEXT NOT NULL,
    discovered_at TEXT NOT NULL,
    UNIQUE(src_symbol_id, dst_symbol_id, kind, source)
);

CREATE INDEX IF NOT EXISTS idx_relationships_repo_kind ON relationships(repo_id, kind);
CREATE INDEX IF NOT EXISTS idx_relationships_src_kind ON relationships(src_symbol_id, kind);
CREATE INDEX IF NOT EXISTS idx_relationships_dst_kind ON relationships(dst_symbol_id, kind);
CREATE INDEX IF NOT EXISTS idx_relationships_source_kind ON relationships(source, kind);

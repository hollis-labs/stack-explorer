CREATE TABLE IF NOT EXISTS embeddings (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    target_kind  TEXT NOT NULL,
    target_id    INTEGER NOT NULL,
    model        TEXT NOT NULL,
    dim          INTEGER NOT NULL,
    vector       BLOB NOT NULL,
    content_hash TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    UNIQUE(target_kind, target_id, model)
);

CREATE INDEX IF NOT EXISTS idx_emb_target ON embeddings(target_kind, target_id);
CREATE INDEX IF NOT EXISTS idx_emb_model ON embeddings(model);

ALTER TABLE repos ADD COLUMN embedding_profile TEXT NOT NULL DEFAULT 'none';

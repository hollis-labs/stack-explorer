CREATE TABLE IF NOT EXISTS lenses (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS lens_dimensions (
    lens_id      TEXT NOT NULL REFERENCES lenses(id) ON DELETE CASCADE,
    dimension_id TEXT NOT NULL REFERENCES review_dimensions(id) ON DELETE CASCADE,
    weight       REAL NOT NULL DEFAULT 1.0,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (lens_id, dimension_id)
);

CREATE INDEX IF NOT EXISTS idx_lens_dims_lens ON lens_dimensions(lens_id);

-- Add lens_id to scorecards (nullable for backward compat with existing data)
ALTER TABLE scorecards ADD COLUMN lens_id TEXT REFERENCES lenses(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_scorecards_lens ON scorecards(repo_id, lens_id);

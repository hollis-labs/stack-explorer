CREATE TABLE IF NOT EXISTS scorecards (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id     TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    scored_at   TEXT NOT NULL,
    overall     REAL NOT NULL DEFAULT 0,
    notes       TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_scorecards_repo ON scorecards(repo_id, scored_at DESC);

CREATE TABLE IF NOT EXISTS dimension_scores (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    scorecard_id  INTEGER NOT NULL REFERENCES scorecards(id) ON DELETE CASCADE,
    dimension_id  TEXT NOT NULL REFERENCES review_dimensions(id) ON DELETE CASCADE,
    score         REAL NOT NULL DEFAULT 0,
    evidence      TEXT NOT NULL DEFAULT '',
    notes         TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_dim_scores_card ON dimension_scores(scorecard_id);
CREATE INDEX IF NOT EXISTS idx_dim_scores_dim ON dimension_scores(dimension_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_dim_scores_uniq ON dimension_scores(scorecard_id, dimension_id);

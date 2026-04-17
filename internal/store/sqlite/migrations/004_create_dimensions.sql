CREATE TABLE IF NOT EXISTS review_dimensions (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    category    TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    weight      REAL NOT NULL DEFAULT 1.0,
    sort_order  INTEGER NOT NULL DEFAULT 0
);

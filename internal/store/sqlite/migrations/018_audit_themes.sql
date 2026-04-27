CREATE TABLE IF NOT EXISTS audit_themes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id     TEXT REFERENCES repos(id) ON DELETE SET NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE(repo_id, name)
);

CREATE TABLE IF NOT EXISTS finding_themes (
    finding_id INTEGER NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    theme_id   INTEGER NOT NULL REFERENCES audit_themes(id) ON DELETE CASCADE,
    PRIMARY KEY (finding_id, theme_id)
);

CREATE INDEX IF NOT EXISTS idx_finding_themes_theme ON finding_themes(theme_id);

CREATE TABLE IF NOT EXISTS comparison_sets (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS comparison_set_repos (
    set_id   TEXT NOT NULL REFERENCES comparison_sets(id) ON DELETE CASCADE,
    repo_id  TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    role     TEXT NOT NULL DEFAULT 'subject',
    PRIMARY KEY (set_id, repo_id)
);

CREATE INDEX IF NOT EXISTS idx_comp_set_repos_set ON comparison_set_repos(set_id);

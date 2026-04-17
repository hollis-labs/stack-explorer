CREATE TABLE IF NOT EXISTS repos (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    url         TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    stack       TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT '',
    is_own      INTEGER NOT NULL DEFAULT 0,
    local_path  TEXT NOT NULL DEFAULT '',
    homepage    TEXT NOT NULL DEFAULT '',
    license     TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_repos_category ON repos(category);
CREATE INDEX IF NOT EXISTS idx_repos_is_own ON repos(is_own);
CREATE INDEX IF NOT EXISTS idx_repos_stack ON repos(stack);

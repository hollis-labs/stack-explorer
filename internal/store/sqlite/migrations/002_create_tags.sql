CREATE TABLE IF NOT EXISTS tags (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS repo_tags (
    repo_id TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    tag_id  TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (repo_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_repo_tags_tag ON repo_tags(tag_id);

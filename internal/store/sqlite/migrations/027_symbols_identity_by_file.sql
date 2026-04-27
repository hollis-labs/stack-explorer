DROP INDEX IF EXISTS idx_symbols_repo_qualified_name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_symbols_repo_file_qualified_name
ON symbols(repo_id, file_path, qualified_name);

CREATE INDEX IF NOT EXISTS idx_symbols_repo_qualified_name_lookup
ON symbols(repo_id, qualified_name);

CREATE INDEX IF NOT EXISTS idx_symbols_repo_file_content_hash
ON symbols(repo_id, file_path, content_hash);

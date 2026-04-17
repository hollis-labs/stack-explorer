CREATE TABLE IF NOT EXISTS report_configs (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    lens_id     TEXT REFERENCES lenses(id) ON DELETE SET NULL,
    audience    TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS report_config_filters (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    config_id   TEXT NOT NULL REFERENCES report_configs(id) ON DELETE CASCADE,
    filter_type TEXT NOT NULL,
    filter_value TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_report_config_filters ON report_config_filters(config_id);

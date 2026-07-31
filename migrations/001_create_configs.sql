CREATE TABLE IF NOT EXISTS configs (
    id         TEXT PRIMARY KEY,
    host       TEXT NOT NULL,
    port       INTEGER NOT NULL,
    app_name   TEXT NOT NULL,
    log_level  TEXT NOT NULL DEFAULT 'INFO',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

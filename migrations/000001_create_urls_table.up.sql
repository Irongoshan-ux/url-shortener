CREATE TABLE IF NOT EXISTS urls (
    short_url    TEXT PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

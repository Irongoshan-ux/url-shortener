CREATE TABLE urls (
    short_url    TEXT PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL,
    user_id      TEXT NOT NULL,
    is_deleted   BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX urls_original_url_not_deleted_key ON urls(original_url) WHERE (is_deleted = false);

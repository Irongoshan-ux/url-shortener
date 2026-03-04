ALTER TABLE urls DROP CONSTRAINT IF EXISTS urls_original_url_key;
CREATE UNIQUE INDEX IF NOT EXISTS urls_original_url_not_deleted_key ON urls(original_url) WHERE (is_deleted = false);

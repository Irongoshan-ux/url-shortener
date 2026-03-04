DROP INDEX IF EXISTS urls_original_url_not_deleted_key;
ALTER TABLE urls ADD CONSTRAINT urls_original_url_key UNIQUE (original_url);

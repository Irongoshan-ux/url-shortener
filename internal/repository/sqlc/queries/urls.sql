-- name: CreateUrl :exec
INSERT INTO urls (short_url, original_url, created_at, user_id, is_deleted)
VALUES ($1, $2, $3, $4, $5);

-- name: CreateUrlBatch :batchexec
INSERT INTO urls (short_url, original_url, created_at, user_id, is_deleted)
VALUES ($1, $2, $3, $4, $5);

-- name: GetByShortURL :one
SELECT short_url, original_url, created_at, user_id, is_deleted
FROM urls
WHERE short_url = $1;

-- name: GetByOriginalURL :one
SELECT short_url, original_url, created_at, user_id, is_deleted
FROM urls
WHERE original_url = $1 AND is_deleted = false;

-- name: GetByUserID :many
SELECT short_url, original_url, created_at, user_id, is_deleted
FROM urls
WHERE user_id = $1 AND is_deleted = false;

-- name: MarkDeletedByShortURLs :exec
UPDATE urls
SET is_deleted = true
WHERE user_id = $1 AND short_url = ANY($2::text[]);

-- name: CountURLs :one
SELECT COUNT(*)::int FROM urls WHERE is_deleted = false;

-- name: CountUsers :one
SELECT COUNT(DISTINCT user_id)::int FROM urls WHERE is_deleted = false AND user_id <> '';

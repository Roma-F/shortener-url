-- name: create-migrations-table
CREATE TABLE IF NOT EXISTS migrations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: check-migration-applied
SELECT COUNT(*) FROM migrations WHERE name = $1;

-- name: record-migration
INSERT INTO migrations (name) VALUES ($1);

-- name: save-url
INSERT INTO urls (short_url, original_url, is_deleted)
VALUES ($1, $2, FALSE)
ON CONFLICT (short_url) DO NOTHING;

-- name: save-url-check-conflict
INSERT INTO urls (short_url, original_url, is_deleted)
VALUES ($1, $2, FALSE)
ON CONFLICT (original_url) DO NOTHING;

-- name: save-url-with-user
INSERT INTO urls (short_url, original_url, user_id, is_deleted)
VALUES ($1, $2, $3, FALSE)
ON CONFLICT (short_url) DO NOTHING;

-- name: save-url-with-user-check-conflict
INSERT INTO urls (short_url, original_url, user_id, is_deleted)
VALUES ($1, $2, $3, FALSE)
ON CONFLICT (original_url) DO NOTHING;

-- name: fetch-url
SELECT original_url FROM urls WHERE short_url = $1 AND is_deleted = FALSE;

-- name: fetch-url-with-deleted
SELECT original_url, is_deleted FROM urls WHERE short_url = $1;

-- name: find-by-original-url
SELECT short_url FROM urls WHERE original_url = $1 AND is_deleted = FALSE LIMIT 1;

-- name: check-short-url-exists
SELECT COUNT(*) FROM urls WHERE short_url = $1;

-- name: get-user-urls
SELECT short_url, original_url FROM urls WHERE user_id = $1 AND is_deleted = FALSE ORDER BY id;

-- name: mark-url-as-deleted
UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url = $2 AND is_deleted = FALSE;

-- name: check-url-deleted
SELECT is_deleted FROM urls WHERE short_url = $1;
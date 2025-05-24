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
INSERT INTO urls (short_url, original_url)
VALUES ($1, $2)
ON CONFLICT (short_url) DO NOTHING;

-- name: save-url-check-conflict
INSERT INTO urls (short_url, original_url)
VALUES ($1, $2)
ON CONFLICT (original_url) DO NOTHING;

-- name: save-url-with-user
INSERT INTO urls (short_url, original_url, user_id)
VALUES ($1, $2, $3)
ON CONFLICT (short_url) DO NOTHING;

-- name: save-url-with-user-check-conflict
INSERT INTO urls (short_url, original_url, user_id)
VALUES ($1, $2, $3)
ON CONFLICT (original_url) DO NOTHING;

-- name: fetch-url
SELECT original_url FROM urls WHERE short_url = $1;

-- name: find-by-original-url
SELECT short_url FROM urls WHERE original_url = $1 LIMIT 1;

-- name: check-short-url-exists
SELECT COUNT(*) FROM urls WHERE short_url = $1;

-- name: get-user-urls
SELECT short_url, original_url FROM urls WHERE user_id = $1 ORDER BY id;
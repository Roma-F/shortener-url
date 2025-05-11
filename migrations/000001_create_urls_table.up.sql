CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL UNIQUE,
    original_url TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_original_url ON urls (original_url);
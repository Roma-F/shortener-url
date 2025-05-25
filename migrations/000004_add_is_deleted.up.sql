ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_is_deleted ON urls (is_deleted);
CREATE INDEX IF NOT EXISTS idx_user_id_deleted ON urls (user_id, is_deleted);
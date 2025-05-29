DROP INDEX IF EXISTS idx_user_id_deleted;
DROP INDEX IF EXISTS idx_is_deleted;

ALTER TABLE urls DROP COLUMN IF EXISTS is_deleted;
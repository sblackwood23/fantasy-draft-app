-- Remove index
DROP INDEX IF EXISTS idx_users_event_id;

-- Remove unique constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_event_username_unique;

-- Remove event_id column
ALTER TABLE users DROP COLUMN IF EXISTS event_id;

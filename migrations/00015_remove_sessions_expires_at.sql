-- Remove expires_at from sessions table
-- WHY: Sessions are now permanent until explicit logout
-- HOW: Drop column and its index

-- Drop index first
DROP INDEX IF EXISTS idx_sessions_expires_at;

-- Drop expires_at column
ALTER TABLE sessions DROP COLUMN IF EXISTS expires_at;

-- Update comment
COMMENT ON TABLE sessions IS 'Active user sessions for authentication (permanent until logout)';

-- +goose Up

-- Add missing fields to users table for auth system
ALTER TABLE users ADD COLUMN name TEXT DEFAULT '';

-- Update existing sessions table for auth system compatibility
ALTER TABLE sessions ADD COLUMN is_revoked INTEGER NOT NULL DEFAULT 0;

-- Update sessions table index to include is_revoked field
CREATE INDEX IF NOT EXISTS idx_sessions_is_revoked ON sessions (is_revoked);
CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions (user_id, is_revoked, expires_at);

-- Create refresh_tokens table for JWT refresh token management
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_used INTEGER NOT NULL DEFAULT 0
);

-- Create indexes for refresh_tokens table
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id ON refresh_tokens (session_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_is_used ON refresh_tokens (is_used);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_active ON refresh_tokens (is_used, expires_at);

-- +goose Down

-- Drop refresh_tokens table and its indexes
DROP INDEX IF EXISTS idx_refresh_tokens_active;
DROP INDEX IF EXISTS idx_refresh_tokens_is_used;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP INDEX IF EXISTS idx_refresh_tokens_session_id;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP TABLE IF EXISTS refresh_tokens;

-- Drop new sessions indexes
DROP INDEX IF EXISTS idx_sessions_user_active;
DROP INDEX IF EXISTS idx_sessions_is_revoked;

-- Remove added columns from sessions and users tables
-- Note: SQLite doesn't support dropping columns directly, so we'd need to recreate tables
-- For now, we'll leave the columns as they won't cause issues
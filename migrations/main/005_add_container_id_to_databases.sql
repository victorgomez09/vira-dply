-- +goose Up
-- Add container_id field to databases table for container management
ALTER TABLE databases ADD COLUMN container_id TEXT DEFAULT '';

-- Create index for container_id lookups
CREATE INDEX IF NOT EXISTS idx_databases_container_id ON databases(container_id) WHERE container_id != '';

-- +goose Down
-- Drop the index
DROP INDEX IF EXISTS idx_databases_container_id;

-- Remove the container_id column
ALTER TABLE databases DROP COLUMN container_id;
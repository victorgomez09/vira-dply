-- +goose Up
-- Add base_path column to applications table for specifying subdirectory in repo
ALTER TABLE applications ADD COLUMN base_path TEXT DEFAULT '';

-- +goose Down
ALTER TABLE applications DROP COLUMN base_path;

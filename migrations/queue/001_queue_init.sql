-- +goose Up
-- Queue database initialization
-- Note: Redis/Dragonfly typically don't require schema migrations
-- This file serves as a placeholder for any queue configuration needed

-- Create a configuration marker to track queue database initialization
-- This is stored as a simple key-value pair in Redis/Dragonfly

-- +goose Down
-- Queue database cleanup
-- Note: This would typically clear any persistent queue configuration
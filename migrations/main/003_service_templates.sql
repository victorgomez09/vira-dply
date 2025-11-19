-- +goose Up
-- Service templates table for template marketplace
CREATE TABLE IF NOT EXISTS service_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    version TEXT NOT NULL,
    git_url TEXT, -- JSON string for GitURL object
    build_config TEXT, -- JSON string for BuildConfig object  
    environment TEXT NOT NULL DEFAULT '{}', -- JSON string for environment variables
    ports TEXT NOT NULL DEFAULT '[]', -- JSON string for Port array
    volumes TEXT NOT NULL DEFAULT '[]', -- JSON string for Volume array
    is_official INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL, -- RFC3339 timestamp
    updated_at TEXT NOT NULL  -- RFC3339 timestamp
);

-- Create indexes for service templates
CREATE INDEX IF NOT EXISTS idx_service_templates_name ON service_templates(name);
CREATE INDEX IF NOT EXISTS idx_service_templates_category ON service_templates(category);
CREATE INDEX IF NOT EXISTS idx_service_templates_official ON service_templates(is_official);
CREATE INDEX IF NOT EXISTS idx_service_templates_created_at ON service_templates(created_at);

-- +goose Down
-- Drop indexes first
DROP INDEX IF EXISTS idx_service_templates_created_at;
DROP INDEX IF EXISTS idx_service_templates_official;
DROP INDEX IF EXISTS idx_service_templates_category;
DROP INDEX IF EXISTS idx_service_templates_name;

-- Drop table
DROP TABLE IF EXISTS service_templates;
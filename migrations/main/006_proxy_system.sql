-- +goose Up
-- Create proxy configurations table
CREATE TABLE IF NOT EXISTS proxy_configs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    project_id TEXT NOT NULL,
    service_name TEXT NOT NULL,
    container_id TEXT DEFAULT '',
    hostnames TEXT NOT NULL, -- JSON array of hostnames
    target_url TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 80,
    protocol TEXT NOT NULL DEFAULT 'http',
    path_prefix TEXT DEFAULT '',
    strip_prefix BOOLEAN DEFAULT FALSE,
    middlewares TEXT DEFAULT '[]', -- JSON array of middleware configs
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create traefik global configurations table
CREATE TABLE IF NOT EXISTS traefik_configs (
    id TEXT PRIMARY KEY,
    version TEXT NOT NULL DEFAULT 'v3.0',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_proxy_configs_project_id ON proxy_configs(project_id);
CREATE INDEX IF NOT EXISTS idx_proxy_configs_service_name ON proxy_configs(project_id, service_name);
CREATE INDEX IF NOT EXISTS idx_proxy_configs_container_id ON proxy_configs(container_id) WHERE container_id != '';
CREATE INDEX IF NOT EXISTS idx_proxy_configs_status ON proxy_configs(status);
CREATE INDEX IF NOT EXISTS idx_proxy_configs_hostnames ON proxy_configs(hostnames); -- For hostname lookups

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_proxy_configs_hostnames;
DROP INDEX IF EXISTS idx_proxy_configs_status;
DROP INDEX IF EXISTS idx_proxy_configs_container_id;
DROP INDEX IF EXISTS idx_proxy_configs_service_name;
DROP INDEX IF EXISTS idx_proxy_configs_project_id;

-- Drop tables
DROP TABLE IF EXISTS traefik_configs;
DROP TABLE IF EXISTS proxy_configs;
-- +goose Up
-- Databases table for managing database instances
CREATE TABLE IF NOT EXISTS databases (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL CHECK(type IN (
        'postgresql', 'mysql', 'mariadb', 'redis', 
        'keydb', 'dragonfly', 'mongodb', 'clickhouse'
    )),
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id TEXT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    
    -- Configuration stored as JSON for flexibility
    config TEXT DEFAULT '{}' NOT NULL,
    
    -- Current status of the database
    status TEXT NOT NULL DEFAULT 'created' CHECK(status IN (
        'created', 'provisioning', 'running', 'stopped', 'failed', 'deleting'
    )),
    
    -- Connection information
    connection_string TEXT,
    ports TEXT DEFAULT '{}' NOT NULL, -- JSON map of service name to port
    
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique database names within a project
    UNIQUE(name, project_id)
);

-- Create indexes for databases table
CREATE INDEX IF NOT EXISTS idx_databases_project_id ON databases(project_id);
CREATE INDEX IF NOT EXISTS idx_databases_environment_id ON databases(environment_id);
CREATE INDEX IF NOT EXISTS idx_databases_type ON databases(type);
CREATE INDEX IF NOT EXISTS idx_databases_status ON databases(status);
CREATE INDEX IF NOT EXISTS idx_databases_created_at ON databases(created_at);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_databases_timestamp
    AFTER UPDATE ON databases
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE databases SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down

-- Drop trigger first
DROP TRIGGER IF EXISTS update_databases_timestamp;

-- Drop all indexes
DROP INDEX IF EXISTS idx_databases_created_at;
DROP INDEX IF EXISTS idx_databases_status;
DROP INDEX IF EXISTS idx_databases_type;
DROP INDEX IF EXISTS idx_databases_environment_id;
DROP INDEX IF EXISTS idx_databases_project_id;

-- Drop the table
DROP TABLE IF EXISTS databases;
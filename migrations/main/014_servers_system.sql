-- +goose Up
CREATE TABLE IF NOT EXISTS servers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    hostname TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    port INTEGER DEFAULT 22,
    server_type TEXT NOT NULL,
    status TEXT DEFAULT 'online',
    cpu_cores INTEGER,
    memory_mb INTEGER,
    disk_gb INTEGER,
    os TEXT,
    os_version TEXT,
    metadata TEXT DEFAULT '{}',
    organization_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_servers_organization_id ON servers(organization_id);
CREATE INDEX idx_servers_server_type ON servers(server_type);
CREATE INDEX idx_servers_status ON servers(status);

-- +goose Down
DROP INDEX IF EXISTS idx_servers_status;
DROP INDEX IF EXISTS idx_servers_server_type;
DROP INDEX IF EXISTS idx_servers_organization_id;
DROP TABLE IF EXISTS servers;

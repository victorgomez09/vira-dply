-- +goose Up
CREATE TABLE IF NOT EXISTS activities (
    id TEXT PRIMARY KEY,
    activity_type TEXT NOT NULL,
    description TEXT NOT NULL,
    initiator_id TEXT,
    initiator_name TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    resource_name TEXT,
    metadata TEXT DEFAULT '{}',
    organization_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_activities_organization_id ON activities(organization_id);
CREATE INDEX idx_activities_created_at ON activities(created_at DESC);
CREATE INDEX idx_activities_activity_type ON activities(activity_type);
CREATE INDEX idx_activities_resource ON activities(resource_type, resource_id);

-- +goose Down
DROP INDEX IF EXISTS idx_activities_resource;
DROP INDEX IF EXISTS idx_activities_activity_type;
DROP INDEX IF EXISTS idx_activities_created_at;
DROP INDEX IF EXISTS idx_activities_organization_id;
DROP TABLE IF EXISTS activities;

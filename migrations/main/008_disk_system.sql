-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS disks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    project_id TEXT NOT NULL,
    service_id TEXT,
    size_bytes BIGINT NOT NULL,
    mount_path TEXT NOT NULL,
    filesystem TEXT NOT NULL,
    status TEXT NOT NULL,
    persistent BOOLEAN NOT NULL DEFAULT true,
    backup_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_disks_project_id ON disks(project_id);
CREATE INDEX idx_disks_service_id ON disks(service_id);
CREATE INDEX idx_disks_status ON disks(status);

CREATE TABLE IF NOT EXISTS disk_backups (
    id TEXT PRIMARY KEY,
    disk_id TEXT NOT NULL,
    name TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (disk_id) REFERENCES disks(id) ON DELETE CASCADE
);

CREATE INDEX idx_disk_backups_disk_id ON disk_backups(disk_id);
CREATE INDEX idx_disk_backups_status ON disk_backups(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_disk_backups_status;
DROP INDEX IF EXISTS idx_disk_backups_disk_id;
DROP TABLE IF EXISTS disk_backups;

DROP INDEX IF EXISTS idx_disks_status;
DROP INDEX IF EXISTS idx_disks_service_id;
DROP INDEX IF EXISTS idx_disks_project_id;
DROP TABLE IF EXISTS disks;
-- +goose StatementEnd

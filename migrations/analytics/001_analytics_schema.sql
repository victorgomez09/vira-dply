-- +goose Up
-- Analytics database schema for metrics, events, and logs

-- Metrics table for time-series data
CREATE TABLE IF NOT EXISTS metrics (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    value REAL NOT NULL,
    project_id TEXT NOT NULL DEFAULT '',
    service_id TEXT,
    unit TEXT NOT NULL DEFAULT 'count',
    tags TEXT NOT NULL DEFAULT '{}', -- JSON object
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for time-based queries
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);
CREATE INDEX IF NOT EXISTS idx_metrics_name_timestamp ON metrics(name, timestamp);
CREATE INDEX IF NOT EXISTS idx_metrics_project_id ON metrics(project_id);
CREATE INDEX IF NOT EXISTS idx_metrics_service_id ON metrics(service_id);
CREATE INDEX IF NOT EXISTS idx_metrics_project_service ON metrics(project_id, service_id);

-- Events table for application events
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    source TEXT NOT NULL,
    data TEXT NOT NULL DEFAULT '{}', -- JSON object
    tags TEXT NOT NULL DEFAULT '{}', -- JSON object
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for event queries
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_events_source ON events(source);
CREATE INDEX IF NOT EXISTS idx_events_type_timestamp ON events(type, timestamp);

-- Logs table for application logs
CREATE TABLE IF NOT EXISTS logs (
    id TEXT PRIMARY KEY,
    level TEXT NOT NULL CHECK(level IN ('debug', 'info', 'warn', 'error', 'fatal')),
    message TEXT NOT NULL,
    source TEXT NOT NULL,
    fields TEXT NOT NULL DEFAULT '{}', -- JSON object
    tags TEXT NOT NULL DEFAULT '{}', -- JSON object
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for log queries
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);
CREATE INDEX IF NOT EXISTS idx_logs_level_timestamp ON logs(level, timestamp);

-- +goose Down
DROP INDEX IF EXISTS idx_logs_level_timestamp;
DROP INDEX IF EXISTS idx_logs_source;
DROP INDEX IF EXISTS idx_logs_level;
DROP INDEX IF EXISTS idx_logs_timestamp;
DROP TABLE IF EXISTS logs;

DROP INDEX IF EXISTS idx_events_type_timestamp;
DROP INDEX IF EXISTS idx_events_source;
DROP INDEX IF EXISTS idx_events_type;
DROP INDEX IF EXISTS idx_events_timestamp;
DROP TABLE IF EXISTS events;

DROP INDEX IF EXISTS idx_metrics_project_service;
DROP INDEX IF EXISTS idx_metrics_service_id;
DROP INDEX IF EXISTS idx_metrics_project_id;
DROP INDEX IF EXISTS idx_metrics_name_timestamp;
DROP INDEX IF EXISTS idx_metrics_name;
DROP INDEX IF EXISTS idx_metrics_timestamp;
DROP TABLE IF EXISTS metrics;
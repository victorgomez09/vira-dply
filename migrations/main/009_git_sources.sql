-- +goose Up
-- +goose StatementBegin
CREATE TABLE git_sources (
    id TEXT PRIMARY KEY,
    org_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    name TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expires_at TIMESTAMP,
    custom_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_git_sources_user ON git_sources(user_id);
CREATE INDEX idx_git_sources_org ON git_sources(org_id);
CREATE INDEX idx_git_sources_provider ON git_sources(provider);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_git_sources_provider;
DROP INDEX IF EXISTS idx_git_sources_org;
DROP INDEX IF EXISTS idx_git_sources_user;
DROP TABLE IF EXISTS git_sources;
-- +goose StatementEnd

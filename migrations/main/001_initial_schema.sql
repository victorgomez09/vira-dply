-- +goose Up
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE COLLATE nocase,
    password_hash TEXT NOT NULL,
    username TEXT UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('active', 'inactive', 'suspended', 'pending')),
    email_verified_at DATETIME,
    last_login_at DATETIME,
    timezone TEXT DEFAULT 'UTC',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- User roles and permissions
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT, -- JSON array of permissions
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_roles (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_id TEXT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    granted_by TEXT REFERENCES users (id),
    granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    UNIQUE (user_id, role_id)
);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    last_accessed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    is_active INTEGER NOT NULL DEFAULT 1
);

-- Organizations for multi-tenancy
CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    owner_id TEXT NOT NULL REFERENCES users(id),
    billing_email TEXT,
    plan TEXT NOT NULL DEFAULT 'free' CHECK(plan IN ('free', 'pro', 'enterprise')),
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'suspended', 'deleted')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS organization_members (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('owner', 'admin', 'developer', 'member', 'viewer')),
    invited_by TEXT REFERENCES users(id),
    invited_at DATETIME,
    joined_at DATETIME,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'pending', 'inactive')),
    UNIQUE(organization_id, user_id)
);


-- Create indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at);

-- Create indexes for sessions table
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions (token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions (is_active, expires_at);

-- Create indexes for user roles
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (role_id);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_users_timestamp
    AFTER UPDATE ON users
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by TEXT NOT NULL REFERENCES users(id),
    settings TEXT DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, name)
);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_projects_timestamp
    AFTER UPDATE ON projects
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- Environments table
CREATE TABLE IF NOT EXISTS environments (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_production INTEGER NOT NULL DEFAULT 0,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_environments_project_id ON environments(project_id);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_environments_timestamp
    AFTER UPDATE ON environments
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE environments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- Applications table
CREATE TABLE IF NOT EXISTS applications (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id TEXT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    
    repo_url TEXT,
    repo_branch TEXT DEFAULT 'main',
    repo_path TEXT, -- subdir inside repo
    
    domain TEXT,
    buildpack_type TEXT NOT NULL CHECK(buildpack_type IN ('nixpacks', 'static', 'dockerfile', 'docker-compose', 'buildpacks')),
    config JSON DEFAULT '{}',
    auto_deploy INTEGER NOT NULL DEFAULT 1,
    
    status TEXT NOT NULL DEFAULT 'created' CHECK(status IN ('created', 'building', 'deploying', 'running', 'stopped', 'failed')),
    
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique application names within a project
    UNIQUE(name, project_id)
);

CREATE INDEX IF NOT EXISTS idx_applications_project_id ON applications(project_id);
CREATE INDEX IF NOT EXISTS idx_applications_environment_id ON applications(environment_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_applications_timestamp
    AFTER UPDATE ON applications
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE applications SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- Services table
CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id TEXT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    config TEXT DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique service names within a project
    UNIQUE(name, project_id)
);

CREATE INDEX IF NOT EXISTS idx_services_project_id ON services(project_id);
CREATE INDEX IF NOT EXISTS idx_services_environment_id ON services(environment_id);
CREATE INDEX IF NOT EXISTS idx_services_type ON services(type);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_services_timestamp
    AFTER UPDATE ON services
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE services SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- Deployments table
CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    
    deployment_number INTEGER NOT NULL,
    is_production INTEGER NOT NULL DEFAULT 0,
    triggered_by TEXT REFERENCES users(id),
    trigger_type TEXT NOT NULL DEFAULT 'manual' CHECK(trigger_type IN ('manual', 'git_push', 'api', 'scheduled', 'rollback')),
    
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'queued', 'building', 'deploying', 'running', 'failed', 'cancelled', 'stopped')), 
    
    container_id TEXT,
    image_tag TEXT NOT NULL,
    image_digest TEXT,
    
    git_commit_hash TEXT,
    git_commit_message TEXT,
    git_branch TEXT,
    git_author_name TEXT,
    
    build_logs TEXT,
    deploy_logs TEXT,
    error_message TEXT,
    
    -- Timing
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    build_started_at DATETIME,
    build_completed_at DATETIME,
    deploy_started_at DATETIME,
    deploy_completed_at DATETIME,
    stopped_at DATETIME,
    
    -- Resources used
    build_duration_seconds INTEGER,
    deploy_duration_seconds INTEGER,
    
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_deployments_application_id ON deployments(application_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_started_at ON deployments(started_at);
CREATE INDEX IF NOT EXISTS idx_deployments_number ON deployments(application_id, deployment_number);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_deployments_timestamp
    AFTER UPDATE ON deployments
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE deployments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- Environment variables table
CREATE TABLE IF NOT EXISTS environment_variables (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    environment_id TEXT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(application_id, environment_id, key)
);

CREATE INDEX IF NOT EXISTS idx_env_vars_application_id ON environment_variables(application_id);
CREATE INDEX IF NOT EXISTS idx_env_vars_environment_id ON environment_variables(environment_id);

-- +goose Down

-- Drop all triggers first
DROP TRIGGER IF EXISTS update_deployments_timestamp;
DROP TRIGGER IF EXISTS update_services_timestamp;
DROP TRIGGER IF EXISTS update_applications_timestamp;
DROP TRIGGER IF EXISTS update_environments_timestamp;
DROP TRIGGER IF EXISTS update_projects_timestamp;
DROP TRIGGER IF EXISTS update_users_timestamp;

-- Drop all indexes
DROP INDEX IF EXISTS idx_env_vars_environment_id;
DROP INDEX IF EXISTS idx_env_vars_application_id;
DROP INDEX IF EXISTS idx_deployments_number;
DROP INDEX IF EXISTS idx_deployments_started_at;
DROP INDEX IF EXISTS idx_deployments_status;
DROP INDEX IF EXISTS idx_deployments_application_id;
DROP INDEX IF EXISTS idx_services_type;
DROP INDEX IF EXISTS idx_services_environment_id;
DROP INDEX IF EXISTS idx_services_project_id;
DROP INDEX IF EXISTS idx_applications_status;
DROP INDEX IF EXISTS idx_applications_environment_id;
DROP INDEX IF EXISTS idx_applications_project_id;
DROP INDEX IF EXISTS idx_environments_project_id;
DROP INDEX IF EXISTS idx_projects_created_by;
DROP INDEX IF EXISTS idx_projects_user_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_sessions_active;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email;

-- Drop all tables in reverse dependency order
DROP TABLE IF EXISTS environment_variables;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS environments;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;

-- +goose Up

-- Insert default roles with permissions
INSERT INTO roles (id, name, description, permissions, created_at) VALUES
(
    'role-admin-00000000',
    'admin',
    'Full system access with all permissions',
    '["user:create","user:read","user:update","user:delete","project:create","project:read","project:update","project:delete","application:create","application:read","application:update","application:delete","deployment:create","deployment:read","deployment:cancel","service:create","service:read","service:update","service:delete","database:create","database:read","database:update","database:delete","environment:create","environment:read","environment:update","environment:delete","organization:manage","role:manage","settings:manage"]',
    CURRENT_TIMESTAMP
),
(
    'role-leader-00000000',
    'leader',
    'Team leadership with project management permissions',
    '["project:create","project:read","project:update","application:create","application:read","application:update","application:delete","deployment:create","deployment:read","deployment:cancel","service:create","service:read","service:update","database:create","database:read","database:update","environment:create","environment:read","environment:update"]',
    CURRENT_TIMESTAMP
),
(
    'role-member-00000000',
    'member',
    'Standard member with basic read and create permissions',
    '["project:read","application:read","application:create","deployment:create","deployment:read","service:read","database:read","environment:read"]',
    CURRENT_TIMESTAMP
);

-- Note: Default organization will be created in application code when first user is created
-- This ensures we have a valid owner_id reference

-- +goose Down

-- Remove default organization
DELETE FROM organizations WHERE slug = 'default';

-- Remove default roles
DELETE FROM roles WHERE id IN ('role-admin-00000000', 'role-leader-00000000', 'role-member-00000000');

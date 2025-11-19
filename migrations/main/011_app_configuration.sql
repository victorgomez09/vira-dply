-- +goose Up
ALTER TABLE applications ADD COLUMN generated_domain TEXT DEFAULT '';
ALTER TABLE applications ADD COLUMN exposed_ports TEXT DEFAULT '[]';
ALTER TABLE applications ADD COLUMN port_mappings TEXT DEFAULT '[]';

-- +goose Down
ALTER TABLE applications DROP COLUMN generated_domain;
ALTER TABLE applications DROP COLUMN exposed_ports;
ALTER TABLE applications DROP COLUMN port_mappings;

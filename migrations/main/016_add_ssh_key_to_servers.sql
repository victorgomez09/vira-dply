-- +goose Up
ALTER TABLE servers ADD COLUMN ssh_key TEXT DEFAULT '';

-- +goose Down
ALTER TABLE servers DROP COLUMN ssh_key;

-- +goose Up
ALTER TABLE servers ADD COLUMN tags TEXT DEFAULT '[]';

-- +goose Down
ALTER TABLE servers DROP COLUMN tags;

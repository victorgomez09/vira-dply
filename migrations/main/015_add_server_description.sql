-- +goose Up
ALTER TABLE servers ADD COLUMN description TEXT DEFAULT '';

-- +goose Down
ALTER TABLE servers DROP COLUMN description;

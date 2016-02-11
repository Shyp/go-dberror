-- +goose Up
ALTER TABLE accounts ADD COLUMN data JSONB;

-- +goose Down
ALTER TABLE accounts DROP COLUMN data;

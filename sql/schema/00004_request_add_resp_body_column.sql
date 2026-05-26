-- +goose Up
ALTER TABLE request ADD COLUMN response_body BLOB;

-- +goose Down
ALTER TABLE request DROP COLUMN response_body;

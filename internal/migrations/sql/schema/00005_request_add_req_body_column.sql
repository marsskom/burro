-- +goose Up
ALTER TABLE request ADD COLUMN request_body BLOB;

-- +goose Down
ALTER TABLE request DROP COLUMN request_body;

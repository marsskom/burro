-- +goose Up
CREATE TABLE workspace (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at INTEGER,
  updated_at INTEGER
);

-- +goose Down
DROP TABLE workspace;

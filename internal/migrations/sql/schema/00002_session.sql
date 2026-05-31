-- +goose Up
CREATE TABLE session (
  id TEXT PRIMARY KEY,
  name TEXT,
  description TEXT,
  metadata TEXT,
  created_at INTEGER,
  updated_at INTEGER
);

-- +goose Down
DROP TABLE session;

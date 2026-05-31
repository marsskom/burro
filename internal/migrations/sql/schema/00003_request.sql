-- +goose Up
CREATE TABLE request (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  host TEXT NOT NULL,
  url TEXT NOT NULL,
  method TEXT NOT NULL,
  request_raw BLOB NOT NULL,
  response_raw BLOB,
  start_time INTEGER NOT NULL,
  state INTEGER,
  is_finished INTEGER,
  metadata TEXT,
  created_at INTEGER,
  updated_at INTEGER
);

CREATE INDEX request_session_id_idx ON request(session_id);

-- +goose Down
DROP TABLE request;

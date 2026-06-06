-- +goose Up
CREATE TABLE request (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  start_time INTEGER NOT NULL,
  state INTEGER,
  is_finished INTEGER,
  metadata TEXT,
  created_at INTEGER,
  updated_at INTEGER,

  req_proto TEXT NOT NULL,
  req_host TEXT NOT NULL,
  req_method TEXT NOT NULL,
  req_scheme TEXT NOT NULL,
  req_url TEXT NOT NULL,
  req_path TEXT NOT NULL,
  req_query_params TEXT NOT NULL,
  req_headers TEXT NOT NULL,
  req_cookies TEXT NOT NULL,
  req_content_length INTEGER NOT NULL,
  req_remote_addr TEXT NOT NULL,
  req_body BLOB NOT NULL,

  res_proto TEXT DEFAULT NULL,
  res_status TEXT DEFAULT NULL,
  res_status_code INTEGER DEFAULT NULL,
  res_headers TEXT DEFAULT NULL,
  res_content_length INTEGER DEFAULT NULL,
  res_body BLOB DEFAULT NULL
);

CREATE INDEX request_session_id_idx ON request(session_id);
CREATE INDEX request_req_host_idx ON request(req_host);
CREATE INDEX request_req_method_idx ON request(req_method);
CREATE INDEX request_req_status_idx ON request(res_status);
CREATE INDEX request_req_status_code_idx ON request(res_status_code);

-- +goose Down
DROP TABLE request;

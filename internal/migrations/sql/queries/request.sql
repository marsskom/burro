-- name: GetRequest :one
SELECT * FROM request WHERE id = ?;

-- name: GetBySessionID :many
SELECT * FROM request WHERE session_id = ?;

-- name: CreateRequest :one
INSERT INTO request (
  id,
  session_id,
  start_time,
  state,
  is_finished,
  metadata,
  created_at,
  updated_at,

  req_proto,
  req_host,
  req_method,
  req_scheme,
  req_url,
  req_path,
  req_query_params,
  req_headers,
  req_cookies,
  req_content_length,
  req_remote_addr,
  req_body,

  res_proto,
  res_status,
  res_status_code,
  res_headers,
  res_content_length,
  res_body
)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateRequest :exec
UPDATE request
SET
  state = ?,
  is_finished = ?,
  metadata = ?,
  updated_at = ?,
  res_body = ?,
  res_status = ?,
  res_status_code = ?,
  res_proto = ?,
  res_headers = ?,
  res_content_length = ?
WHERE id = ?;

-- name: UpsertRequest :exec
INSERT INTO request (
  id,
  session_id,
  start_time,
  state,
  is_finished,
  metadata,
  created_at,
  updated_at,

  req_proto,
  req_host,
  req_method,
  req_scheme,
  req_url,
  req_path,
  req_query_params,
  req_headers,
  req_cookies,
  req_content_length,
  req_remote_addr,
  req_body,

  res_proto,
  res_status,
  res_status_code,
  res_headers,
  res_content_length,
  res_body
)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
state = excluded.state,
is_finished = excluded.is_finished,
metadata = excluded.metadata,
updated_at = excluded.updated_at,

res_body = excluded.res_body,
res_status = excluded.res_status,
res_status_code = excluded.res_status_code,
res_proto = excluded.res_proto,
res_headers = excluded.res_headers,
res_content_length = excluded.res_content_length
;

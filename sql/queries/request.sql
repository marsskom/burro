-- name: GetRequest :one
SELECT * FROM request WHERE id = ?;

-- name: GetBySessionID :many
SELECT * FROM request WHERE session_id = ?;

-- name: CreateRequest :one
INSERT INTO request (id, session_id, host, url, method, request_raw, request_body, response_raw, response_body, start_time, state, is_finished, metadata, created_at, updated_at, scheme, path, proto, headers, cookies, query_params, content_length, remote_addr, resp_status, resp_status_code, resp_proto, resp_headers, resp_content_length)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateRequest :exec
UPDATE request
SET state = ?, is_finished = ?, metadata = ?, updated_at = ?, response_raw = ?, response_body = ?, resp_status = ?, resp_status_code = ?, resp_proto = ?, resp_headers = ?, resp_content_length = ?
WHERE id = ?;

-- name: UpsertRequest :exec
INSERT INTO request (id, session_id, host, url, method, request_raw, request_body, response_raw, response_body, start_time, state, is_finished, metadata, created_at, updated_at, scheme, path, proto, headers, cookies, query_params, content_length, remote_addr, resp_status, resp_status_code, resp_proto, resp_headers, resp_content_length)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
state = excluded.state,
is_finished = excluded.is_finished,
metadata = excluded.metadata,
updated_at = excluded.updated_at,
response_raw = excluded.response_raw,
response_body = excluded.response_body,
resp_status = excluded.resp_status,
resp_status_code = excluded.resp_status_code,
resp_proto = excluded.resp_proto,
resp_headers = excluded.resp_headers,
resp_content_length = excluded.resp_content_length
;

-- name: GetRequest :one
SELECT * FROM request WHERE id = ?;

-- name: GetBySessionID :many
SELECT * FROM request WHERE session_id = ?;

-- name: CreateRequest :one
INSERT INTO request (id, session_id, host, url, method, request_raw, request_body, response_raw, response_body, start_time, state, is_finished, metadata, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateRequest :exec
UPDATE request
SET request_raw = ?, request_body = ?, response_raw = ?, response_body = ?, state = ?, is_finished = ?, metadata = ?, updated_at = ?
WHERE id = ?;

-- name: UpsertRequest :exec
INSERT INTO request (id, session_id, host, url, method, request_raw, request_body, response_raw, response_body, start_time, state, is_finished, metadata, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
request_raw = excluded.request_raw,
request_body = excluded.request_body,
response_raw = excluded.response_raw,
response_body = excluded.response_body,
state = excluded.state,
is_finished = excluded.is_finished,
metadata = excluded.metadata,
updated_at = excluded.updated_at;

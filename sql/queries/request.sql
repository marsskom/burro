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

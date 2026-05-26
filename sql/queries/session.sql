-- name: GetSession :one
SELECT * FROM session WHERE id = ?;

-- name: GetAll :many
SELECT * FROM session;

-- name: CreateSession :one
INSERT INTO session (id, name, description, metadata, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateSession :exec
UPDATE session
SET name = ?, description = ?, metadata = ?, updated_at = ?
WHERE id = ?;

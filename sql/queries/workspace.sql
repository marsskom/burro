-- name: GetWorkspace :one
SELECT * FROM workspace WHERE id = ?;

-- name: GetWorkspaceByName :one
SELECT * FROM workspace WHERE name = ?;

-- name: CreateWorkspace :one
INSERT INTO workspace (id, name, created_at, updated_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateWorkspace :exec
UPDATE workspace SET updated_at = ? WHERE id = ?;

-- name: UpsertWorkspace :exec
INSERT INTO workspace (id, name, created_at, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
updated_at = excluded.updated_at;

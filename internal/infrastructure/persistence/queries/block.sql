-- name: CreateBlock :exec
INSERT INTO blocks (id, profile_id, type, content, position, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListBlocksByProfile :many
SELECT * FROM blocks WHERE profile_id = $1 ORDER BY position ASC, created_at ASC;

-- name: ListActiveBlocksByProfile :many
SELECT * FROM blocks WHERE profile_id = $1 AND is_active = true ORDER BY position ASC, created_at ASC;

-- name: UpdateBlock :exec
UPDATE blocks
SET type = $2, content = $3, is_active = $4, updated_at = $5
WHERE id = $1 AND profile_id = $6;

-- name: UpdateBlockPosition :exec
UPDATE blocks SET position = $2, updated_at = $3 WHERE id = $1 AND profile_id = $4;

-- name: DeleteBlock :exec
DELETE FROM blocks WHERE id = $1 AND profile_id = $2;

-- name: IncrementBlockClick :exec
UPDATE blocks SET click_count = click_count + 1 WHERE id = $1;
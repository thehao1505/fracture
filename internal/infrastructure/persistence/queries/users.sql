-- name: GetUserByID :one
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :exec
INSERT INTO users (id, email, password, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateUser :exec
UPDATE users
SET email = $1, name = $2, updated_at = $3
WHERE id = $4;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE
    @keyword::text = ''
    OR name  ILIKE '%' || @keyword || '%'
    OR email ILIKE '%' || @keyword || '%'
ORDER BY created_at DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users
WHERE
    @keyword::text = ''
    OR name  ILIKE '%' || @keyword || '%'
    OR email ILIKE '%' || @keyword || '%';

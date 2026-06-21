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


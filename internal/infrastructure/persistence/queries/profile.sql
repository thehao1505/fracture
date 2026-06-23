-- name: CreateProfile :exec
INSERT INTO profiles (id, user_id, username, display_name, bio, avatar_url, appearance, is_published, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetProfileByUserID :one
SELECT * FROM profiles WHERE user_id = $1;

-- name: GetPublishedProfileByUsername :one
SELECT * FROM profiles WHERE username = $1 AND is_published = true;

-- name: UpdateProfile :exec
UPDATE profiles
SET username = $2, display_name = $3, bio = $4, avatar_url = $5,
    appearance = $6, is_published = $7, updated_at = $8
WHERE id = $1;

-- name: ProfileUsernameExists :one
SELECT EXISTS (SELECT 1 FROM profiles WHERE username = $1 AND id <> $2);
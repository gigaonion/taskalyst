-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, name, role
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: UpdateUserPreferences :one
UPDATE users
SET
    preferences = COALESCE(sqlc.narg('preferences'), preferences),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

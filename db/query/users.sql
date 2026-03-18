-- name: GetUserByID :one
SELECT id, email, username, created_at, updated_at
FROM users
WHERE id = $1;  

-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, email, username, created_at, updated_at;


-- name: ListUsers :many
SELECT id, email, username, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET username = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, email, username, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
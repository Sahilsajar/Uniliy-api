-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :exec
INSERT INTO users (username, email, name, password_hash, course, yop)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token_hash = $1;-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserFollowers :many
SELECT u.*
FROM users u
JOIN user_follows uf ON u.id = uf.follower_user_id
WHERE uf.following_user_id = $1;

-- name: GetUserFollowing :many
SELECT u.*
FROM users u
JOIN user_follows uf ON u.id = uf.following_user_id
WHERE uf.follower_user_id = $1;
-- name: CreatePost :one
INSERT INTO posts (title, body, status, user_id)
VALUES ($1, $2, $3, $4)
RETURNING id, title, body, status, user_id, created_at, updated_at;

-- name: CreatePostTag :exec
INSERT INTO post_tags (post_id, user_id, tagged_by)
VALUES ($1, $2, $3);

-- name: GetPostByID :one
SELECT id, title, body, status, user_id, created_at, updated_at
FROM posts
WHERE id = $1;

-- name: CreatePostTagsBulk :exec
INSERT INTO post_tags (post_id, user_id, tagged_by)
SELECT $1, unnest($2::bigint[]), $3
ON CONFLICT DO NOTHING;

-- name: CreatePostImagesBulk :exec
INSERT INTO post_images (post_id, media_id, image_url)
SELECT $1, m.id, m.url
FROM media m
WHERE m.id = ANY($2::bigint[]);

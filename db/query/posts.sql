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

-- name: GetPostDetailsByID :one
SELECT
    p.id,
    p.title,
    p.body,
    p.status,
    p.user_id,
    p.created_at,
    p.updated_at,
    u.username,
    u.name,
    u.profile_pic,
    COALESCE(pl.likes_count, 0)::bigint AS likes_count,
    COALESCE(pc.comments_count, 0)::bigint AS comments_count,
    EXISTS (
        SELECT 1
        FROM post_likes viewer_like
        WHERE viewer_like.post_id = p.id
          AND viewer_like.user_id = $2
    ) AS is_liked
FROM posts p
JOIN users u ON u.id = p.user_id
LEFT JOIN (
    SELECT post_id, COUNT(*)::bigint AS likes_count
    FROM post_likes
    GROUP BY post_id
) pl ON pl.post_id = p.id
LEFT JOIN (
    SELECT post_id, COUNT(*)::bigint AS comments_count
    FROM comments
    GROUP BY post_id
) pc ON pc.post_id = p.id
WHERE p.id = $1;

-- name: ListFeedPosts :many
SELECT
    p.id,
    p.title,
    p.body,
    p.status,
    p.user_id,
    p.created_at,
    p.updated_at,
    u.username,
    u.name,
    u.profile_pic,
    COALESCE(pl.likes_count, 0)::bigint AS likes_count,
    COALESCE(pc.comments_count, 0)::bigint AS comments_count,
    EXISTS (
        SELECT 1
        FROM post_likes viewer_like
        WHERE viewer_like.post_id = p.id
          AND viewer_like.user_id = sqlc.arg(viewer_user_id)
    ) AS is_liked
FROM posts p
JOIN users u ON u.id = p.user_id
LEFT JOIN (
    SELECT post_id, COUNT(*)::bigint AS likes_count
    FROM post_likes
    GROUP BY post_id
) pl ON pl.post_id = p.id
LEFT JOIN (
    SELECT post_id, COUNT(*)::bigint AS comments_count
    FROM comments
    GROUP BY post_id
) pc ON pc.post_id = p.id
WHERE p.status = 'published'
  AND (
      sqlc.arg(scope) = 'all'
      OR (sqlc.arg(scope) = 'mine' AND p.user_id = sqlc.arg(viewer_user_id))
      OR (
          sqlc.arg(scope) = 'following'
          AND (
              p.user_id = sqlc.arg(viewer_user_id)
              OR EXISTS (
                  SELECT 1
                  FROM user_follows uf
                  WHERE uf.follower_user_id = sqlc.arg(viewer_user_id)
                    AND uf.following_user_id = p.user_id
              )
          )
      )
  )
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: CountFeedPosts :one
SELECT COUNT(*)::bigint
FROM posts p
WHERE p.status = 'published'
  AND (
      sqlc.arg(scope) = 'all'
      OR (sqlc.arg(scope) = 'mine' AND p.user_id = sqlc.arg(viewer_user_id))
      OR (
          sqlc.arg(scope) = 'following'
          AND (
              p.user_id = sqlc.arg(viewer_user_id)
              OR EXISTS (
                  SELECT 1
                  FROM user_follows uf
                  WHERE uf.follower_user_id = sqlc.arg(viewer_user_id)
                    AND uf.following_user_id = p.user_id
              )
          )
      )
  );

-- name: CreatePostTagsBulk :exec
INSERT INTO post_tags (post_id, user_id, tagged_by)
SELECT $1, unnest($2::bigint[]), $3
ON CONFLICT DO NOTHING;

-- name: CreatePostImagesBulk :exec
INSERT INTO post_images (post_id, media_id, image_url)
SELECT $1, m.id, m.url
FROM media m
WHERE m.id = ANY($2::bigint[]);

-- name: AddComment :one
INSERT INTO comments (message, post_id, user_id, parent_comment_id)
VALUES ($1, $2, $3, $4)
RETURNING id, post_id, user_id, message, parent_comment_id, created_at, updated_at;

-- name: GetCommentsByPostID :many
SELECT c.id, c.post_id, c.user_id, c.message, c.parent_comment_id, c.created_at, c.updated_at, u.username, u.name, u.profile_pic
FROM comments c
LEFT JOIN users u ON u.id = c.user_id
WHERE c.post_id = $1
ORDER BY c.created_at ASC, c.id ASC;

-- name: GetCommentByID :one
SELECT c.id, c.post_id, c.user_id, c.message, c.parent_comment_id, c.created_at, c.updated_at, u.username, u.name, u.profile_pic
FROM comments c
LEFT JOIN users u ON u.id = c.user_id
WHERE c.id = $1;

-- name: LikePost :exec
INSERT INTO post_likes (post_id, user_id)
VALUES ($1, $2);

-- name: UnlikePost :exec
DELETE FROM post_likes
WHERE post_id = $1 AND user_id = $2;

-- name: CheckPostLikeExists :one
SELECT EXISTS (
    SELECT 1 FROM post_likes
    WHERE post_id = $1 AND user_id = $2
);

-- name: CreateMedia :one
INSERT INTO media (public_id, url, user_id, is_temp)
VALUES ($1, $2, $3, $4)
RETURNING id, public_id, url, user_id, is_temp, created_at;

-- name: GetUserTempMediaByIDs :many
SELECT id, public_id, url, user_id, is_temp, created_at
FROM media
WHERE user_id = $1
  AND is_temp = TRUE
  AND id = ANY($2::bigint[]);

-- name: MarkMediaPermanentByIDs :exec
UPDATE media
SET is_temp = FALSE
WHERE user_id = $1
  AND id = ANY($2::bigint[]);

-- name: CreatePostImage :exec
INSERT INTO post_images (image_url, media_id, post_id)
VALUES ($1, $2, $3);

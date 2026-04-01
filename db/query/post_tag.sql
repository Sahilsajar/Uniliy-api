-- name: GetTaggedUserIDs :many
SELECT user_id
FROM post_tags
WHERE post_id = $1; 
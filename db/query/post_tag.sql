-- name: GetTaggedUserIDs :many
SELECT user_id
FROM post_tags
WHERE post_id = $1; 

-- name: GetTaggedUsersByPostIDs :many
SELECT
    pt.post_id,
    u.id,
    u.username,
    u.name,
    u.profile_pic
FROM post_tags pt
JOIN users u ON u.id = pt.user_id
WHERE pt.post_id = ANY($1::bigint[])
ORDER BY pt.post_id, pt.id;

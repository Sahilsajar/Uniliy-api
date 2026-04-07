-- name: GetPostImageURLs :many
SELECT image_url
FROM post_images
WHERE post_id = $1;

-- name: GetPostImageURLsByPostIDs :many
SELECT post_id, image_url
FROM post_images
WHERE post_id = ANY($1::bigint[])
ORDER BY post_id, id;

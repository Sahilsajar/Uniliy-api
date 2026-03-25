
ALTER TABLE post_images
DROP CONSTRAINT fk_post_images_media;

ALTER TABLE post_images
DROP COLUMN media_id;

DELETE TABLE IF EXISTS media;
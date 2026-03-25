ALTER TABLE post_images
ADD COLUMN media_id BIGINT;

ALTER TABLE post_images
ADD CONSTRAINT fk_post_images_media
FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE;

ALTER TABLE post_images
DROP COLUMN image_url;
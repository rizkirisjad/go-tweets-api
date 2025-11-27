-- migrate:up
-- 1. Rename table
RENAME TABLE post_like TO post_likes;

-- 2. Drop old constraints 
ALTER TABLE post_likes
    DROP FOREIGN KEY fk_user_id_post_like,
    DROP FOREIGN KEY fk_post_id_post_like;

-- 3. Add new constraints 
ALTER TABLE post_likes
    ADD CONSTRAINT fk_post_likes_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    ADD CONSTRAINT fk_post_likes_post_id FOREIGN KEY (post_id) REFERENCES posts(id);

-- migrate:down
-- 1. Drop new constraints
ALTER TABLE post_likes
    DROP FOREIGN KEY fk_post_likes_user_id,
    DROP FOREIGN KEY fk_post_likes_post_id;

-- 2. Add old constraints 
ALTER TABLE post_likes
    ADD CONSTRAINT fk_user_id_post_like FOREIGN KEY (user_id) REFERENCES users(id),
    ADD CONSTRAINT fk_post_id_post_like FOREIGN KEY (post_id) REFERENCES posts(id);

-- 3. Rename table 
RENAME TABLE post_likes TO post_like;

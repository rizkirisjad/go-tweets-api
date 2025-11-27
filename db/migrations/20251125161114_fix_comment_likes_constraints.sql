-- migrate:up
-- 1. Drop old constraints on correct table
ALTER TABLE comment_likes
    DROP FOREIGN KEY fk_comment_id_comment_like,
    DROP FOREIGN KEY fk_user_id_comment_like;

-- 2. Add new consistent constraints
ALTER TABLE comment_likes
    ADD CONSTRAINT fk_comment_likes_comment_id FOREIGN KEY (comment_id) REFERENCES comments(id),
    ADD CONSTRAINT fk_comment_likes_user_id FOREIGN KEY (user_id) REFERENCES users(id);

-- migrate:down
-- 1. Drop new constraints
ALTER TABLE comment_likes
    DROP FOREIGN KEY fk_comment_likes_comment_id,
    DROP FOREIGN KEY fk_comment_likes_user_id;

-- 2. Restore old constraints
ALTER TABLE comment_likes
    ADD CONSTRAINT fk_comment_id_comment_like FOREIGN KEY (comment_id) REFERENCES comments(id),
    ADD CONSTRAINT fk_user_id_comment_like FOREIGN KEY (user_id) REFERENCES users(id);

-- migrate:up
CREATE TABLE IF NOT EXISTS post_like (
    id INT AUTO_INCREMENT PRIMARY KEY,
    post_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_id_post_like FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_post_id_post_like FOREIGN KEY (post_id) REFERENCES posts(id)
)

-- migrate:down
DROP TABLE IF NOT EXISTS post_like

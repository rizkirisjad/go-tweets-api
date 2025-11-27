-- migrate:up
ALTER TABLE refresh_tokens ADD COLUMN expired_at TIMESTAMP NULL AFTER refresh_token

-- migrate:down
ALTER TABLE refresh_tokens DROP COLUMN expired_at

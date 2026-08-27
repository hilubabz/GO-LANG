-- +goose up
CREATE TABLE refresh_tokens (token TEXT PRIMARY KEY, created_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL, user_id UUID NOT NULL, FOREIGN KEY (user_id) REFERENCES users(id),expires_at TIMESTAMP NOT NULL,revoked_at TIMESTAMP);

-- +goose down
DROP TABLE refresh_tokens;
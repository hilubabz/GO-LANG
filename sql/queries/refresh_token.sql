-- name: AddRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at) VALUES ($1,now(), now(), $2, $3, NULL) RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1 AND revoked_at IS NULL AND expires_at > now();

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET revoked_at = now(), updated_at = now() WHERE token = $1 RETURNING *;
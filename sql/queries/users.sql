-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
RETURNING *;

-- name: GetUser :many
SELECT * from users;

-- name: DeleteUser :exec
DELETE from users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1; 
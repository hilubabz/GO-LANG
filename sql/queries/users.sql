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

-- name: UpdateUser :one
UPDATE users SET email=$1,hashed_password=$2 WHERE id = $3 RETURNING *;

-- name: UpdateUserToChirpyRed :one
UPDATE users SET is_chirpy_red=true WHERE id = $1 RETURNING *;
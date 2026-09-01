-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetPassHashByUsername :one
SELECT id, password_hash FROM users
WHERE
    username = $1 OR
    email = $1;

-- name: CreateUser :one
INSERT INTO users (
    username, email, password_hash
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
    set password_hash = $2
WHERE $1;

-- name: UpdateLoginTime :exec
UPDATE users
SET last_login = NOW()
WHERE $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
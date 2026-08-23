-- name: GetUserData :one
SELECT username, mail, password, created_at , last_access
FROM user_table
WHERE username = $1 LIMIT 1;

-- name: GetAuthData :one
SELECT id, username, password
FROM user_table
WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT id, username, mail, created_at, last_access, mail_token
FROM user_table
WHERE id = $1 LIMIT 1;

-- name: TouchLastAccess :exec
UPDATE user_table
SET last_access = NOW()
WHERE id = $1;

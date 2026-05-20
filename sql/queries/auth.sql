-- name: GetUserData :one
SELECT username, mail, password, created_at , last_access
FROM user_table
WHERE username LIKE $1 LIMIT 1;

-- name: GetAuthData :one
SELECT id, username, password
FROM user_table
WHERE username LIKE $1 LIMIT 1;
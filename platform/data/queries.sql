-- -- name: UpdateUserEmail :one
-- UPDATE users SET
-- email = sqlc.arg(email), updated_at = unixepoch()
-- WHERE id = sqlc.arg(userId) RETURNING *;

-- name: NewUser :one
INSERT INTO users (username, password) VALUES (?, ?) RETURNING id;

-- name: UpdateUserPassword :one
UPDATE users SET
password = sqlc.arg(newPassword), updated_at = unixepoch()
WHERE id = sqlc.arg(userId) RETURNING *;

-- name: GetUsers :many
SELECT * FROM users ORDER BY username;

-- name: 


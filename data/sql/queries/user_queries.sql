-- name: GetUserById :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY name;

-- name: IsUserExist :one
SELECT EXISTS(SELECT id FROM users WHERE email = ? LIMIT 1);

-- name: InsertUser :execlastid
INSERT INTO users (
  email, username, user_detail, privilege_type
) VALUES (
  ?, ?, ?, ?
);

-- name: ActivateUser :exec
UPDATE users SET is_deactivated = 0 WHERE id = ?;

-- name: DeactivateUser :exec
UPDATE users SET is_deactivated = 1 WHERE id = ?;

-- name: DeleteUserById :exec
DELETE FROM users
WHERE id = ?;

-- name: GetUserCredentialById :one
SELECT user_id, email, password FROM user_credentials WHERE user_id = ?;

-- name: GetUserCredentialByEmail :one
SELECT user_id, email, password FROM user_credentials WHERE email = ?;

-- name: InsertUserCredential :execlastid
INSERT INTO user_credentials (
  user_id, email, password
) VALUES (
  ?, ?, ?
);

-- name: UpdatePasswordByUserId :exec
UPDATE user_credentials SET password = ? WHERE user_id = ?;

-- name: DeleteUserCredentialByUserId :exec
DELETE FROM user_credentials
WHERE user_id = ?;

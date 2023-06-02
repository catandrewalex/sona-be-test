/* ============================== USER ============================== */
-- name: GetUserById :one
SELECT * FROM user
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM user
WHERE email = ? LIMIT 1;

-- name: GetUsersByIds :many
SELECT * from user
WHERE id IN (sqlc.slice('ids'));

-- name: GetUsers :many
SELECT * FROM user
ORDER BY id
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT Count(*) as total FROM user;

-- name: IsUserExist :one
SELECT EXISTS(SELECT id FROM user WHERE email = ? LIMIT 1);

-- name: InsertUser :execlastid
INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  ?, ?, ?, ?
);

-- name: UpdateUser :exec
UPDATE user SET username = ?, email = ?, user_detail = ?, privilege_type = ?, is_deactivated = ? WHERE id = ?;

-- name: UpdateUserInfo :exec
UPDATE user SET username = ?, email = ?, user_detail = ? WHERE id = ?;

-- name: UpdateUserPrivilege :exec
UPDATE user SET privilege_type = ? WHERE id = ?;

-- name: ActivateUser :exec
UPDATE user SET is_deactivated = 0 WHERE id = ?;

-- name: DeactivateUser :exec
UPDATE user SET is_deactivated = 1 WHERE id = ?;

-- name: DeleteUserById :exec
DELETE FROM user
WHERE id = ?;

/* ============================== USER_CREDENTIAL ============================== */
-- name: GetUserCredentialById :one
SELECT * FROM user_credential WHERE user_id = ?;

-- name: GetUserCredentialByEmail :one
SELECT * FROM user_credential WHERE email = ?;

-- name: GetUserCredentialByUsername :one
SELECT * FROM user_credential WHERE username = ?;

-- name: InsertUserCredential :execlastid
INSERT INTO user_credential (
  user_id, username, email, password
) VALUES (
  ?, ?, ?, ?
);

-- name: UpdateUserCredentialInfoByUserId :exec
UPDATE user_credential SET username = ?, email = ? WHERE user_id = ?;

-- name: UpdatePasswordByUserId :exec
UPDATE user_credential SET password = ? WHERE user_id = ?;

-- name: DeleteUserCredentialByUserId :exec
DELETE FROM user_credential
WHERE user_id = ?;

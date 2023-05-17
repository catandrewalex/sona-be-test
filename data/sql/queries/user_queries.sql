/* ============================== USER ============================== */
-- name: GetUserById :one
SELECT * FROM user
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM user
WHERE email = ? LIMIT 1;

-- name: GetUser :many
SELECT * FROM user
ORDER BY name;

-- name: IsUserExist :one
SELECT EXISTS(SELECT id FROM user WHERE email = ? LIMIT 1);

-- name: InsertUser :execlastid
INSERT INTO user (
  email, username, user_detail, privilege_type
) VALUES (
  ?, ?, ?, ?
);

-- name: UpdateUserInfo :exec
UPDATE user SET email = ?, username = ?, user_detail = ? WHERE id = ?;

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
SELECT user_id, email, password FROM user_credential WHERE user_id = ?;

-- name: GetUserCredentialByEmail :one
SELECT user_id, email, password FROM user_credential WHERE email = ?;

-- name: InsertUserCredential :execlastid
INSERT INTO user_credential (
  user_id, email, password
) VALUES (
  ?, ?, ?
);

-- name: UpdateEmailByUserId :exec
UPDATE user_credential SET email = ? WHERE user_id = ?;

-- name: UpdatePasswordByUserId :exec
UPDATE user_credential SET password = ? WHERE user_id = ?;

-- name: DeleteUserCredentialByUserId :exec
DELETE FROM user_credential
WHERE user_id = ?;

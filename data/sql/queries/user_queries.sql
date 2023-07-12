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
WHERE is_deactivated IN (sqlc.slice('isDeactivateds'))
ORDER BY id
LIMIT ? OFFSET ?;

-- name: GetUsersNotTeacher :many
SELECT sqlc.embed(user) FROM user
LEFT JOIN teacher on user.id = teacher.user_id
WHERE is_deactivated IN (sqlc.slice('isDeactivateds')) AND teacher.user_id IS NULL
ORDER BY user.id
LIMIT ? OFFSET ?;

-- name: GetUsersNotStudent :many
SELECT sqlc.embed(user) FROM user
LEFT JOIN student on user.id = student.user_id
WHERE is_deactivated IN (sqlc.slice('isDeactivateds')) AND student.user_id IS NULL
ORDER BY user.id
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT Count(*) AS total FROM user
WHERE is_deactivated IN (sqlc.slice('isDeactivateds'));

-- name: CountUsersNotTeacher :one
SELECT Count(*) AS total FROM user
LEFT JOIN teacher on user.id = teacher.user_id
WHERE is_deactivated IN (sqlc.slice('isDeactivateds')) AND teacher.user_id IS NULL;

-- name: CountUsersNotStudent :one
SELECT Count(*) AS total FROM user
LEFT JOIN student on user.id = student.user_id
WHERE is_deactivated IN (sqlc.slice('isDeactivateds')) AND student.user_id IS NULL;

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

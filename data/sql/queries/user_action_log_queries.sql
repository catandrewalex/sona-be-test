-- name: GetUserActionLogById :one
SELECT * FROM user_action_log
WHERE id = ? LIMIT 1;

-- name: GetUserActionLogs :many
SELECT user_action_log.id AS id, date, user_id, user.username AS username, user_action_log.privilege_type, endpoint, method, status_code, request_body
FROM user_action_log
  LEFT JOIN user ON user_id = user.id
WHERE
  date >= sqlc.arg('startDate') AND date <= sqlc.arg('endDate')
  AND (user_id = sqlc.arg('user_id') OR sqlc.arg('use_user_id_filter') = false)
  AND (user_action_log.privilege_type = sqlc.arg('privilege_type') OR sqlc.arg('use_privilege_type_filter') = false)
  AND (method = sqlc.arg('method') OR sqlc.arg('use_method_filter') = false)
  AND (status_code = sqlc.arg('status_code') OR sqlc.arg('use_status_code_filter') = false)
ORDER BY date DESC
LIMIT ? OFFSET ?;

-- name: GetUserActionLogsByUserId :many
SELECT * FROM user_action_log
WHERE user_id = ?
ORDER BY date DESC
LIMIT ? OFFSET ?;

-- name: CountUserActionLogs :one
SELECT Count(*) AS total FROM user_action_log
WHERE
  date >= sqlc.arg('startDate') AND date <= sqlc.arg('endDate')
  AND (user_id = sqlc.arg('user_id') OR sqlc.arg('use_user_id_filter') = false)
  AND (privilege_type = sqlc.arg('privilege_type') OR sqlc.arg('use_privilege_type_filter') = false)
  AND (method = sqlc.arg('method') OR sqlc.arg('use_method_filter') = false)
  AND (status_code = sqlc.arg('status_code') OR sqlc.arg('use_status_code_filter') = false);

-- name: InsertUserActionLog :execlastid
INSERT INTO user_action_log (
  date, user_id, privilege_type, endpoint, method, status_code, request_body) VALUES (
  ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateUserActionLog :exec
UPDATE user_action_log SET date = ?, user_id = ?, privilege_type = ?, endpoint = ?, method = ?, status_code = ?, request_body = ? WHERE id = ?;

-- name: DeleteUserActionLogsByIds :exec
DELETE FROM user_action_log
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteUserActionLogs :execrows
DELETE FROM user_action_log
WHERE 
  date >= sqlc.arg('startDate') AND date <= sqlc.arg('endDate')
  AND (user_id = sqlc.arg('user_id') OR sqlc.arg('use_user_id_filter') = false)
  AND (privilege_type = sqlc.arg('privilege_type') OR sqlc.arg('use_privilege_type_filter') = false)
  AND (method = sqlc.arg('method') OR sqlc.arg('use_method_filter') = false)
  AND (status_code = sqlc.arg('status_code') OR sqlc.arg('use_status_code_filter') = false);

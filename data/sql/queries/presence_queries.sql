/* ============================== PRESENCE ============================== */
-- name: GetPresenceById :one
SELECT * FROM presence
WHERE id = ? LIMIT 1;

-- name: GetPresencesByClassId :many
SELECT * FROM presence
WHERE class_id = ?;

-- name: GetPresencesByTeacherId :many
SELECT * FROM presence
WHERE teacher_id = ?;

-- name: InsertPresence :execlastid
INSERT INTO presence (
    date, used_student_token_quota, duration, class_id, teacher_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: DeletePresenceById :exec
DELETE FROM presence
WHERE id = ?;

/* ============================== STUDENT_ATTEND ============================== */
-- name: GetStudentAttendsByStudentId :many
SELECT * FROM student_attend
WHERE student_id = ?;

-- name: GetStudentAttendsByPresenceId :many
SELECT * FROM student_attend
WHERE presence_id = ?;

-- name: InsertStudentAttend :exec
INSERT INTO student_attend (
    student_id, presence_id
) VALUES (
    ?, ?
);

-- name: DeleteStudentAttend :exec
DELETE FROM student_attend
WHERE student_id = ? AND presence_id = ?;

-- name: DeleteStudentAttendByStudentId :exec
DELETE FROM student_attend
WHERE student_id = ?;

-- name: DeleteStudentAttendByPresenceId :exec
DELETE FROM student_attend
WHERE presence_id = ?;

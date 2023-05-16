/* ============================== TEACHER ============================== */
-- name: GetTeacherById :one
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
WHERE teacher.id = ? LIMIT 1;

-- name: GetTeacherByUserId :one
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
WHERE user_id = ? LIMIT 1;

-- name: GetTeachers :many
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at, Count(user_id) as total_results
FROM teacher JOIN user ON teacher.user_id = user.id
ORDER BY username
LIMIT ? OFFSET ?;

-- name: CountTeachers :one
SELECT Count(user_id) as total_results FROM teacher;

-- name: InsertTeacher :execlastid
INSERT INTO teacher ( user_id ) VALUES ( ? );

-- name: DeleteTeacherById :exec
DELETE FROM teacher
WHERE id = ?;

-- name: DeleteTeacherByUserId :exec
DELETE FROM teacher
WHERE user_id = ?;

/* ============================== STUDENT ============================== */
-- name: GetStudentById :one
SELECT student.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM student JOIN user ON student.user_id = user.id
WHERE student.id = ? LIMIT 1;

-- name: GetStudentByUserId :one
SELECT student.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM student JOIN user ON student.user_id = user.id
WHERE user_id = ? LIMIT 1;

-- name: GetStudents :many
SELECT student.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM student JOIN user ON student.user_id = user.id
ORDER BY username
LIMIT ? OFFSET ?;

-- name: CountStudents :one
SELECT Count(user_id) as total_results FROM student;

-- name: InsertStudent :execlastid
INSERT INTO student ( user_id ) VALUES ( ? );

-- name: DeleteStudentById :exec
DELETE FROM student
WHERE id = ?;

-- name: DeleteStudentByUserId :exec
DELETE FROM student
WHERE user_id = ?;

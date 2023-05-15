/* ============================== TEACHER ============================== */
-- name: GetTeacherById :one
SELECT * FROM teacher
WHERE id = ? LIMIT 1;

-- name: GetTeacherByUserId :one
SELECT * FROM teacher
WHERE user_id = ? LIMIT 1;

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
SELECT * FROM student
WHERE id = ? LIMIT 1;

-- name: GetStudentByUserId :one
SELECT * FROM student
WHERE user_id = ? LIMIT 1;

-- name: InsertStudent :execlastid
INSERT INTO student ( user_id ) VALUES ( ? );

-- name: DeleteStudentById :exec
DELETE FROM student
WHERE id = ?;

-- name: DeleteStudentByUserId :exec
DELETE FROM student
WHERE user_id = ?;

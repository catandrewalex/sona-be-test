/* ============================== ATTENDANCE ============================== */
-- name: GetAttendanceIdsOfSameClassAndDate :many
WITH ref_attendance AS (
    SELECT *
    FROM attendance
    WHERE attendance.id = ?
)
SELECT attendance.id AS id, attendance.is_paid AS is_paid
FROM attendance
    JOIN ref_attendance ON attendance.class_id = ref_attendance.class_id AND attendance.date = ref_attendance.date
ORDER by attendance.id;

-- name: SetAttendancesIsPaidStatusByIds :exec
UPDATE attendance SET is_paid = ?
WHERE id IN (sqlc.slice('ids'));

-- name: GetAttendancesByTeacherId :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= sqlc.arg('startDate') AND attendance.date <= sqlc.arg('endDate'))
    AND (attendance.teacher_id = sqlc.arg('teacher_id') OR sqlc.arg('use_teacher_filter') = false)
    AND is_paid = 0
ORDER BY attendance.teacher_id, class.id, attendance.student_id, date, attendance.id;

-- name: GetAttendanceById :one
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id = ? LIMIT 1;

-- name: GetAttendancesByIds :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id IN (sqlc.slice('ids'));

-- name: GetAttendances :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= sqlc.arg('startDate') AND attendance.date <= sqlc.arg('endDate'))
    AND (class_id = sqlc.arg('class_id') OR sqlc.arg('use_class_filter') = false)
    AND (student_id = sqlc.arg('student_id') OR sqlc.arg('use_student_filter') = false)
    AND (is_paid = 0 OR sqlc.arg('use_unpaid_filter') = false)
ORDER BY class.id
LIMIT ? OFFSET ?;

-- name: CountAttendances :one
SELECT Count(id) AS total FROM attendance
WHERE
    (date >= sqlc.arg('startDate') AND date <= sqlc.arg('endDate'))
    AND (class_id = sqlc.arg('class_id') OR sqlc.arg('use_class_filter') = false)
    AND (student_id = sqlc.arg('student_id') OR sqlc.arg('use_student_filter') = false)
    AND (is_paid = 0 OR sqlc.arg('use_unpaid_filter') = false);

-- name: CountAttendancesByIds :one
SELECT Count(id) AS total FROM attendance
WHERE id IN (sqlc.slice('ids'));

-- name: InsertAttendance :execlastid
INSERT INTO attendance (
    date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateAttendance :exec
UPDATE attendance
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, is_paid = ?, class_id = ?, teacher_id = ?, student_id = ?, token_id = ?
WHERE id = ?;

-- name: EditAttendances :exec
UPDATE attendance
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, teacher_id = ?
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteAttendanceById :exec
DELETE FROM attendance
WHERE id = ?;

-- name: DeleteAttendancesByIds :exec
DELETE FROM attendance
WHERE id IN (sqlc.slice('ids'));

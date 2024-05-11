/* ============================== ATTENDANCE ============================== */
-- name: GetAttendanceIdsOfSameClassAndDate :many
WITH ref_attendance AS (
    SELECT *
    FROM attendance
    WHERE attendance.id = ?
)
SELECT attendance.id AS id, attendance.is_paid AS is_paid, attendance.token_id, attendance.used_student_token_quota
FROM attendance
    JOIN ref_attendance ON attendance.class_id = ref_attendance.class_id AND attendance.date = ref_attendance.date
ORDER by attendance.id;

-- name: SetAttendancesIsPaidStatusByIds :exec
UPDATE attendance SET is_paid = ?
WHERE id IN (sqlc.slice('ids'));

-- name: GetUnpaidAttendancesByTeacherId :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
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
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= sqlc.arg('startDate') AND attendance.date <= sqlc.arg('endDate'))
    AND attendance.teacher_id = sqlc.arg('teacher_id')
    AND is_paid = 0
ORDER BY date DESC, attendance.id ASC;

-- name: GetAttendanceById :one
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
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
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id = ? LIMIT 1;

-- name: GetAttendancesByIds :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
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
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id IN (sqlc.slice('ids'));

-- name: GetAttendances :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
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
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= sqlc.arg('startDate') AND attendance.date <= sqlc.arg('endDate'))
    AND (class_id = sqlc.arg('class_id') OR sqlc.arg('use_class_filter') = false)
    AND (student_id = sqlc.arg('student_id') OR sqlc.arg('use_student_filter') = false)
    AND (is_paid = 0 OR sqlc.arg('use_unpaid_filter') = false)
ORDER BY attendance.id
LIMIT ? OFFSET ?;

-- name: GetAttendancesDescendingDate :many
-- GetAttendancesDescendingDate is a copy of GetAttendances, with additional sort by date parameter. TODO: find alternative: sqlc's dynamic query which is mature enough, so that we need to do this.
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
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
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= sqlc.arg('startDate') AND attendance.date <= sqlc.arg('endDate'))
    AND (class_id = sqlc.arg('class_id') OR sqlc.arg('use_class_filter') = false)
    AND (student_id = sqlc.arg('student_id') OR sqlc.arg('use_student_filter') = false)
    AND (is_paid = 0 OR sqlc.arg('use_unpaid_filter') = false)
ORDER BY attendance.date DESC, attendance.id ASC
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

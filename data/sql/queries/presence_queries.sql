/* ============================== PRESENCE ============================== */
-- name: GetPresenceById :one
SELECT presence.id AS presence_id, date, used_student_token_quota, duration,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    class.id AS class_id, course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    sa.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    slt.course_fee_value AS course_fee_value, slt.transport_fee_value AS transport_fee_value
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id

    LEFT JOIN student_attend AS sa ON presence.id = sa.presence_id
    LEFT JOIN user AS user_student ON sa.student_id = user_student.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE presence.id = ? LIMIT 1;

-- name: GetPresencesByClassId :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    class.id AS class_id, course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    sa.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    slt.course_fee_value AS course_fee_value, slt.transport_fee_value AS transport_fee_value
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id

    LEFT JOIN student_attend AS sa ON presence.id = sa.presence_id
    LEFT JOIN user AS user_student ON sa.student_id = user_student.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE class.id = ?
ORDER BY class.id;

-- name: GetPresencesByTeacherId :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    class.id AS class_id, course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    sa.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    slt.course_fee_value AS course_fee_value, slt.transport_fee_value AS transport_fee_value
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id

    LEFT JOIN student_attend AS sa ON presence.id = sa.presence_id
    LEFT JOIN user AS user_student ON sa.student_id = user_student.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE presence.teacher_id = ?
ORDER BY class.id;

-- name: GetPresences :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    class.id AS class_id, course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    sa.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    slt.course_fee_value AS course_fee_value, slt.transport_fee_value AS transport_fee_value
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id

    LEFT JOIN student_attend AS sa ON presence.id = sa.presence_id
    LEFT JOIN user AS user_student ON sa.student_id = user_student.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
ORDER BY class.id;

-- name: InsertPresence :execlastid
INSERT INTO presence (
    date, used_student_token_quota, duration, class_id, teacher_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: UpdatePresence :exec
UPDATE presence
SET date = ?, used_student_token_quota = ?, duration = ?, class_id = ?, teacher_id = ?, token_id = ?
WHERE id = ?;

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

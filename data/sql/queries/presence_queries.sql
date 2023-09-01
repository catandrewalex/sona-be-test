/* ============================== PRESENCE ============================== */
-- name: GetPresenceIdsOfSameClassAndDate :many
WITH ref_presence AS (
    SELECT *
    FROM presence
    WHERE presence.id = ?
)
SELECT presence.id AS id, presence.is_paid AS is_paid
FROM presence
    JOIN ref_presence ON presence.class_id = ref_presence.class_id AND presence.date = ref_presence.date
ORDER by presence.id;


-- name: GetPresenceById :one
SELECT presence.id AS presence_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE presence.id = ? LIMIT 1;

-- name: GetPresencesByIds :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE presence.id IN (sqlc.slice('ids'));

-- name: GetPresences :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration, note, is_paid,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE
    (presence.date >= sqlc.arg('startDate') AND presence.date <= sqlc.arg('endDate'))
    AND (class_id = sqlc.arg('class_id') OR sqlc.arg('use_class_filter') = false)
    AND (student_id = sqlc.arg('student_id') OR sqlc.arg('use_student_filter') = false)
ORDER BY class.id
LIMIT ? OFFSET ?;

-- name: CountPresences :one
SELECT Count(id) AS total FROM presence
WHERE
    (date >= sqlc.arg('startDate') AND date <= sqlc.arg('endDate'))
    AND (class_id = sqlc.arg('class_id') OR sqlc.arg('use_class_filter') = false)
    AND (student_id = sqlc.arg('student_id') OR sqlc.arg('use_student_filter') = false);

-- name: CountPresencesByIds :one
SELECT Count(id) AS total FROM presence
WHERE id IN (sqlc.slice('ids'));

-- name: InsertPresence :execlastid
INSERT INTO presence (
    date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdatePresence :exec
UPDATE presence
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, is_paid = ?, class_id = ?, teacher_id = ?, student_id = ?, token_id = ?
WHERE id = ?;

-- name: EditPresences :exec
UPDATE presence
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, teacher_id = ?
WHERE id IN (sqlc.slice('ids'));

-- name: DeletePresenceById :exec
DELETE FROM presence
WHERE id = ?;

-- name: DeletePresencesByIds :exec
DELETE FROM presence
WHERE id IN (sqlc.slice('ids'));

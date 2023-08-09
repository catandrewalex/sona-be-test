/* ============================== PRESENCE ============================== */
-- name: GetPresenceById :one
SELECT presence.id AS presence_id, date, used_student_token_quota, duration,
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
SELECT presence.id AS presence_id, date, used_student_token_quota, duration,
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

-- name: GetPresencesByClassId :many
WITH presence_paginated AS (
    SELECT * FROM presence
    WHERE presence.class_id = ?
    LIMIT ? OFFSET ?
)
SELECT presence_paginated.id AS presence_id, date, used_student_token_quota, duration,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    presence_paginated.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence_paginated.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM presence_paginated
    LEFT JOIN teacher ON presence_paginated.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence_paginated.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence_paginated.token_id = slt.id
WHERE presence.date >= sqlc.arg('startDate') AND presence.date <= sqlc.arg('endDate')
ORDER BY class.id;

-- name: GetPresencesByTeacherId :many
WITH presence_paginated AS (
    SELECT * FROM presence
    WHERE presence.teacher_id = ?
    LIMIT ? OFFSET ?
)
SELECT presence_paginated.id AS presence_id, date, used_student_token_quota, duration,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    presence_paginated.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence_paginated.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM presence_paginated
    LEFT JOIN teacher ON presence_paginated.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence_paginated.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence_paginated.token_id = slt.id
WHERE presence.date >= sqlc.arg('startDate') AND presence.date <= sqlc.arg('endDate')
ORDER BY class.id;

-- name: GetPresencesByStudentId :many
WITH presence_paginated AS (
    SELECT * FROM presence
    WHERE presence.student_id = ?
    LIMIT ? OFFSET ?
)
SELECT presence_paginated.id AS presence_id, date, used_student_token_quota, duration,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    presence_paginated.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence_paginated.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM presence_paginated
    LEFT JOIN teacher ON presence_paginated.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence_paginated.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence_paginated.token_id = slt.id
WHERE presence.date >= sqlc.arg('startDate') AND presence.date <= sqlc.arg('endDate')
ORDER BY class.id;

-- name: GetPresences :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration,
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
WHERE presence.date >= sqlc.arg('startDate') AND presence.date <= sqlc.arg('endDate')
ORDER BY class.id
LIMIT ? OFFSET ?;

-- name: CountPresencesByClassId :one
SELECT Count(id) AS total FROM presence
WHERE class_id = ? AND date >= ? AND date <= ?;

-- name: CountPresencesByTeacherId :one
SELECT Count(id) AS total FROM presence
WHERE teacher_id = ? AND date >= ? AND date <= ?;

-- name: CountPresencesByStudentId :one
SELECT Count(id) AS total FROM presence
WHERE student_id = ? AND date >= ? AND date <= ?;

-- name: CountPresences :one
SELECT Count(id) AS total FROM presence
WHERE date >= ? AND date <= ?;

-- name: CountPresencesByIds :one
SELECT Count(id) AS total FROM presence
WHERE id IN (sqlc.slice('ids'));

-- name: InsertPresence :execlastid
INSERT INTO presence (
    date, used_student_token_quota, duration, class_id, teacher_id, student_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdatePresence :exec
UPDATE presence
SET date = ?, used_student_token_quota = ?, duration = ?, class_id = ?, teacher_id = ?, student_id = ?, token_id = ?
WHERE id = ?;

-- name: DeletePresenceById :exec
DELETE FROM presence
WHERE id = ?;

-- name: DeletePresencesByIds :exec
DELETE FROM presence
WHERE id IN (sqlc.slice('ids'));

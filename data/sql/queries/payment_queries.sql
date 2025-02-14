/* ============================== ENROLLMENT_PAYMENT ============================== */
-- name: GetEnrollmentPaymentById :one
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, balance_bonus, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE ep.id = ? LIMIT 1;

-- name: GetEnrollmentPaymentsByIds :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, balance_bonus, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE ep.id IN (sqlc.slice('ids'));

-- name: GetEnrollmentPayments :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, balance_bonus, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE
    ep.payment_date >= sqlc.arg('startDate') AND ep.payment_date <= sqlc.arg('endDate')
ORDER BY ep.id
LIMIT ? OFFSET ?;

-- name: GetLatestEnrollmentPaymentDateByStudentEnrollmentId :one
SELECT MAX(payment_date) AS last_payment_date
FROM enrollment_payment
WHERE enrollment_id = ?
GROUP BY enrollment_id LIMIT 1;

-- name: GetEnrollmentPaymentsDescendingDate :many
-- GetEnrollmentPaymentsDescendingDate is a copy of GetEnrollmentPayments, with additional sort by date parameter. TODO: find alternative: sqlc's dynamic query which is mature enough, so that we need to do this.
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, balance_bonus, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE
    ep.payment_date >= sqlc.arg('startDate') AND ep.payment_date <= sqlc.arg('endDate')
ORDER BY ep.payment_date DESC, ep.id DESC
LIMIT ? OFFSET ?;

-- name: CountEnrollmentPaymentsByIds :one
SELECT Count(id) AS total FROM enrollment_payment
WHERE id IN (sqlc.slice('ids'));

-- name: CountEnrollmentPayments :one
SELECT Count(id) AS total FROM enrollment_payment;

-- name: InsertEnrollmentPayment :execlastid
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, balance_bonus, course_fee_value, transport_fee_value, penalty_fee_value, discount_fee_value, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateEnrollmentPayment :exec
UPDATE enrollment_payment SET payment_date = ?, balance_top_up = ?, balance_bonus = ?, course_fee_value = ?, transport_fee_value = ?, penalty_fee_value = ?, discount_fee_value = ?
WHERE id = ?;

-- name: UpdateEnrollmentPaymentOnSafeAttributes :exec
UPDATE enrollment_payment SET payment_date = ?, balance_bonus = ?, discount_fee_value = ?
WHERE id = ?;

-- name: DeleteEnrollmentPaymentById :exec
DELETE FROM enrollment_payment
WHERE id = ?;

-- name: DeleteEnrollmentPaymentsByIds :exec
DELETE FROM enrollment_payment
WHERE id IN (sqlc.slice('ids'));

/* ============================== STUDENT_LEARNING_TOKEN ============================== */
-- name: GetSLTByEnrollmentIdAndCourseFeeQuarterAndTransportFeeQuarter :one
SELECT * FROM student_learning_token
WHERE enrollment_id = ? AND course_fee_quarter_value = ? AND transport_fee_quarter_value = ?;

-- name: GetEarliestAvailableSLTsByStudentEnrollmentIds :many
WITH slt_min_max AS (
    -- fetch earliest SLT with quota > 0
    SELECT enrollment_id, MIN(last_updated_at) AS updateDateWithNonZeroQuota_or_maxUpdateDate
    FROM student_learning_token
    WHERE quota > 0
    GROUP BY enrollment_id
    UNION
    -- combined with latest SLT, to cover case when all SLT has <= 0 quota
    SELECT enrollment_id, MAX(last_updated_at) AS updateDateWithNonZeroQuota_or_maxUpdateDate
    FROM student_learning_token
    GROUP BY enrollment_id
    -- each record will be unique if all non-latest SLTs has 0 quota; OR duplicated (2 records) if there exists non-latest SLT with quota > 0
)
SELECT slt.id AS student_learning_token_id, quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, slt.enrollment_id AS enrollment_id,
    se.student_id AS student_id
FROM student_learning_token AS slt
    JOIN (
        -- we have 1-2 SLT option per enrollment_id from `slt_min_max`, pick the earliest
        SELECT enrollment_id, MIN(updateDateWithNonZeroQuota_or_maxUpdateDate) AS earliestUpdateDateWithNonZeroQuota
        FROM slt_min_max
        GROUP BY enrollment_id
    ) AS slt_min ON (
        slt.enrollment_id = slt_min.enrollment_id
        AND last_updated_at = earliestUpdateDateWithNonZeroQuota
    )

    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
WHERE slt.enrollment_id IN (sqlc.slice('student_enrollment_ids'));

-- name: IncrementSLTQuotaById :exec
UPDATE student_learning_token SET quota = ROUND(quota + ?, 3), last_updated_at = ?
WHERE id = ?;

-- name: GetSLTByClassIdForAttendanceInfo :many
SELECT slt.id AS student_learning_token_id, quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, se.student_id AS student_id
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
WHERE se.class_id = ?
ORDER BY last_updated_at DESC, slt.id DESC;

-- name: GetStudentLearningTokenById :one
SELECT slt.id AS student_learning_token_id, quota, course_fee_quarter_value, transport_fee_quarter_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE slt.id = ? LIMIT 1;

-- name: GetStudentLearningTokensByIds :many
SELECT slt.id AS student_learning_token_id, quota, course_fee_quarter_value, transport_fee_quarter_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE slt.id IN (sqlc.slice('ids'));

-- name: GetStudentLearningTokensByEnrollmentId :many
SELECT slt.id AS student_learning_token_id, quota, course_fee_quarter_value, transport_fee_quarter_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE slt.enrollment_id = ?;

-- name: GetStudentLearningTokens :many
SELECT slt.id AS student_learning_token_id, quota, course_fee_quarter_value, transport_fee_quarter_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
ORDER BY slt.id
LIMIT ? OFFSET ?;

-- name: CountStudentLearningTokensByIds :one
SELECT Count(id) AS total FROM student_learning_token
WHERE id IN (sqlc.slice('ids'));

-- name: CountStudentLearningTokens :one
SELECT Count(id) AS total FROM student_learning_token;

-- name: InsertStudentLearningToken :execlastid
INSERT INTO student_learning_token (
    quota, course_fee_quarter_value, transport_fee_quarter_value, created_at, last_updated_at, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: UpdateStudentLearningToken :exec
UPDATE student_learning_token SET quota = ?, course_fee_quarter_value = ?, transport_fee_quarter_value = ?, last_updated_at = ?
WHERE id = ?;

-- name: DeleteStudentLearningTokenById :exec
DELETE FROM student_learning_token
WHERE id = ?;

-- name: DeleteStudentLearningTokensByIds :exec
DELETE FROM student_learning_token
WHERE id IN (sqlc.slice('ids'));

/* ============================== TEACHER_PAYMENT ============================== */
-- name: GetTeacherPaymentAttendanceIdsByIds :many
SELECT attendance_id AS id FROM teacher_payment
WHERE teacher_payment.id IN (sqlc.slice('teacher_payment_ids'));

-- name: GetTeacherPaymentsByTeacherId :many
SELECT tp.id AS teacher_payment_id, paid_course_fee_value, paid_transport_fee_value, added_at,
    sqlc.embed(attendance),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM teacher_payment AS tp
    JOIN attendance ON attendance_id = attendance.id
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class ON attendance.class_id = class.id
    LEFT JOIN course ON class.course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
    
    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (tp.added_at >= sqlc.arg('startDate') AND tp.added_at <= sqlc.arg('endDate'))
    AND attendance.teacher_id = sqlc.arg('teacher_id')
ORDER BY attendance.date DESC, tp.id ASC;

-- name: GetTeacherPaymentById :one
SELECT tp.id AS teacher_payment_id, paid_course_fee_value, paid_transport_fee_value, added_at,
    sqlc.embed(attendance),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM teacher_payment AS tp
    JOIN attendance ON attendance_id = attendance.id
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class ON attendance.class_id = class.id
    LEFT JOIN course ON class.course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
    
    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE tp.id = ? LIMIT 1;

-- name: GetTeacherPaymentsByIds :many
SELECT tp.id AS teacher_payment_id, paid_course_fee_value, paid_transport_fee_value, added_at,
    sqlc.embed(attendance),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM teacher_payment AS tp
    JOIN attendance ON attendance_id = attendance.id
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class ON attendance.class_id = class.id
    LEFT JOIN course ON class.course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
    
    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE tp.id IN (sqlc.slice('ids'));

-- name: GetTeacherPayments :many
SELECT tp.id AS teacher_payment_id, paid_course_fee_value, paid_transport_fee_value, added_at,
    sqlc.embed(attendance),
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    sqlc.embed(slt)
FROM teacher_payment AS tp
    JOIN attendance ON attendance_id = attendance.id
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class ON attendance.class_id = class.id
    LEFT JOIN course ON class.course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
    
    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (tp.added_at >= sqlc.arg('startDate') AND tp.added_at <= sqlc.arg('endDate'))
    AND (attendance.teacher_id = sqlc.arg('teacher_id') OR sqlc.arg('use_teacher_filter') = false)
ORDER BY tp.id
LIMIT ? OFFSET ?;

-- name: CountTeacherPaymentsByIds :one
SELECT Count(id) AS total FROM teacher_payment
WHERE id IN (sqlc.slice('ids'));

-- name: CountTeacherPayments :one
SELECT Count(teacher_payment.id) AS total
FROM teacher_payment
    JOIN attendance ON attendance_id = attendance.id
WHERE
    (attendance.teacher_id = sqlc.arg('teacher_id') OR sqlc.arg('use_teacher_filter') = false);

-- name: InsertTeacherPayment :execlastid
INSERT INTO teacher_payment (
    attendance_id, paid_course_fee_value, paid_transport_fee_value
) VALUES (
    ?, ?, ?
);

-- name: UpdateTeacherPayment :exec
UPDATE teacher_payment SET attendance_id = ?, paid_course_fee_value = ?, paid_transport_fee_value = ?, added_at = ?
WHERE id = ?;

-- name: EditTeacherPayment :exec
UPDATE teacher_payment SET paid_course_fee_value = ?, paid_transport_fee_value = ?
WHERE id = ?;

-- name: DeleteTeacherPaymentById :exec
DELETE FROM teacher_payment
WHERE id = ?;

-- name: DeleteTeacherPaymentsByIds :exec
DELETE FROM teacher_payment
WHERE id IN (sqlc.slice('ids'));

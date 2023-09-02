/* ============================== ENROLLMENT_PAYMENT ============================== */
-- name: GetEnrollmentPaymentById :one
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
WHERE ep.id = ? LIMIT 1;

-- name: GetEnrollmentPaymentsByIds :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
WHERE ep.id IN (sqlc.slice('ids'));

-- name: GetEnrollmentPayments :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
WHERE
    ep.payment_date >= sqlc.arg('startDate') AND ep.payment_date <= sqlc.arg('endDate')
ORDER BY ep.id
LIMIT ? OFFSET ?;

-- name: GetEnrollmentPaymentsDescendingDate :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
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
    payment_date, balance_top_up, course_fee_value, transport_fee_value, penalty_fee_value, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: UpdateEnrollmentPayment :exec
UPDATE enrollment_payment SET payment_date = ?, balance_top_up = ?, course_fee_value = ?, transport_fee_value = ?, penalty_fee_value = ?
WHERE id = ?;

-- name: UpdateEnrollmentPaymentDateAndBalance :exec
UPDATE enrollment_payment SET payment_date = ?, balance_top_up = ?
WHERE id = ?;

-- name: DeleteEnrollmentPaymentById :exec
DELETE FROM enrollment_payment
WHERE id = ?;

-- name: DeleteEnrollmentPaymentsByIds :exec
DELETE FROM enrollment_payment
WHERE id IN (sqlc.slice('ids'));

/* ============================== STUDENT_LEARNING_TOKEN ============================== */
-- name: GetSLTWithNegativeQuotaByEnrollmentId :many
SELECT * FROM student_learning_token
WHERE enrollment_id = ? AND quota < 0;

-- name: GetSLTByEnrollmentIdAndCourseFeeAndTransportFee :one
SELECT * FROM student_learning_token
WHERE enrollment_id = ? AND course_fee_value = ? AND transport_fee_value = ?;

-- name: GetEarliestPenaltyDateSLTByStudentEnrollmentId :one
SELECT MIN(penalty_start_at) AS penalty_date FROM student_learning_token
WHERE enrollment_id = ?
GROUP BY enrollment_id;

-- name: GetEarliestAvailableSLTsByStudentEnrollmentIds :many
WITH slt_min_max AS (
    -- fetch earliest SLT with quota > 0
    SELECT enrollment_id, MIN(created_at) AS createDateWithNonZeroQuota_or_maxCreateDate
    FROM student_learning_token
    WHERE quota > 0
    GROUP BY enrollment_id
    UNION
    -- combined with latest SLT, to cover case when all SLT has <= 0 quota
    SELECT enrollment_id, MAX(created_at) AS createDateWithNonZeroQuota_or_maxCreateDate
    FROM student_learning_token
    GROUP BY enrollment_id
    -- each record will be unique if all non-latest SLTs has 0 quota; OR duplicated (2 records) if there exists non-latest SLT with quota > 0
)
SELECT slt.id AS student_learning_token_id, quota, course_fee_value, transport_fee_value, created_at, last_updated_at, slt.enrollment_id AS enrollment_id,
    -- we have 2 SLT option, pick the earliest
    se.student_id AS student_id, MIN(createDateWithNonZeroQuota_or_maxCreateDate)
FROM student_learning_token AS slt
    JOIN slt_min_max ON (
        slt.enrollment_id = slt_min_max.enrollment_id
        AND created_at = createDateWithNonZeroQuota_or_maxCreateDate
    )

    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
WHERE slt.enrollment_id IN (sqlc.slice('student_enrollment_ids'))
GROUP BY slt.enrollment_id;

-- name: IncrementSLTQuotaById :exec
UPDATE student_learning_token SET quota = quota + ?
WHERE id = ?;

-- name: GetStudentLearningTokenById :one
SELECT slt.id AS student_learning_token_id, quota, course_fee_value, transport_fee_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
WHERE slt.id = ? LIMIT 1;

-- name: GetStudentLearningTokensByIds :many
SELECT slt.id AS student_learning_token_id, quota, course_fee_value, transport_fee_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
WHERE slt.id IN (sqlc.slice('ids'));

-- name: GetStudentLearningTokensByEnrollmentId :many
SELECT slt.id AS student_learning_token_id, quota, course_fee_value, transport_fee_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
WHERE slt.enrollment_id = ?;

-- name: GetStudentLearningTokens :many
SELECT slt.id AS student_learning_token_id, quota, course_fee_value, transport_fee_value, slt.created_at, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id
    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
ORDER BY slt.id
LIMIT ? OFFSET ?;

-- name: CountStudentLearningTokensByIds :one
SELECT Count(id) AS total FROM student_learning_token
WHERE id IN (sqlc.slice('ids'));

-- name: CountStudentLearningTokens :one
SELECT Count(id) AS total FROM student_learning_token;

-- name: InsertStudentLearningToken :execlastid
INSERT INTO student_learning_token (
    quota, course_fee_value, transport_fee_value, enrollment_id
) VALUES (
    ?, ?, ?, ?
);

-- name: UpdateStudentLearningToken :exec
UPDATE student_learning_token SET quota = ?, course_fee_value = ?, transport_fee_value = ?
WHERE id = ?;

-- name: ResetStudentLearningTokenQuotaByIds :exec
UPDATE student_learning_token SET quota = 0
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteStudentLearningTokenById :exec
DELETE FROM student_learning_token
WHERE id = ?;

-- name: DeleteStudentLearningTokensByIds :exec
DELETE FROM student_learning_token
WHERE id IN (sqlc.slice('ids'));

/* ============================== TEACHER_SALARY ============================== */
-- name: GetTeacherSalaryById :one
SELECT * FROM teacher_salary
WHERE id = ? LIMIT 1;

-- name: GetTeacherSalaries :many
SELECT ts.id AS teacher_salary_id, profit_sharing_percentage, added_at,
    presence.id AS presence_id, date, used_student_token_quota, duration,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM teacher_salary AS ts
    JOIN presence ON presence_id = presence.id
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class ON presence.class_id = class.id
    LEFT JOIN course ON class.course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
ORDER BY ts.id;

-- name: InsertTeacherSalary :execlastid
INSERT INTO teacher_salary (
    presence_id, profit_sharing_percentage, added_at
) VALUES (
    ?, ?, ?
);

-- name: DeleteTeacherSalaryById :exec
DELETE FROM teacher_salary
WHERE id = ?;

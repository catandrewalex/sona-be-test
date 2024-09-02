-- All CAST(.. AS SIGNED) in this file (around aggregations: SUM()) are to force SQLC to generate the field type as int64. Else, they will be interface{}.
-- To make things even worse: as MySQL's "SUM() of INT" type is decimal(32,0), Go's 'row.Scan(&fieldName)' will read the value as string (i.e. []uint8) instead of int!

/* ============================== EXPENSE ============================== */
-- name: GetExpenseOverview :many
SELECT DATE_FORMAT(tp.added_at, '%Y-%m') AS year_with_month, CAST(sum(tp.paid_course_fee_value) AS SIGNED) AS total_paid_course_fee, CAST(sum(tp.paid_transport_fee_value) AS SIGNED) AS total_paid_transport_fee
FROM teacher_payment AS tp
    -- we need this joins just for the filtering (teacher_id & instrument_id)
    JOIN attendance ON tp.attendance_id = attendance.id
    JOIN class ON attendance.class_id = class.id
    JOIN course ON class.course_id = course.id
WHERE
    (tp.added_at >= sqlc.arg('startDate') AND tp.added_at <= sqlc.arg('endDate'))
    AND (class.teacher_id IN (sqlc.slice('teacher_ids')) OR sqlc.arg('use_teacher_filter') = false)
    AND (course.instrument_id IN (sqlc.slice('instrument_ids')) OR sqlc.arg('use_instrument_filter') = false)
GROUP BY year_with_month
ORDER BY year_with_month ASC;

-- name: GetExpenseMonthlySummaryGroupedByTeacher :many
SELECT teacher.id AS teacher_id, user.id AS user_id, user_detail, CAST(sum(tp.paid_course_fee_value) AS SIGNED) AS total_paid_course_fee, CAST(sum(tp.paid_transport_fee_value) AS SIGNED) AS total_paid_transport_fee
FROM teacher_payment AS tp
    JOIN attendance ON tp.attendance_id = attendance.id
    JOIN teacher ON attendance.teacher_id = teacher.id
    JOIN user ON teacher.user_id = user.id
    
    -- we need this joins just for the filtering (instrument_id)
    JOIN class ON attendance.class_id = class.id
    JOIN course ON class.course_id = course.id
WHERE
    (tp.added_at >= sqlc.arg('startDate') AND tp.added_at <= sqlc.arg('endDate'))
    AND (class.teacher_id IN (sqlc.slice('teacher_ids')) OR sqlc.arg('use_teacher_filter') = false)
    AND (course.instrument_id IN (sqlc.slice('instrument_ids')) OR sqlc.arg('use_instrument_filter') = false)
GROUP BY teacher.id
ORDER BY total_paid_course_fee;

-- name: GetExpenseMonthlySummaryGroupedByInstrument :many
SELECT sqlc.embed(instrument), CAST(sum(tp.paid_course_fee_value) AS SIGNED) AS total_paid_course_fee, CAST(sum(tp.paid_transport_fee_value) AS SIGNED) AS total_paid_transport_fee
FROM teacher_payment AS tp
    JOIN attendance ON tp.attendance_id = attendance.id
    JOIN class ON attendance.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON instrument_id = instrument.id
WHERE
    (tp.added_at >= sqlc.arg('startDate') AND tp.added_at <= sqlc.arg('endDate'))
    AND (class.teacher_id IN (sqlc.slice('teacher_ids')) OR sqlc.arg('use_teacher_filter') = false)
    AND (course.instrument_id IN (sqlc.slice('instrument_ids')) OR sqlc.arg('use_instrument_filter') = false)
GROUP BY instrument.id
ORDER BY total_paid_course_fee;

/* ============================== INCOME ============================== */
-- name: GetIncomeOverview :many
SELECT DATE_FORMAT(ep.payment_date, '%Y-%m') AS year_with_month, CAST(sum(ep.course_fee_value) AS SIGNED) AS total_course_fee, CAST(sum(ep.transport_fee_value) AS SIGNED) AS total_transport_fee, CAST(sum(ep.penalty_fee_value) AS SIGNED) AS total_penalty_fee_value, CAST(sum(ep.discount_fee_value) AS SIGNED) AS total_discount_fee_value
FROM enrollment_payment AS ep
    -- we need this joins just for the filtering (student_id & instrument_id)
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id
    JOIN class ON se.class_id = class.id
    JOIN course ON class.course_id = course.id
WHERE
    (ep.payment_date >= sqlc.arg('startDate') AND ep.payment_date <= sqlc.arg('endDate'))
    AND (se.student_id IN (sqlc.slice('student_ids')) OR sqlc.arg('use_student_filter') = false)
    AND (course.instrument_id IN (sqlc.slice('instrument_ids')) OR sqlc.arg('use_instrument_filter') = false)
GROUP BY year_with_month
ORDER BY year_with_month ASC;

-- name: GetIncomeMonthlySummaryGroupedByStudent :many
SELECT student.id AS student_id, user.id AS user_id, user_detail, CAST(sum(ep.course_fee_value) AS SIGNED) AS total_course_fee, CAST(sum(ep.transport_fee_value) AS SIGNED) AS total_transport_fee, CAST(sum(ep.penalty_fee_value) AS SIGNED) AS total_penalty_fee_value, CAST(sum(ep.discount_fee_value) AS SIGNED) AS total_discount_fee_value
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id
    JOIN student ON se.student_id = student.id
    JOIN user ON student.user_id = user.id
    
    -- we need this joins just for the filtering (student_id & instrument_id)
    JOIN class ON se.class_id = class.id
    JOIN course ON class.course_id = course.id
WHERE
    (ep.payment_date >= sqlc.arg('startDate') AND ep.payment_date <= sqlc.arg('endDate'))
    AND (se.student_id IN (sqlc.slice('student_ids')) OR sqlc.arg('use_student_filter') = false)
    AND (course.instrument_id IN (sqlc.slice('instrument_ids')) OR sqlc.arg('use_instrument_filter') = false)
GROUP BY student.id
ORDER BY total_course_fee;

-- name: GetIncomeMonthlySummaryGroupedByInstrument :many
SELECT sqlc.embed(instrument), CAST(sum(ep.course_fee_value) AS SIGNED) AS total_course_fee, CAST(sum(ep.transport_fee_value) AS SIGNED) AS total_transport_fee, CAST(sum(ep.penalty_fee_value) AS SIGNED) AS total_penalty_fee_value, CAST(sum(ep.discount_fee_value) AS SIGNED) AS total_discount_fee_value
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id
    JOIN class ON se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON instrument_id = instrument.id
WHERE
    (ep.payment_date >= sqlc.arg('startDate') AND ep.payment_date <= sqlc.arg('endDate'))
    AND (se.student_id IN (sqlc.slice('student_ids')) OR sqlc.arg('use_student_filter') = false)
    AND (course.instrument_id IN (sqlc.slice('instrument_ids')) OR sqlc.arg('use_instrument_filter') = false)
GROUP BY instrument.id
ORDER BY total_course_fee;

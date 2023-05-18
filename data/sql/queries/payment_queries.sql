/* ============================== ENROLLMENT_PAYMENT ============================== */
-- name: GetEnrollmentPaymentById :one
SELECT * FROM enrollment_payment
WHERE id = ? LIMIT 1;

-- name: GetEnrollmentPayments :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, value, value_penalty,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id AS class_id, class.course_id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
ORDER BY ep.id;

-- name: InsertEnrollmentPayment :execlastid
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, value, value_penalty, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: DeleteEnrollmentPaymentById :exec
DELETE FROM enrollment_payment
WHERE id = ?;

/* ============================== STUDENT_LEARNING_TOKEN ============================== */
-- name: GetStudentLearningTokenById :one
SELECT * FROM student_learning_token
WHERE id = ? LIMIT 1;

-- name: GetStudentLearningTokensByEnrollmentId :many
SELECT * FROM student_learning_token
WHERE enrollment_id = ?;

-- name: GetStudentLearningTokens :many
SELECT slt.id AS student_learning_token_id, quota, quota_bonus, course_fee_value, transport_fee_value, last_updated_at,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id AS class_id, class.course_id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
ORDER BY slt.id;

-- name: InsertStudentLearningToken :execlastid
INSERT INTO student_learning_token (
    quota, quota_bonus, course_fee_value, transport_fee_value, last_updated_at, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: DeleteStudentLearningTokenById :exec
DELETE FROM student_learning_token
WHERE id = ?;

/* ============================== TEACHER_SALARY ============================== */
-- name: GetTeacherSalaryById :one
SELECT * FROM teacher_salary
WHERE id = ? LIMIT 1;

-- name: GetTeacherSalaries :many
SELECT ts.id AS teacher_salary_id, profit_sharing_percentage, added_at,
    presence.id AS presence_id, date, used_student_token_quota, duration,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    class.id AS class_id, course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    sa.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail
FROM teacher_salary AS ts
    JOIN presence ON presence_id = presence.id
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON class.course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id

    LEFT JOIN student_attend AS sa ON presence.id = sa.presence_id
    LEFT JOIN user AS user_student ON sa.student_id = user_student.id
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

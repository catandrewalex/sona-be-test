/* ============================== ENROLLMENT_PAYMENT ============================== */
-- name: GetEnrollmentPaymentById :one
SELECT * FROM enrollment_payment
WHERE id = ? LIMIT 1;

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

-- name: InsertTeacherSalary :execlastid
INSERT INTO teacher_salary (
    presence_id, course_fee_value, transport_fee_value, profit_sharing_percentage, added_at
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: DeleteTeacherSalaryById :exec
DELETE FROM teacher_salary
WHERE id = ?;

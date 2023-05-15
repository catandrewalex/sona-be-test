/* ============================== INSTRUMENT ============================== */
-- name: GetInstrumentById :one
SELECT * FROM instrument
WHERE id = ? LIMIT 1;

-- name: InsertInstrument :execlastid
INSERT INTO instrument (
    id, name
) VALUES (
    ?, ?
);

-- name: DeleteInstrumentById :exec
DELETE FROM instrument
WHERE id = ?;

/* ============================== GRADE ============================== */
-- name: GetGradeById :one
SELECT * FROM grade
WHERE id = ? LIMIT 1;

-- name: InsertGrade :execlastid
INSERT INTO grade (
    id, name
) VALUES (
    ?, ?
);

-- name: DeleteGradeById :exec
DELETE FROM grade
WHERE id = ?;

/* ============================== COURSE ============================== */
-- name: GetCourseById :one
SELECT * FROM course
WHERE id = ? LIMIT 1;

-- name: InsertCourse :execlastid
INSERT INTO course (
    id, default_fee, instrument_id, grade_id
) VALUES (
    ?, ?, ?, ?
);

-- name: DeleteCourseById :exec
DELETE FROM course
WHERE id = ?;

/* ============================== CLASS ============================== */
-- name: GetClassById :one
SELECT * FROM class
WHERE id = ? LIMIT 1;

-- name: InsertClass :execlastid
INSERT INTO class (
    id, default_transport_fee, teacher_id, course_id, is_deactivated
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: DeleteClassById :exec
DELETE FROM class
WHERE id = ?;

/* ============================== STUDENT_ENROLLMENT ============================== */
-- name: GetStudentEnrollmentsByStudentId :many
SELECT * FROM student_enrollment
WHERE student_id = ?;

-- name: GetStudentEnrollmentsByClassId :many
SELECT * FROM student_enrollment
WHERE class_id = ?;

-- name: InsertStudentEnrollment :exec
INSERT INTO student_enrollment (
    id, student_id, class_id
) VALUES (
    ?, ?, ?
);

-- name: DeleteStudentEnrollmentById :exec
DELETE FROM student_enrollment
WHERE id = ?;

-- name: DeleteStudentEnrollmentByStudentId :exec
DELETE FROM student_enrollment
WHERE student_id = ?;

-- name: DeleteStudentEnrollmentByClassId :exec
DELETE FROM student_enrollment
WHERE class_id = ?;

/* ============================== ENROLLMENT_PAYMENT ============================== */
-- name: GetEnrollmentPaymentById :one
SELECT * FROM enrollment_payment
WHERE id = ? LIMIT 1;

-- name: InsertEnrollmentPayment :execlastid
INSERT INTO enrollment_payment (
    id, payment_date, balance_top_up, value, value_penalty, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?, ?
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
    id, quota, quota_bonus, course_fee_value, transport_fee_value, last_updated_at, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
);

-- name: DeleteStudentLearningTokenById :exec
DELETE FROM student_learning_token
WHERE id = ?;

/* ============================== TEACHER_SPECIAL_FEE ============================== */
-- name: GetTeacherSpecialFeeById :one
SELECT * FROM teacher_special_fee
WHERE id = ? LIMIT 1;

-- name: GetTeacherSpecialFeesByTeacherId :many
SELECT * FROM teacher_special_fee
WHERE teacher_id = ?;

-- name: InsertTeacherSpecialFee :execlastid
INSERT INTO teacher_special_fee (
    id, fee, teacher_id, course_id
) VALUES (
    ?, ?, ?, ?
);

-- name: DeleteTeacherSpecialFeeById :exec
DELETE FROM teacher_special_fee
WHERE id = ?;

-- name: DeleteTeacherSpecialFeeByTeacherId :exec
DELETE FROM teacher_special_fee
WHERE teacher_id = ?;

/* ============================== PRESENCE ============================== */
-- name: GetPresenceById :one
SELECT * FROM presence
WHERE id = ? LIMIT 1;

-- name: GetPresencesByClassId :many
SELECT * FROM presence
WHERE class_id = ?;

-- name: GetPresencesByTeacherId :many
SELECT * FROM presence
WHERE teacher_id = ?;

-- name: InsertPresence :execlastid
INSERT INTO presence (
    id, date, used_student_token_quota, duration, class_id, teacher_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
);

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

/* ============================== TEACHER_SALARY ============================== */
-- name: GetTeacherSalaryById :one
SELECT * FROM teacher_salary
WHERE id = ? LIMIT 1;

-- name: InsertTeacherSalary :execlastid
INSERT INTO teacher_salary (
    id, presence_id, course_fee_value, transport_fee_value, profit_sharing_percentage, added_at
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: DeleteTeacherSalaryById :exec
DELETE FROM teacher_salary
WHERE id = ?;

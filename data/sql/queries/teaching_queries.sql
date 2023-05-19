/* ============================== TEACHER ============================== */
-- name: GetTeacherById :one
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
WHERE teacher.id = ? LIMIT 1;

-- name: GetTeacherByUserId :one
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
WHERE user_id = ? LIMIT 1;

-- name: GetTeachers :many
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
ORDER BY username
LIMIT ? OFFSET ?;

-- name: CountTeachers :one
SELECT Count(*) as total FROM teacher;

-- name: InsertTeacher :execlastid
INSERT INTO teacher ( user_id ) VALUES ( ? );

-- name: DeleteTeacherById :exec
DELETE FROM teacher
WHERE id = ?;

-- name: DeleteTeacherByUserId :exec
DELETE FROM teacher
WHERE user_id = ?;

/* ============================== STUDENT ============================== */
-- name: GetStudentById :one
SELECT student.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM student JOIN user ON student.user_id = user.id
WHERE student.id = ? LIMIT 1;

-- name: GetStudentByUserId :one
SELECT student.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM student JOIN user ON student.user_id = user.id
WHERE user_id = ? LIMIT 1;

-- name: GetStudents :many
SELECT student.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM student JOIN user ON student.user_id = user.id
ORDER BY username
LIMIT ? OFFSET ?;

-- name: CountStudents :one
SELECT Count(*) as total FROM student;

-- name: InsertStudent :execlastid
INSERT INTO student ( user_id ) VALUES ( ? );

-- name: DeleteStudentById :exec
DELETE FROM student
WHERE id = ?;

-- name: DeleteStudentByUserId :exec
DELETE FROM student
WHERE user_id = ?;

/* ============================== INSTRUMENT ============================== */
-- name: GetInstrumentById :one
SELECT * FROM instrument
WHERE id = ? LIMIT 1;

-- name: InsertInstrument :execlastid
INSERT INTO instrument ( name ) VALUES ( ? );

-- name: UpdateInstrument :exec
UPDATE instrument SET name = ?
WHERE id = ?;

-- name: DeleteInstrumentById :exec
DELETE FROM instrument
WHERE id = ?;

/* ============================== GRADE ============================== */
-- name: GetGradeById :one
SELECT * FROM grade
WHERE id = ? LIMIT 1;

-- name: InsertGrade :execlastid
INSERT INTO grade ( name ) VALUES ( ? );

-- name: UpdateGrade :exec
UPDATE grade SET name = ?
WHERE id = ?;

-- name: DeleteGradeById :exec
DELETE FROM grade
WHERE id = ?;

/* ============================== COURSE ============================== */
-- name: GetCourses :many
SELECT course.id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name, default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
ORDER BY course.id;

-- name: GetCoursesByInstrumentId :many
SELECT course.id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name, default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
WHERE instrument.id = ?
ORDER BY course.id;

-- name: GetCoursesByGradeId :many
SELECT course.id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name, default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
WHERE grade.id = ?
ORDER BY course.id;

-- name: GetCourseById :one
SELECT course.id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name, default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
WHERE course.id = ? LIMIT 1;

-- name: InsertCourse :execlastid
INSERT INTO course (
    default_fee, default_duration_minute, instrument_id, grade_id
) VALUES (
    ?, ?, ?, ?
);

-- name: UpdateCourseInfo :exec
UPDATE course SET default_fee = ?, default_duration_minute = ?
WHERE id = ?;

-- name: UpdateCourseInstrument :exec
UPDATE course SET instrument_id = ?
WHERE id = ?;

-- name: UpdateCourseGrade :exec
UPDATE course SET grade_id = ?
WHERE id = ?;

-- name: DeleteCourseById :exec
DELETE FROM course
WHERE id = ?;

/* ============================== CLASS ============================== */
-- name: GetClasses :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, teacher_id, se.student_id AS student_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON class.id = se.class_id
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
ORDER BY class.id
LIMIT ? OFFSET ?;

-- name: GetClassesByTeacherId :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, se.student_id AS student_id,
    CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON class.id = se.class_id
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
WHERE teacher_id = ?
ORDER BY class.id;

-- name: GetClassesByStudentId :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, teacher_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON class.id = se.class_id
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
WHERE se.student_id = ?
ORDER BY class.id;

-- name: GetClassById :one
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, teacher_id, se.student_id AS student_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    CONCAT_WS(' ', instrument.name, grade.name) AS course_name,
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON class.id = se.class_id
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
WHERE class.id = ? LIMIT 1;

-- name: InsertClass :execlastid
INSERT INTO class (
    transport_fee, teacher_id, course_id, is_deactivated
) VALUES (
    ?, ?, ?, ?
);

-- name: UpdateClassInfo :exec
UPDATE class SET transport_fee = ?
WHERE id = ?;

-- name: UpdateClassTeacher :exec
UPDATE class SET teacher_id = ?
WHERE id = ?;

-- name: UpdateClassCourse :exec
UPDATE class SET course_id = ?
WHERE id = ?;

-- name: ActivateClass :exec
UPDATE class SET is_deactivated = 1
WHERE id = ?;

-- name: DeactivateClass :exec
UPDATE class SET is_deactivated = 0
WHERE id = ?;

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

-- name: GetStudentEnrollments :many
SELECT se.id as student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id AS class_id, class.transport_fee AS class_transport_fee, course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name, course.default_fee AS course_default_fee
FROM student_enrollment as se
    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
ORDER BY se.id;

-- name: InsertStudentEnrollment :exec
INSERT INTO student_enrollment (
    student_id, class_id
) VALUES (
    ?, ?
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

/* ============================== TEACHER_SPECIAL_FEE ============================== */
-- name: GetTeacherSpecialFeeById :one
SELECT teacher_special_fee.id AS teacher_special_fee_id, fee,
    teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    course.id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name, default_fee AS original_course_fee
FROM teacher_special_fee
    JOIN teacher ON teacher_id = teacher.id
    JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    JOIN course on course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE teacher_special_fee.id = ? LIMIT 1;

-- name: GetTeacherSpecialFeesByTeacherId :many
SELECT teacher_special_fee.id AS teacher_special_fee_id, fee,
    teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    course.id AS course_id, CONCAT_WS(' ', instrument.name, grade.name) AS course_name, default_fee AS original_course_fee
FROM teacher_special_fee
    JOIN teacher ON teacher_id = teacher.id
    JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    JOIN course on course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE teacher_id = ?
ORDER BY course.id;

-- name: GetTeacherSpecialFeesByTeacherIdAndCourseId :one
SELECT id, fee FROM teacher_special_fee
WHERE teacher_id = ? AND course_id = ? LIMIT 1;

-- name: InsertTeacherSpecialFee :execlastid
INSERT INTO teacher_special_fee (
    fee, teacher_id, course_id
) VALUES (
    ?, ?, ?
);

-- name: UpdateTeacherSpecialFee :exec
UPDATE teacher_special_fee SET fee = ?
WHERE teacher_id = ? AND course_id = ?;

-- name: DeleteTeacherSpecialFeeById :exec
DELETE FROM teacher_special_fee
WHERE id = ?;

-- name: DeleteTeacherSpecialFeeByTeacherId :exec
DELETE FROM teacher_special_fee
WHERE teacher_id = ?;

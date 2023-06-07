/* ============================== TEACHER ============================== */
-- name: GetTeacherById :one
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
WHERE teacher.id = ? LIMIT 1;

-- name: GetTeacherByUserId :one
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
WHERE user_id = ? LIMIT 1;

-- name: GetTeachersByIds :many
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
WHERE teacher.id IN (sqlc.slice('ids'));

-- name: GetTeachers :many
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM teacher JOIN user ON teacher.user_id = user.id
ORDER BY teacher.id
LIMIT ? OFFSET ?;

-- name: CountTeachers :one
SELECT Count(*) as total FROM teacher;

-- name: InsertTeacher :execlastid
INSERT INTO teacher ( user_id ) VALUES ( ? );

-- name: DeleteTeachersByIds :exec
DELETE FROM teacher
WHERE id IN (sqlc.slice('ids'));

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
ORDER BY student.id
LIMIT ? OFFSET ?;

-- name: GetStudentsByIds :many
SELECT student.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at
FROM student JOIN user ON student.user_id = user.id
WHERE student.id IN (sqlc.slice('ids'));

-- name: CountStudents :one
SELECT Count(*) as total FROM student;

-- name: InsertStudent :execlastid
INSERT INTO student ( user_id ) VALUES ( ? );

-- name: DeleteStudentsByIds :exec
DELETE FROM student
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteStudentByUserId :exec
DELETE FROM student
WHERE user_id = ?;

/* ============================== INSTRUMENT ============================== */
-- name: GetInstrumentById :one
SELECT * FROM instrument
WHERE id = ? LIMIT 1;

-- name: GetInstrumentsByIds :many
SELECT * FROM instrument
WHERE id IN (sqlc.slice('ids'));

-- name: GetInstruments :many
SELECT * FROM instrument
ORDER BY id
LIMIT ? OFFSET ?;

-- name: CountInstruments :one
SELECT Count(*) as total FROM instrument;

-- name: InsertInstrument :execlastid
INSERT INTO instrument ( name ) VALUES ( ? );

-- name: UpdateInstrument :exec
UPDATE instrument SET name = ?
WHERE id = ?;

-- name: DeleteInstrumentsByIds :exec
DELETE FROM instrument
WHERE id IN (sqlc.slice('ids'));

/* ============================== GRADE ============================== */
-- name: GetGradeById :one
SELECT * FROM grade
WHERE id = ? LIMIT 1;

-- name: GetGradesByIds :many
SELECT * FROM grade
WHERE id IN (sqlc.slice('ids'));

-- name: GetGrades :many
SELECT * FROM grade
ORDER BY id
LIMIT ? OFFSET ?;

-- name: CountGrades :one
SELECT Count(*) as total FROM grade;

-- name: InsertGrade :execlastid
INSERT INTO grade ( name ) VALUES ( ? );

-- name: UpdateGrade :exec
UPDATE grade SET name = ?
WHERE id = ?;

-- name: DeleteGradesByIds :exec
DELETE FROM grade
WHERE id IN (sqlc.slice('ids'));

/* ============================== COURSE ============================== */
-- name: GetCourses :many
SELECT course.id AS course_id, sqlc.embed(instrument), sqlc.embed(grade), default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
ORDER BY course.id
LIMIT ? OFFSET ?;

-- name: CountCourses :one
SELECT Count(*) as total FROM course;

-- name: GetCoursesByIds :many
SELECT course.id AS course_id, sqlc.embed(instrument), sqlc.embed(grade), default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
WHERE course.id IN (sqlc.slice('ids'));

-- name: GetCoursesByInstrumentId :many
SELECT course.id AS course_id, sqlc.embed(instrument), sqlc.embed(grade), default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
WHERE instrument.id = ?
ORDER BY course.id
LIMIT ? OFFSET ?;

-- name: GetCoursesByGradeId :many
SELECT course.id AS course_id, sqlc.embed(instrument), sqlc.embed(grade), default_fee, default_duration_minute
FROM course
    JOIN instrument ON instrument_id = instrument.id
    JOIN grade ON grade_id = grade.id
WHERE grade.id = ?
ORDER BY course.id
LIMIT ? OFFSET ?;

-- name: GetCourseById :one
SELECT course.id AS course_id, sqlc.embed(instrument), sqlc.embed(grade), default_fee, default_duration_minute
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

-- name: DeleteCoursesByIds :exec
DELETE FROM course
WHERE id IN (sqlc.slice('ids'));

/* ============================== CLASS ============================== */
-- name: GetClasses :many
WITH class_paginated AS (
    SELECT * FROM class
    WHERE class.is_deactivated IN (sqlc.slice('isDeactivateds'))
    LIMIT ? OFFSET ?
)
SELECT class_paginated.id AS class_id, transport_fee, class_paginated.is_deactivated, course_id, teacher_id, se.student_id AS student_id, se.id AS enrollment_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    sqlc.embed(instrument), sqlc.embed(grade),
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute
FROM class_paginated
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON (class_paginated.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
ORDER BY class_paginated.id;

-- name: CountClasses :one
SELECT Count(*) as total FROM class
WHERE is_deactivated IN (sqlc.slice('isDeactivateds'));

-- name: GetClassesByIds :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, teacher_id, se.student_id AS student_id, se.id AS enrollment_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    sqlc.embed(instrument), sqlc.embed(grade),
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON (class.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
WHERE class.id in (sqlc.slice('ids'))
ORDER BY class.id;

-- name: GetClassesByTeacherId :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, se.student_id AS student_id, se.id AS enrollment_id,
    sqlc.embed(instrument), sqlc.embed(grade),
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON (class.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
WHERE teacher_id = ?
ORDER BY class.id;

-- name: GetClassesByStudentId :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, teacher_id, se.id AS enrollment_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    sqlc.embed(instrument), sqlc.embed(grade),
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON (class.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
WHERE se.student_id = ?
ORDER BY class.id;

-- name: GetClassById :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, course_id, teacher_id, se.student_id AS student_id, se.id AS enrollment_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    sqlc.embed(instrument), sqlc.embed(grade),
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id

    LEFT JOIN student_enrollment AS se ON (class.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN user AS user_student ON se.student_id = user_student.id
WHERE class.id = ?;

-- name: InsertClass :execlastid
INSERT INTO class (
    transport_fee, teacher_id, course_id, is_deactivated
) VALUES (
    ?, ?, ?, ?
);

-- name: UpdateClass :exec
UPDATE class SET transport_fee = ?, teacher_id = ?, is_deactivated = ?
WHERE id = ?;

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

-- name: DeleteClassesByIds :exec
DELETE FROM class
WHERE id IN (sqlc.slice('ids'));

/* ============================== STUDENT_ENROLLMENT ============================== */
-- name: GetStudentEnrollmentsByIds :many
SELECT * FROM student_enrollment
WHERE id IN (sqlc.slice('ids'));

-- name: GetStudentEnrollmentsByStudentId :many
SELECT * FROM student_enrollment
WHERE student_id = ?;

-- name: GetStudentEnrollmentsByClassId :many
SELECT * FROM student_enrollment
WHERE class_id = ?;

-- name: GetStudentEnrollments :many
SELECT se.id as student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id AS class_id, class.transport_fee AS class_transport_fee, course_id, sqlc.embed(instrument), sqlc.embed(grade), course.default_fee AS course_default_fee
FROM student_enrollment as se
    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE se.is_deleted = 0
ORDER BY se.id;

-- name: InsertStudentEnrollment :exec
INSERT INTO student_enrollment (
    student_id, class_id
) VALUES (
    ?, ?
);

-- name: EnableStudentEnrollment :exec
UPDATE student_enrollment SET is_deleted = 0
WHERE id = ?;

-- name: DisableStudentEnrollment :exec
UPDATE student_enrollment SET is_deleted = 1
WHERE id = ?;

-- name: DeleteStudentEnrollmentById :exec
DELETE FROM student_enrollment
WHERE id = ?;

-- name: DeleteStudentEnrollmentsByIds :exec
DELETE FROM student_enrollment
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteStudentEnrollmentByStudentId :exec
DELETE FROM student_enrollment
WHERE student_id = ?;

-- name: DeleteStudentEnrollmentByClassIds :exec
DELETE FROM student_enrollment
WHERE class_id IN (sqlc.slice('classIds'));

/* ============================== TEACHER_SPECIAL_FEE ============================== */
-- name: GetTeacherSpecialFeeById :one
SELECT teacher_special_fee.id AS teacher_special_fee_id, fee,
    teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    course.id AS course_id, sqlc.embed(instrument), sqlc.embed(grade), default_fee AS original_course_fee
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
    course.id AS course_id, sqlc.embed(instrument), sqlc.embed(grade), default_fee AS original_course_fee
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

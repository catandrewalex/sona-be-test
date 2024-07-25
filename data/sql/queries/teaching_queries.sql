-- name: GetUserTeacherIdAndStudentId :one
SELECT user.id, teacher.id AS teacher_id, student.id AS student_id
FROM user
  LEFT JOIN teacher ON user.id = teacher.user_id
  LEFT JOIN student ON user.id = student.user_id
WHERE user.id = ? LIMIT 1;

-- name: IsUserIdInvolvedInClassId :one
SELECT EXISTS(
    -- check whether the user is enrolled in the class as a student
    SELECT user.id, se.class_id
    FROM user 
        JOIN student ON user.id = student.user_id
        JOIN student_enrollment AS se ON student.id = se.student_id
    WHERE user.id = sqlc.arg('user_id') AND se.class_id = sqlc.arg('class_id')
    UNION
    -- check whether the user is teaching the class
    SELECT user.id, class.id
    FROM user 
        JOIN teacher ON user.id = teacher.user_id
        JOIN class ON teacher.id = class.teacher_id
    WHERE user.id = sqlc.arg('user_id') AND class.id = sqlc.arg('class_id')
) AS is_involved;

-- name: IsUserIdInvolvedInAttendanceId :one
SELECT EXISTS(
    -- check whether the user is enrolled in the attendance's class as a student
    SELECT user.id, attendance.id
    FROM user 
        JOIN student ON user.id = student.user_id
        JOIN attendance ON student.id = attendance.student_id
    WHERE user.id = sqlc.arg('user_id') AND attendance.id = sqlc.arg('attendance_id')
    UNION
    -- check whether the user is teaching the attendance's class
    SELECT user.id, attendance.id
    FROM user 
        JOIN teacher ON user.id = teacher.user_id
        JOIN attendance ON teacher.id = attendance.teacher_id
    WHERE user.id = sqlc.arg('user_id') AND attendance.id = sqlc.arg('attendance_id')
) AS is_involved;

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

-- name: CountTeachersByIds :one
SELECT Count(id) AS total FROM teacher
WHERE id IN (sqlc.slice('ids'));

-- name: CountTeachers :one
SELECT Count(id) AS total FROM teacher;

-- name: InsertTeacher :execlastid
INSERT INTO teacher ( user_id ) VALUES ( ? );

-- name: DeleteTeachersByIds :exec
DELETE FROM teacher
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteTeacherByUserId :exec
DELETE FROM teacher
WHERE user_id = ?;

-- name: GetUnpaidTeachers :many
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at, sum(attendance.used_student_token_quota) AS total_attendances
FROM teacher
    JOIN user ON teacher.user_id = user.id
    JOIN attendance ON teacher.id = attendance.teacher_id
WHERE
    attendance.is_paid = 0
    AND (attendance.date >= sqlc.arg('startDate') AND attendance.date <= sqlc.arg('endDate'))
GROUP BY teacher.id
ORDER BY total_attendances DESC, teacher.id ASC
LIMIT ? OFFSET ?;

-- name: CountUnpaidTeachers :one
WITH unpaid_teacher AS (
    SELECT teacher_id, sum(attendance.used_student_token_quota) AS total_attendances
    FROM attendance
    WHERE
        is_paid = 0
        AND (attendance.date >= sqlc.arg('startDate') AND attendance.date <= sqlc.arg('endDate'))
    GROUP BY teacher_id
    ORDER BY total_attendances DESC, teacher_id ASC
)
SELECT Count(teacher_id) AS total FROM unpaid_teacher;

-- name: GetPaidTeachers :many
SELECT teacher.id, user.id AS user_id, username, email, user_detail, privilege_type, is_deactivated, created_at, sum(attendance.used_student_token_quota) AS total_attendances
FROM teacher
    JOIN user ON teacher.user_id = user.id
    JOIN attendance ON teacher.id = attendance.teacher_id
    RIGHT JOIN teacher_payment AS tp ON attendance.id = tp.attendance_id
WHERE
    (tp.added_at >= sqlc.arg('startDate') AND tp.added_at <= sqlc.arg('endDate'))
GROUP BY teacher.id
ORDER BY total_attendances DESC, teacher.id ASC
LIMIT ? OFFSET ?;

-- name: CountPaidTeachers :one
WITH paid_teacher AS (
    SELECT attendance.teacher_id, sum(attendance.used_student_token_quota) AS total_attendances
    FROM teacher_payment AS tp
        JOIN attendance ON tp.attendance_id = attendance.id
    WHERE
        (tp.added_at >= sqlc.arg('startDate') AND tp.added_at <= sqlc.arg('endDate'))
    GROUP BY teacher_id
    ORDER BY total_attendances DESC, teacher_id ASC
)
SELECT Count(teacher_id) AS total FROM paid_teacher;

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

-- name: CountStudentsByIds :one
SELECT Count(id) AS total FROM student
WHERE id IN (sqlc.slice('ids'));

-- name: CountStudents :one
SELECT Count(id) AS total FROM student;

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

-- name: CountInstrumentsByIds :one
SELECT Count(id) AS total FROM instrument
WHERE id IN (sqlc.slice('ids'));

-- name: CountInstruments :one
SELECT Count(id) AS total FROM instrument;

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

-- name: CountGradesByIds :one
SELECT Count(id) AS total FROM grade
WHERE id IN (sqlc.slice('ids'));

-- name: CountGrades :one
SELECT Count(*) AS total FROM grade;

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

-- name: CountCoursesByIds :one
SELECT Count(id) AS total FROM course
WHERE id IN (sqlc.slice('ids'));

-- name: CountCourses :one
SELECT Count(id) AS total FROM course;

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
-- name: GetClassTeacherId :one
SELECT class.teacher_id FROM class
WHERE class.id = ?
LIMIT 1;

-- name: GetClasses :many
WITH class_paginated AS (
    -- class & student_enrollment has a Many-to-many relationship, therefore:
    --   1. we need to join them to be able to filter by student_id
    --   2. we use SELECT DISCTINCT just to be safe
    SELECT DISTINCT class.id AS id, transport_fee, teacher_id, course_id, is_deactivated
    FROM class
        LEFT JOIN student_enrollment ON class.id = student_enrollment.class_id
    WHERE
        class.is_deactivated IN (sqlc.slice('isDeactivateds'))
        AND (class.teacher_id = sqlc.arg('teacher_id') OR sqlc.arg('use_teacher_filter') = false)
        AND (student_enrollment.student_id = sqlc.narg('student_id') OR sqlc.arg('use_student_filter') = false)
        AND (class.course_id = sqlc.narg('course_id') OR sqlc.arg('use_course_filter') = false)
    LIMIT ? OFFSET ?
)
SELECT class_paginated.id AS class_id, transport_fee, class_paginated.is_deactivated, class_paginated.course_id AS course_id, class_paginated.teacher_id AS teacher_id, se.student_id AS student_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    sqlc.embed(instrument), sqlc.embed(grade),
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute, tsf.fee AS teacher_special_fee
FROM class_paginated
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_enrollment AS se ON (class_paginated.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN student ON se.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id
ORDER BY class_paginated.id;

-- name: CountClassesByIds :one
SELECT Count(id) AS total FROM class
WHERE id IN (sqlc.slice('ids'));

-- name: CountClasses :one
WITH class_filtered AS (
    -- class & student_enrollment has a Many-to-many relationship, therefore:
    --   1. we need to join them to be able to filter by student_id
    --   2. we use SELECT DISCTINCT just to be safe
    SELECT DISTINCT class.id AS id, transport_fee, teacher_id, course_id, is_deactivated
    FROM class
        LEFT JOIN student_enrollment ON class.id = student_enrollment.class_id
    WHERE
        class.is_deactivated IN (sqlc.slice('isDeactivateds'))
        AND (teacher_id = sqlc.arg('teacher_id') OR sqlc.arg('use_teacher_filter') = false)
        AND (student_enrollment.student_id = sqlc.narg('student_id') OR sqlc.arg('use_student_filter') = false)
        AND (course_id = sqlc.narg('course_id') OR sqlc.arg('use_course_filter') = false)
)
SELECT Count(id) AS total FROM class_filtered;

-- name: GetClassesTotalStudentsByClassIds :many
SELECT class.id AS class_id, Count(student_enrollment.id) AS total_students
    FROM class
        LEFT JOIN student_enrollment ON class.id = student_enrollment.class_id
    WHERE class.id IN (sqlc.slice('ids'))
    GROUP BY class.id;

-- name: GetClassesByIds :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, class.course_id AS course_id, class.teacher_id AS teacher_id, se.student_id AS student_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    sqlc.embed(instrument), sqlc.embed(grade),
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute, tsf.fee AS teacher_special_fee
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_enrollment AS se ON (class.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN student ON se.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id
WHERE class.id in (sqlc.slice('ids'))
ORDER BY class.id;

-- name: GetClassById :many
SELECT class.id AS class_id, transport_fee, class.is_deactivated, class.course_id AS course_id, class.teacher_id AS teacher_id, se.student_id AS student_id,
    user_teacher.username AS teacher_username,
    user_teacher.user_detail AS teacher_detail,
    sqlc.embed(instrument), sqlc.embed(grade),
    user_student.username AS student_username,
    user_student.user_detail AS student_detail,
    course.default_fee, course.default_duration_minute, tsf.fee AS teacher_special_fee
FROM class
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id

    LEFT JOIN teacher ON teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_enrollment AS se ON (class.id = se.class_id AND se.is_deleted=0)
    LEFT JOIN student ON se.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id
WHERE class.id = ?;

-- name: InsertClass :execlastid
INSERT INTO class (
    transport_fee, teacher_id, course_id, is_deactivated
) VALUES (
    ?, ?, ?, ?
);

-- name: UpdateClass :exec
UPDATE class SET transport_fee = ?, teacher_id = ?, course_id = ?, is_deactivated = ?
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
-- name: GetStudentEnrollmentById :one
SELECT se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_enrollment AS se
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE se.is_deleted = 0 AND se.id = ?;

-- name: GetStudentEnrollmentsByIds :many
SELECT se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_enrollment AS se
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE se.is_deleted = 0 AND se.id IN (sqlc.slice('ids'));

-- name: GetStudentEnrollmentsByStudentId :many
SELECT se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_enrollment AS se
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE se.is_deleted = 0 AND student_id = ?;

-- name: GetStudentEnrollmentsByClassId :many
SELECT se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_enrollment AS se
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE se.is_deleted = 0 AND class_id = ?;

-- name: GetStudentEnrollments :many
SELECT se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    sqlc.embed(class), tsf.fee AS teacher_special_fee, sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade),
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail
FROM student_enrollment AS se
    JOIN student ON se.student_id = student.id
    JOIN user AS user_student ON student.user_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)
WHERE se.is_deleted = 0
ORDER BY se.id
LIMIT ? OFFSET ?;

-- name: CountStudentEnrollments :one
SELECT COUNT(id) FROM student_enrollment
WHERE is_deleted = 0;

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
-- name: GetTeacherSpecialFees :many
SELECT teacher_special_fee.id AS teacher_special_fee_id, fee,
    teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade)
FROM teacher_special_fee
    JOIN teacher ON teacher_id = teacher.id
    JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    JOIN course on course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
ORDER BY course.id
LIMIT ? OFFSET ?;

-- name: CountTeacherSpecialFeesByIds :one
SELECT Count(id) AS total FROM teacher_special_fee
WHERE id IN (sqlc.slice('ids'));

-- name: CountTeacherSpecialFees :one
SELECT Count(id) AS total FROM teacher_special_fee;

-- name: GetTeacherSpecialFeeById :one
SELECT teacher_special_fee.id AS teacher_special_fee_id, fee,
    teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade)
FROM teacher_special_fee
    JOIN teacher ON teacher_id = teacher.id
    JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    JOIN course on course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE teacher_special_fee.id = ? LIMIT 1;

-- name: GetTeacherSpecialFeesByIds :many
SELECT teacher_special_fee.id AS teacher_special_fee_id, fee,
    teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade)
FROM teacher_special_fee
    JOIN teacher ON teacher_id = teacher.id
    JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    JOIN course on course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE teacher_special_fee.id IN (sqlc.slice('ids'));

-- name: GetTeacherSpecialFeesByTeacherId :many
SELECT teacher_special_fee.id AS teacher_special_fee_id, fee,
    teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    sqlc.embed(course), sqlc.embed(instrument), sqlc.embed(grade)
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
WHERE id = ?;

-- name: DeleteTeacherSpecialFeeById :exec
DELETE FROM teacher_special_fee
WHERE id = ?;

-- name: DeleteTeacherSpecialFeesByIds :exec
DELETE FROM teacher_special_fee
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteTeacherSpecialFeeByTeacherId :exec
DELETE FROM teacher_special_fee
WHERE teacher_id = ?;

-- name: DeleteTeacherSpecialFeeByCourseId :exec
DELETE FROM teacher_special_fee
WHERE course_id = ?;

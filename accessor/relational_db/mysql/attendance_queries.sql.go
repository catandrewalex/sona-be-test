// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: attendance_queries.sql

package mysql

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

const countAttendances = `-- name: CountAttendances :one
SELECT Count(id) AS total FROM attendance
WHERE
    (date >= ? AND date <= ?)
    AND (class_id = ? OR ? = false)
    AND (student_id = ? OR ? = false)
    AND (is_paid = 0 OR ? = false)
`

type CountAttendancesParams struct {
	StartDate        time.Time
	EndDate          time.Time
	ClassID          sql.NullInt64
	UseClassFilter   interface{}
	StudentID        sql.NullInt64
	UseStudentFilter interface{}
	UsePaidFilter    interface{}
}

func (q *Queries) CountAttendances(ctx context.Context, arg CountAttendancesParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countAttendances,
		arg.StartDate,
		arg.EndDate,
		arg.ClassID,
		arg.UseClassFilter,
		arg.StudentID,
		arg.UseStudentFilter,
		arg.UsePaidFilter,
	)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const countAttendancesByIds = `-- name: CountAttendancesByIds :one
SELECT Count(id) AS total FROM attendance
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) CountAttendancesByIds(ctx context.Context, ids []int64) (int64, error) {
	sql := countAttendancesByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		sql = strings.Replace(sql, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		sql = strings.Replace(sql, "/*SLICE:ids*/?", "NULL", 1)
	}
	row := q.db.QueryRowContext(ctx, sql, queryParams...)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const deleteAttendanceById = `-- name: DeleteAttendanceById :exec
DELETE FROM attendance
WHERE id = ?
`

func (q *Queries) DeleteAttendanceById(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteAttendanceById, id)
	return err
}

const deleteAttendancesByIds = `-- name: DeleteAttendancesByIds :exec
DELETE FROM attendance
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) DeleteAttendancesByIds(ctx context.Context, ids []int64) error {
	sql := deleteAttendancesByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		sql = strings.Replace(sql, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		sql = strings.Replace(sql, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := q.db.ExecContext(ctx, sql, queryParams...)
	return err
}

const editAttendances = `-- name: EditAttendances :exec
UPDATE attendance
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, teacher_id = ?
WHERE id IN (/*SLICE:ids*/?)
`

type EditAttendancesParams struct {
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	TeacherID             sql.NullInt64
	Ids                   []int64
}

func (q *Queries) EditAttendances(ctx context.Context, arg EditAttendancesParams) error {
	sql := editAttendances
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Date)
	queryParams = append(queryParams, arg.UsedStudentTokenQuota)
	queryParams = append(queryParams, arg.Duration)
	queryParams = append(queryParams, arg.Note)
	queryParams = append(queryParams, arg.TeacherID)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		sql = strings.Replace(sql, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		sql = strings.Replace(sql, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := q.db.ExecContext(ctx, sql, queryParams...)
	return err
}

const getAttendanceById = `-- name: GetAttendanceById :one
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    slt.id, slt.quota, slt.course_fee_value, slt.transport_fee_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id = ? LIMIT 1
`

type GetAttendanceByIdRow struct {
	AttendanceID          int64
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	IsPaid                int32
	Class                 Class
	Course                Course
	Instrument            Instrument
	Grade                 Grade
	TeacherID             sql.NullInt64
	TeacherUsername       sql.NullString
	TeacherDetail         []byte
	StudentID             sql.NullInt64
	StudentUsername       sql.NullString
	StudentDetail         []byte
	ClassTeacherID        sql.NullInt64
	ClassTeacherUsername  sql.NullString
	ClassTeacherDetail    []byte
	StudentLearningToken  StudentLearningToken
}

func (q *Queries) GetAttendanceById(ctx context.Context, id int64) (GetAttendanceByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getAttendanceById, id)
	var i GetAttendanceByIdRow
	err := row.Scan(
		&i.AttendanceID,
		&i.Date,
		&i.UsedStudentTokenQuota,
		&i.Duration,
		&i.Note,
		&i.IsPaid,
		&i.Class.ID,
		&i.Class.TransportFee,
		&i.Class.TeacherID,
		&i.Class.CourseID,
		&i.Class.IsDeactivated,
		&i.Course.ID,
		&i.Course.DefaultFee,
		&i.Course.DefaultDurationMinute,
		&i.Course.InstrumentID,
		&i.Course.GradeID,
		&i.Instrument.ID,
		&i.Instrument.Name,
		&i.Grade.ID,
		&i.Grade.Name,
		&i.TeacherID,
		&i.TeacherUsername,
		&i.TeacherDetail,
		&i.StudentID,
		&i.StudentUsername,
		&i.StudentDetail,
		&i.ClassTeacherID,
		&i.ClassTeacherUsername,
		&i.ClassTeacherDetail,
		&i.StudentLearningToken.ID,
		&i.StudentLearningToken.Quota,
		&i.StudentLearningToken.CourseFeeValue,
		&i.StudentLearningToken.TransportFeeValue,
		&i.StudentLearningToken.CreatedAt,
		&i.StudentLearningToken.LastUpdatedAt,
		&i.StudentLearningToken.EnrollmentID,
	)
	return i, err
}

const getAttendanceIdsOfSameClassAndDate = `-- name: GetAttendanceIdsOfSameClassAndDate :many
WITH ref_attendance AS (
    SELECT id, date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id
    FROM attendance
    WHERE attendance.id = ?
)
SELECT attendance.id AS id, attendance.is_paid AS is_paid
FROM attendance
    JOIN ref_attendance ON attendance.class_id = ref_attendance.class_id AND attendance.date = ref_attendance.date
ORDER by attendance.id
`

type GetAttendanceIdsOfSameClassAndDateRow struct {
	ID     int64
	IsPaid int32
}

// ============================== ATTENDANCE ==============================
func (q *Queries) GetAttendanceIdsOfSameClassAndDate(ctx context.Context, id int64) ([]GetAttendanceIdsOfSameClassAndDateRow, error) {
	rows, err := q.db.QueryContext(ctx, getAttendanceIdsOfSameClassAndDate, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAttendanceIdsOfSameClassAndDateRow
	for rows.Next() {
		var i GetAttendanceIdsOfSameClassAndDateRow
		if err := rows.Scan(&i.ID, &i.IsPaid); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAttendances = `-- name: GetAttendances :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    slt.id, slt.quota, slt.course_fee_value, slt.transport_fee_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= ? AND attendance.date <= ?)
    AND (class_id = ? OR ? = false)
    AND (student_id = ? OR ? = false)
    AND (is_paid = 0 OR ? = false)
ORDER BY class.id
LIMIT ? OFFSET ?
`

type GetAttendancesParams struct {
	StartDate        time.Time
	EndDate          time.Time
	ClassID          sql.NullInt64
	UseClassFilter   interface{}
	StudentID        sql.NullInt64
	UseStudentFilter interface{}
	UseUnpaidFilter  interface{}
	Limit            int32
	Offset           int32
}

type GetAttendancesRow struct {
	AttendanceID          int64
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	IsPaid                int32
	Class                 Class
	Course                Course
	Instrument            Instrument
	Grade                 Grade
	TeacherID             sql.NullInt64
	TeacherUsername       sql.NullString
	TeacherDetail         []byte
	StudentID             sql.NullInt64
	StudentUsername       sql.NullString
	StudentDetail         []byte
	ClassTeacherID        sql.NullInt64
	ClassTeacherUsername  sql.NullString
	ClassTeacherDetail    []byte
	StudentLearningToken  StudentLearningToken
}

func (q *Queries) GetAttendances(ctx context.Context, arg GetAttendancesParams) ([]GetAttendancesRow, error) {
	rows, err := q.db.QueryContext(ctx, getAttendances,
		arg.StartDate,
		arg.EndDate,
		arg.ClassID,
		arg.UseClassFilter,
		arg.StudentID,
		arg.UseStudentFilter,
		arg.UseUnpaidFilter,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAttendancesRow
	for rows.Next() {
		var i GetAttendancesRow
		if err := rows.Scan(
			&i.AttendanceID,
			&i.Date,
			&i.UsedStudentTokenQuota,
			&i.Duration,
			&i.Note,
			&i.IsPaid,
			&i.Class.ID,
			&i.Class.TransportFee,
			&i.Class.TeacherID,
			&i.Class.CourseID,
			&i.Class.IsDeactivated,
			&i.Course.ID,
			&i.Course.DefaultFee,
			&i.Course.DefaultDurationMinute,
			&i.Course.InstrumentID,
			&i.Course.GradeID,
			&i.Instrument.ID,
			&i.Instrument.Name,
			&i.Grade.ID,
			&i.Grade.Name,
			&i.TeacherID,
			&i.TeacherUsername,
			&i.TeacherDetail,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
			&i.ClassTeacherID,
			&i.ClassTeacherUsername,
			&i.ClassTeacherDetail,
			&i.StudentLearningToken.ID,
			&i.StudentLearningToken.Quota,
			&i.StudentLearningToken.CourseFeeValue,
			&i.StudentLearningToken.TransportFeeValue,
			&i.StudentLearningToken.CreatedAt,
			&i.StudentLearningToken.LastUpdatedAt,
			&i.StudentLearningToken.EnrollmentID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAttendancesByIds = `-- name: GetAttendancesByIds :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    slt.id, slt.quota, slt.course_fee_value, slt.transport_fee_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id IN (/*SLICE:ids*/?)
`

type GetAttendancesByIdsRow struct {
	AttendanceID          int64
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	IsPaid                int32
	Class                 Class
	Course                Course
	Instrument            Instrument
	Grade                 Grade
	TeacherID             sql.NullInt64
	TeacherUsername       sql.NullString
	TeacherDetail         []byte
	StudentID             sql.NullInt64
	StudentUsername       sql.NullString
	StudentDetail         []byte
	ClassTeacherID        sql.NullInt64
	ClassTeacherUsername  sql.NullString
	ClassTeacherDetail    []byte
	StudentLearningToken  StudentLearningToken
}

func (q *Queries) GetAttendancesByIds(ctx context.Context, ids []int64) ([]GetAttendancesByIdsRow, error) {
	sql := getAttendancesByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		sql = strings.Replace(sql, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		sql = strings.Replace(sql, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, sql, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAttendancesByIdsRow
	for rows.Next() {
		var i GetAttendancesByIdsRow
		if err := rows.Scan(
			&i.AttendanceID,
			&i.Date,
			&i.UsedStudentTokenQuota,
			&i.Duration,
			&i.Note,
			&i.IsPaid,
			&i.Class.ID,
			&i.Class.TransportFee,
			&i.Class.TeacherID,
			&i.Class.CourseID,
			&i.Class.IsDeactivated,
			&i.Course.ID,
			&i.Course.DefaultFee,
			&i.Course.DefaultDurationMinute,
			&i.Course.InstrumentID,
			&i.Course.GradeID,
			&i.Instrument.ID,
			&i.Instrument.Name,
			&i.Grade.ID,
			&i.Grade.Name,
			&i.TeacherID,
			&i.TeacherUsername,
			&i.TeacherDetail,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
			&i.ClassTeacherID,
			&i.ClassTeacherUsername,
			&i.ClassTeacherDetail,
			&i.StudentLearningToken.ID,
			&i.StudentLearningToken.Quota,
			&i.StudentLearningToken.CourseFeeValue,
			&i.StudentLearningToken.TransportFeeValue,
			&i.StudentLearningToken.CreatedAt,
			&i.StudentLearningToken.LastUpdatedAt,
			&i.StudentLearningToken.EnrollmentID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAttendancesForTeacherSalary = `-- name: GetAttendancesForTeacherSalary :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    slt.id, slt.quota, slt.course_fee_value, slt.transport_fee_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON attendance.student_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= ? AND attendance.date <= ?)
    AND (attendance.teacher_id = ? OR ? = false)
    AND (class_id = ? OR ? = false)
    AND is_paid = 0
ORDER BY attendance.teacher_id, class.id, attendance.student_id, date, attendance.id
`

type GetAttendancesForTeacherSalaryParams struct {
	StartDate        time.Time
	EndDate          time.Time
	TeacherID        sql.NullInt64
	UseTeacherFilter interface{}
	ClassID          sql.NullInt64
	UseClassFilter   interface{}
}

type GetAttendancesForTeacherSalaryRow struct {
	AttendanceID          int64
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	IsPaid                int32
	Class                 Class
	Course                Course
	Instrument            Instrument
	Grade                 Grade
	TeacherID             sql.NullInt64
	TeacherUsername       sql.NullString
	TeacherDetail         []byte
	StudentID             sql.NullInt64
	StudentUsername       sql.NullString
	StudentDetail         []byte
	ClassTeacherID        sql.NullInt64
	ClassTeacherUsername  sql.NullString
	ClassTeacherDetail    []byte
	StudentLearningToken  StudentLearningToken
}

func (q *Queries) GetAttendancesForTeacherSalary(ctx context.Context, arg GetAttendancesForTeacherSalaryParams) ([]GetAttendancesForTeacherSalaryRow, error) {
	rows, err := q.db.QueryContext(ctx, getAttendancesForTeacherSalary,
		arg.StartDate,
		arg.EndDate,
		arg.TeacherID,
		arg.UseTeacherFilter,
		arg.ClassID,
		arg.UseClassFilter,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAttendancesForTeacherSalaryRow
	for rows.Next() {
		var i GetAttendancesForTeacherSalaryRow
		if err := rows.Scan(
			&i.AttendanceID,
			&i.Date,
			&i.UsedStudentTokenQuota,
			&i.Duration,
			&i.Note,
			&i.IsPaid,
			&i.Class.ID,
			&i.Class.TransportFee,
			&i.Class.TeacherID,
			&i.Class.CourseID,
			&i.Class.IsDeactivated,
			&i.Course.ID,
			&i.Course.DefaultFee,
			&i.Course.DefaultDurationMinute,
			&i.Course.InstrumentID,
			&i.Course.GradeID,
			&i.Instrument.ID,
			&i.Instrument.Name,
			&i.Grade.ID,
			&i.Grade.Name,
			&i.TeacherID,
			&i.TeacherUsername,
			&i.TeacherDetail,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
			&i.ClassTeacherID,
			&i.ClassTeacherUsername,
			&i.ClassTeacherDetail,
			&i.StudentLearningToken.ID,
			&i.StudentLearningToken.Quota,
			&i.StudentLearningToken.CourseFeeValue,
			&i.StudentLearningToken.TransportFeeValue,
			&i.StudentLearningToken.CreatedAt,
			&i.StudentLearningToken.LastUpdatedAt,
			&i.StudentLearningToken.EnrollmentID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertAttendance = `-- name: InsertAttendance :execlastid
INSERT INTO attendance (
    date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`

type InsertAttendanceParams struct {
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	IsPaid                int32
	ClassID               sql.NullInt64
	TeacherID             sql.NullInt64
	StudentID             sql.NullInt64
	TokenID               int64
}

func (q *Queries) InsertAttendance(ctx context.Context, arg InsertAttendanceParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertAttendance,
		arg.Date,
		arg.UsedStudentTokenQuota,
		arg.Duration,
		arg.Note,
		arg.IsPaid,
		arg.ClassID,
		arg.TeacherID,
		arg.StudentID,
		arg.TokenID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const setAttendancesIsPaidStatusByIds = `-- name: SetAttendancesIsPaidStatusByIds :exec
UPDATE attendance SET is_paid = ?
WHERE id IN (/*SLICE:ids*/?)
`

type SetAttendancesIsPaidStatusByIdsParams struct {
	IsPaid int32
	Ids    []int64
}

func (q *Queries) SetAttendancesIsPaidStatusByIds(ctx context.Context, arg SetAttendancesIsPaidStatusByIdsParams) error {
	sql := setAttendancesIsPaidStatusByIds
	var queryParams []interface{}
	queryParams = append(queryParams, arg.IsPaid)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		sql = strings.Replace(sql, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		sql = strings.Replace(sql, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := q.db.ExecContext(ctx, sql, queryParams...)
	return err
}

const updateAttendance = `-- name: UpdateAttendance :exec
UPDATE attendance
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, is_paid = ?, class_id = ?, teacher_id = ?, student_id = ?, token_id = ?
WHERE id = ?
`

type UpdateAttendanceParams struct {
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	IsPaid                int32
	ClassID               sql.NullInt64
	TeacherID             sql.NullInt64
	StudentID             sql.NullInt64
	TokenID               int64
	ID                    int64
}

func (q *Queries) UpdateAttendance(ctx context.Context, arg UpdateAttendanceParams) error {
	_, err := q.db.ExecContext(ctx, updateAttendance,
		arg.Date,
		arg.UsedStudentTokenQuota,
		arg.Duration,
		arg.Note,
		arg.IsPaid,
		arg.ClassID,
		arg.TeacherID,
		arg.StudentID,
		arg.TokenID,
		arg.ID,
	)
	return err
}
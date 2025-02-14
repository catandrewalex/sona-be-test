// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: attendance_queries.sql

package mysql

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

const assignAttendanceToken = `-- name: AssignAttendanceToken :exec
UPDATE attendance
SET token_id = ?
WHERE id = ?
`

type AssignAttendanceTokenParams struct {
	TokenID sql.NullInt64
	ID      int64
}

func (q *Queries) AssignAttendanceToken(ctx context.Context, arg AssignAttendanceTokenParams) error {
	_, err := q.db.ExecContext(ctx, assignAttendanceToken, arg.TokenID, arg.ID)
	return err
}

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
	ClassID          int64
	UseClassFilter   interface{}
	StudentID        int64
	UseStudentFilter interface{}
	UseUnpaidFilter  interface{}
}

func (q *Queries) CountAttendances(ctx context.Context, arg CountAttendancesParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countAttendances,
		arg.StartDate,
		arg.EndDate,
		arg.ClassID,
		arg.UseClassFilter,
		arg.StudentID,
		arg.UseStudentFilter,
		arg.UseUnpaidFilter,
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
	query := countAttendancesByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	row := q.db.QueryRowContext(ctx, query, queryParams...)
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
	query := deleteAttendancesByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := q.db.ExecContext(ctx, query, queryParams...)
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
	TeacherID             int64
	Ids                   []int64
}

func (q *Queries) EditAttendances(ctx context.Context, arg EditAttendancesParams) error {
	query := editAttendances
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
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := q.db.ExecContext(ctx, query, queryParams...)
	return err
}

const getAttendanceById = `-- name: GetAttendanceById :one
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.auto_owe_attendance_token, class.is_deactivated, tsf.fee AS teacher_special_fee, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    -- we cannot use sqlc.embed(slt), due to ` + "`" + `Attendance` + "`" + ` may have null ` + "`" + `StudentLearningToken` + "`" + `.
    -- SQLC has not yet had the capability to create pointer to struct, when the join result could be null.
    slt.id, slt.quota, slt.course_fee_quarter_value, slt.transport_fee_quarter_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id = ? LIMIT 1
`

type GetAttendanceByIdRow struct {
	AttendanceID             int64
	Date                     time.Time
	UsedStudentTokenQuota    float64
	Duration                 int32
	Note                     string
	IsPaid                   int32
	Class                    Class
	TeacherSpecialFee        sql.NullInt32
	Course                   Course
	Instrument               Instrument
	Grade                    Grade
	TeacherID                int64
	TeacherUsername          sql.NullString
	TeacherDetail            []byte
	StudentID                int64
	StudentUsername          sql.NullString
	StudentDetail            []byte
	ClassTeacherID           sql.NullInt64
	ClassTeacherUsername     sql.NullString
	ClassTeacherDetail       []byte
	ID                       sql.NullInt64
	Quota                    sql.NullFloat64
	CourseFeeQuarterValue    sql.NullInt32
	TransportFeeQuarterValue sql.NullInt32
	CreatedAt                sql.NullTime
	LastUpdatedAt            sql.NullTime
	EnrollmentID             sql.NullInt64
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
		&i.Class.AutoOweAttendanceToken,
		&i.Class.IsDeactivated,
		&i.TeacherSpecialFee,
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
		&i.ID,
		&i.Quota,
		&i.CourseFeeQuarterValue,
		&i.TransportFeeQuarterValue,
		&i.CreatedAt,
		&i.LastUpdatedAt,
		&i.EnrollmentID,
	)
	return i, err
}

const getAttendanceForTokenAssignment = `-- name: GetAttendanceForTokenAssignment :one
SELECT used_student_token_quota, is_paid, token_id
FROM attendance
WHERE id = ? FOR UPDATE
`

type GetAttendanceForTokenAssignmentRow struct {
	UsedStudentTokenQuota float64
	IsPaid                int32
	TokenID               sql.NullInt64
}

// note that as this query use "FOR UPDATE", it will block other query from reading this attendance record.
func (q *Queries) GetAttendanceForTokenAssignment(ctx context.Context, id int64) (GetAttendanceForTokenAssignmentRow, error) {
	row := q.db.QueryRowContext(ctx, getAttendanceForTokenAssignment, id)
	var i GetAttendanceForTokenAssignmentRow
	err := row.Scan(&i.UsedStudentTokenQuota, &i.IsPaid, &i.TokenID)
	return i, err
}

const getAttendanceIdsOfSameClassAndDate = `-- name: GetAttendanceIdsOfSameClassAndDate :many
WITH ref_attendance AS (
    SELECT id, date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id
    FROM attendance
    WHERE attendance.id = ?
)
SELECT attendance.id AS id, attendance.is_paid AS is_paid, attendance.token_id, attendance.used_student_token_quota
FROM attendance
    JOIN ref_attendance ON attendance.class_id = ref_attendance.class_id AND attendance.date = ref_attendance.date
ORDER by attendance.id
`

type GetAttendanceIdsOfSameClassAndDateRow struct {
	ID                    int64
	IsPaid                int32
	TokenID               sql.NullInt64
	UsedStudentTokenQuota float64
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
		if err := rows.Scan(
			&i.ID,
			&i.IsPaid,
			&i.TokenID,
			&i.UsedStudentTokenQuota,
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

const getAttendances = `-- name: GetAttendances :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.auto_owe_attendance_token, class.is_deactivated, tsf.fee AS teacher_special_fee, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    -- we cannot use sqlc.embed(slt), due to ` + "`" + `Attendance` + "`" + ` may have null ` + "`" + `StudentLearningToken` + "`" + `.
    -- SQLC has not yet had the capability to create pointer to struct, when the join result could be null.
    slt.id, slt.quota, slt.course_fee_quarter_value, slt.transport_fee_quarter_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= ? AND attendance.date <= ?)
    AND (class_id = ? OR ? = false)
    AND (student_id = ? OR ? = false)
    AND (is_paid = 0 OR ? = false)
ORDER BY attendance.id
LIMIT ? OFFSET ?
`

type GetAttendancesParams struct {
	StartDate        time.Time
	EndDate          time.Time
	ClassID          int64
	UseClassFilter   interface{}
	StudentID        int64
	UseStudentFilter interface{}
	UseUnpaidFilter  interface{}
	Limit            int32
	Offset           int32
}

type GetAttendancesRow struct {
	AttendanceID             int64
	Date                     time.Time
	UsedStudentTokenQuota    float64
	Duration                 int32
	Note                     string
	IsPaid                   int32
	Class                    Class
	TeacherSpecialFee        sql.NullInt32
	Course                   Course
	Instrument               Instrument
	Grade                    Grade
	TeacherID                int64
	TeacherUsername          sql.NullString
	TeacherDetail            []byte
	StudentID                int64
	StudentUsername          sql.NullString
	StudentDetail            []byte
	ClassTeacherID           sql.NullInt64
	ClassTeacherUsername     sql.NullString
	ClassTeacherDetail       []byte
	ID                       sql.NullInt64
	Quota                    sql.NullFloat64
	CourseFeeQuarterValue    sql.NullInt32
	TransportFeeQuarterValue sql.NullInt32
	CreatedAt                sql.NullTime
	LastUpdatedAt            sql.NullTime
	EnrollmentID             sql.NullInt64
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
			&i.Class.AutoOweAttendanceToken,
			&i.Class.IsDeactivated,
			&i.TeacherSpecialFee,
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
			&i.ID,
			&i.Quota,
			&i.CourseFeeQuarterValue,
			&i.TransportFeeQuarterValue,
			&i.CreatedAt,
			&i.LastUpdatedAt,
			&i.EnrollmentID,
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
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.auto_owe_attendance_token, class.is_deactivated, tsf.fee AS teacher_special_fee, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    -- we cannot use sqlc.embed(slt), due to ` + "`" + `Attendance` + "`" + ` may have null ` + "`" + `StudentLearningToken` + "`" + `.
    -- SQLC has not yet had the capability to create pointer to struct, when the join result could be null.
    slt.id, slt.quota, slt.course_fee_quarter_value, slt.transport_fee_quarter_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE attendance.id IN (/*SLICE:ids*/?)
`

type GetAttendancesByIdsRow struct {
	AttendanceID             int64
	Date                     time.Time
	UsedStudentTokenQuota    float64
	Duration                 int32
	Note                     string
	IsPaid                   int32
	Class                    Class
	TeacherSpecialFee        sql.NullInt32
	Course                   Course
	Instrument               Instrument
	Grade                    Grade
	TeacherID                int64
	TeacherUsername          sql.NullString
	TeacherDetail            []byte
	StudentID                int64
	StudentUsername          sql.NullString
	StudentDetail            []byte
	ClassTeacherID           sql.NullInt64
	ClassTeacherUsername     sql.NullString
	ClassTeacherDetail       []byte
	ID                       sql.NullInt64
	Quota                    sql.NullFloat64
	CourseFeeQuarterValue    sql.NullInt32
	TransportFeeQuarterValue sql.NullInt32
	CreatedAt                sql.NullTime
	LastUpdatedAt            sql.NullTime
	EnrollmentID             sql.NullInt64
}

func (q *Queries) GetAttendancesByIds(ctx context.Context, ids []int64) ([]GetAttendancesByIdsRow, error) {
	query := getAttendancesByIds
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
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
			&i.Class.AutoOweAttendanceToken,
			&i.Class.IsDeactivated,
			&i.TeacherSpecialFee,
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
			&i.ID,
			&i.Quota,
			&i.CourseFeeQuarterValue,
			&i.TransportFeeQuarterValue,
			&i.CreatedAt,
			&i.LastUpdatedAt,
			&i.EnrollmentID,
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

const getAttendancesDescendingDate = `-- name: GetAttendancesDescendingDate :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.auto_owe_attendance_token, class.is_deactivated, tsf.fee AS teacher_special_fee, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    -- we cannot use sqlc.embed(slt), due to ` + "`" + `Attendance` + "`" + ` may have null ` + "`" + `StudentLearningToken` + "`" + `.
    -- SQLC has not yet had the capability to create pointer to struct, when the join result could be null.
    slt.id, slt.quota, slt.course_fee_quarter_value, slt.transport_fee_quarter_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= ? AND attendance.date <= ?)
    AND (class_id = ? OR ? = false)
    AND (student_id = ? OR ? = false)
    AND (is_paid = 0 OR ? = false)
ORDER BY attendance.date DESC, attendance.id ASC
LIMIT ? OFFSET ?
`

type GetAttendancesDescendingDateParams struct {
	StartDate        time.Time
	EndDate          time.Time
	ClassID          int64
	UseClassFilter   interface{}
	StudentID        int64
	UseStudentFilter interface{}
	UseUnpaidFilter  interface{}
	Limit            int32
	Offset           int32
}

type GetAttendancesDescendingDateRow struct {
	AttendanceID             int64
	Date                     time.Time
	UsedStudentTokenQuota    float64
	Duration                 int32
	Note                     string
	IsPaid                   int32
	Class                    Class
	TeacherSpecialFee        sql.NullInt32
	Course                   Course
	Instrument               Instrument
	Grade                    Grade
	TeacherID                int64
	TeacherUsername          sql.NullString
	TeacherDetail            []byte
	StudentID                int64
	StudentUsername          sql.NullString
	StudentDetail            []byte
	ClassTeacherID           sql.NullInt64
	ClassTeacherUsername     sql.NullString
	ClassTeacherDetail       []byte
	ID                       sql.NullInt64
	Quota                    sql.NullFloat64
	CourseFeeQuarterValue    sql.NullInt32
	TransportFeeQuarterValue sql.NullInt32
	CreatedAt                sql.NullTime
	LastUpdatedAt            sql.NullTime
	EnrollmentID             sql.NullInt64
}

// GetAttendancesDescendingDate is a copy of GetAttendances, with additional sort by date parameter. TODO: find alternative: sqlc's dynamic query which is mature enough, so that we need to do this.
func (q *Queries) GetAttendancesDescendingDate(ctx context.Context, arg GetAttendancesDescendingDateParams) ([]GetAttendancesDescendingDateRow, error) {
	rows, err := q.db.QueryContext(ctx, getAttendancesDescendingDate,
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
	var items []GetAttendancesDescendingDateRow
	for rows.Next() {
		var i GetAttendancesDescendingDateRow
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
			&i.Class.AutoOweAttendanceToken,
			&i.Class.IsDeactivated,
			&i.TeacherSpecialFee,
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
			&i.ID,
			&i.Quota,
			&i.CourseFeeQuarterValue,
			&i.TransportFeeQuarterValue,
			&i.CreatedAt,
			&i.LastUpdatedAt,
			&i.EnrollmentID,
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

const getUnpaidAttendancesByTeacherId = `-- name: GetUnpaidAttendancesByTeacherId :many
SELECT attendance.id AS attendance_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.auto_owe_attendance_token, class.is_deactivated, tsf.fee AS teacher_special_fee, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    attendance.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    attendance.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    -- we cannot use sqlc.embed(slt), due to ` + "`" + `Attendance` + "`" + ` may have null ` + "`" + `StudentLearningToken` + "`" + `.
    -- SQLC has not yet had the capability to create pointer to struct, when the join result could be null.
    slt.id, slt.quota, slt.course_fee_quarter_value, slt.transport_fee_quarter_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM attendance
    LEFT JOIN teacher ON attendance.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN student ON attendance.student_id = student.id
    LEFT JOIN user AS user_student ON student.user_id = user_student.id

    LEFT JOIN class on attendance.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id
    LEFT JOIN teacher_special_fee AS tsf ON (class_teacher.id = tsf.teacher_id AND course.id = tsf.course_id)

    LEFT JOIN student_learning_token as slt ON attendance.token_id = slt.id
WHERE
    (attendance.date >= ? AND attendance.date <= ?)
    AND attendance.teacher_id = ?
    AND is_paid = 0
ORDER BY date DESC, attendance.id ASC
`

type GetUnpaidAttendancesByTeacherIdParams struct {
	StartDate time.Time
	EndDate   time.Time
	TeacherID int64
}

type GetUnpaidAttendancesByTeacherIdRow struct {
	AttendanceID             int64
	Date                     time.Time
	UsedStudentTokenQuota    float64
	Duration                 int32
	Note                     string
	IsPaid                   int32
	Class                    Class
	TeacherSpecialFee        sql.NullInt32
	Course                   Course
	Instrument               Instrument
	Grade                    Grade
	TeacherID                int64
	TeacherUsername          sql.NullString
	TeacherDetail            []byte
	StudentID                int64
	StudentUsername          sql.NullString
	StudentDetail            []byte
	ClassTeacherID           sql.NullInt64
	ClassTeacherUsername     sql.NullString
	ClassTeacherDetail       []byte
	ID                       sql.NullInt64
	Quota                    sql.NullFloat64
	CourseFeeQuarterValue    sql.NullInt32
	TransportFeeQuarterValue sql.NullInt32
	CreatedAt                sql.NullTime
	LastUpdatedAt            sql.NullTime
	EnrollmentID             sql.NullInt64
}

func (q *Queries) GetUnpaidAttendancesByTeacherId(ctx context.Context, arg GetUnpaidAttendancesByTeacherIdParams) ([]GetUnpaidAttendancesByTeacherIdRow, error) {
	rows, err := q.db.QueryContext(ctx, getUnpaidAttendancesByTeacherId, arg.StartDate, arg.EndDate, arg.TeacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUnpaidAttendancesByTeacherIdRow
	for rows.Next() {
		var i GetUnpaidAttendancesByTeacherIdRow
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
			&i.Class.AutoOweAttendanceToken,
			&i.Class.IsDeactivated,
			&i.TeacherSpecialFee,
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
			&i.ID,
			&i.Quota,
			&i.CourseFeeQuarterValue,
			&i.TransportFeeQuarterValue,
			&i.CreatedAt,
			&i.LastUpdatedAt,
			&i.EnrollmentID,
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
	ClassID               int64
	TeacherID             int64
	StudentID             int64
	TokenID               sql.NullInt64
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
	query := setAttendancesIsPaidStatusByIds
	var queryParams []interface{}
	queryParams = append(queryParams, arg.IsPaid)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := q.db.ExecContext(ctx, query, queryParams...)
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
	ClassID               int64
	TeacherID             int64
	StudentID             int64
	TokenID               sql.NullInt64
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

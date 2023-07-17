// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: payment_queries.sql

package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

const countEnrollmentPayments = `-- name: CountEnrollmentPayments :one
SELECT Count(id) AS total from enrollment_payment
`

func (q *Queries) CountEnrollmentPayments(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countEnrollmentPayments)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const countStudentLearningTokens = `-- name: CountStudentLearningTokens :one
SELECT Count(id) AS total FROM student_learning_token
`

func (q *Queries) CountStudentLearningTokens(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countStudentLearningTokens)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const deleteEnrollmentPaymentById = `-- name: DeleteEnrollmentPaymentById :exec
DELETE FROM enrollment_payment
WHERE id = ?
`

func (q *Queries) DeleteEnrollmentPaymentById(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteEnrollmentPaymentById, id)
	return err
}

const deleteEnrollmentPaymentsByIds = `-- name: DeleteEnrollmentPaymentsByIds :exec
DELETE FROM enrollment_payment
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) DeleteEnrollmentPaymentsByIds(ctx context.Context, ids []int64) error {
	sql := deleteEnrollmentPaymentsByIds
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

const deleteStudentLearningTokenById = `-- name: DeleteStudentLearningTokenById :exec
DELETE FROM student_learning_token
WHERE id = ?
`

func (q *Queries) DeleteStudentLearningTokenById(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteStudentLearningTokenById, id)
	return err
}

const deleteStudentLearningTokensByIds = `-- name: DeleteStudentLearningTokensByIds :exec
DELETE FROM student_learning_token
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) DeleteStudentLearningTokensByIds(ctx context.Context, ids []int64) error {
	sql := deleteStudentLearningTokensByIds
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

const deleteTeacherSalaryById = `-- name: DeleteTeacherSalaryById :exec
DELETE FROM teacher_salary
WHERE id = ?
`

func (q *Queries) DeleteTeacherSalaryById(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteTeacherSalaryById, id)
	return err
}

const getEnrollmentPaymentById = `-- name: GetEnrollmentPaymentById :one
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, value, value_penalty, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE ep.id = ? LIMIT 1
`

type GetEnrollmentPaymentByIdRow struct {
	EnrollmentPaymentID int64
	PaymentDate         time.Time
	BalanceTopUp        int32
	Value               int32
	ValuePenalty        int32
	StudentEnrollmentID int64
	StudentID           int64
	StudentUsername     string
	StudentDetail       json.RawMessage
	Class               Class
	Course              Course
	Instrument          Instrument
	Grade               Grade
}

// ============================== ENROLLMENT_PAYMENT ==============================
func (q *Queries) GetEnrollmentPaymentById(ctx context.Context, id int64) (GetEnrollmentPaymentByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getEnrollmentPaymentById, id)
	var i GetEnrollmentPaymentByIdRow
	err := row.Scan(
		&i.EnrollmentPaymentID,
		&i.PaymentDate,
		&i.BalanceTopUp,
		&i.Value,
		&i.ValuePenalty,
		&i.StudentEnrollmentID,
		&i.StudentID,
		&i.StudentUsername,
		&i.StudentDetail,
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
	)
	return i, err
}

const getEnrollmentPayments = `-- name: GetEnrollmentPayments :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, value, value_penalty, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
ORDER BY ep.id
LIMIT ? OFFSET ?
`

type GetEnrollmentPaymentsParams struct {
	Limit  int32
	Offset int32
}

type GetEnrollmentPaymentsRow struct {
	EnrollmentPaymentID int64
	PaymentDate         time.Time
	BalanceTopUp        int32
	Value               int32
	ValuePenalty        int32
	StudentEnrollmentID int64
	StudentID           int64
	StudentUsername     string
	StudentDetail       json.RawMessage
	Class               Class
	Course              Course
	Instrument          Instrument
	Grade               Grade
}

func (q *Queries) GetEnrollmentPayments(ctx context.Context, arg GetEnrollmentPaymentsParams) ([]GetEnrollmentPaymentsRow, error) {
	rows, err := q.db.QueryContext(ctx, getEnrollmentPayments, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetEnrollmentPaymentsRow
	for rows.Next() {
		var i GetEnrollmentPaymentsRow
		if err := rows.Scan(
			&i.EnrollmentPaymentID,
			&i.PaymentDate,
			&i.BalanceTopUp,
			&i.Value,
			&i.ValuePenalty,
			&i.StudentEnrollmentID,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
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

const getEnrollmentPaymentsByIds = `-- name: GetEnrollmentPaymentsByIds :many
SELECT ep.id AS enrollment_payment_id, payment_date, balance_top_up, value, value_penalty, se.id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name
FROM enrollment_payment AS ep
    JOIN student_enrollment AS se ON ep.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE ep.id IN (/*SLICE:ids*/?)
`

type GetEnrollmentPaymentsByIdsRow struct {
	EnrollmentPaymentID int64
	PaymentDate         time.Time
	BalanceTopUp        int32
	Value               int32
	ValuePenalty        int32
	StudentEnrollmentID int64
	StudentID           int64
	StudentUsername     string
	StudentDetail       json.RawMessage
	Class               Class
	Course              Course
	Instrument          Instrument
	Grade               Grade
}

func (q *Queries) GetEnrollmentPaymentsByIds(ctx context.Context, ids []int64) ([]GetEnrollmentPaymentsByIdsRow, error) {
	sql := getEnrollmentPaymentsByIds
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
	var items []GetEnrollmentPaymentsByIdsRow
	for rows.Next() {
		var i GetEnrollmentPaymentsByIdsRow
		if err := rows.Scan(
			&i.EnrollmentPaymentID,
			&i.PaymentDate,
			&i.BalanceTopUp,
			&i.Value,
			&i.ValuePenalty,
			&i.StudentEnrollmentID,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
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

const getStudentLearningTokenById = `-- name: GetStudentLearningTokenById :one
SELECT slt.id AS student_learning_token_id, quota, quota_bonus, course_fee_value, transport_fee_value, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE slt.id = ? LIMIT 1
`

type GetStudentLearningTokenByIdRow struct {
	StudentLearningTokenID int64
	Quota                  int32
	QuotaBonus             int32
	CourseFeeValue         int32
	TransportFeeValue      int32
	LastUpdatedAt          time.Time
	StudentEnrollmentID    int64
	StudentID              int64
	StudentUsername        string
	StudentDetail          json.RawMessage
	Class                  Class
	Course                 Course
	Instrument             Instrument
	Grade                  Grade
}

// ============================== STUDENT_LEARNING_TOKEN ==============================
func (q *Queries) GetStudentLearningTokenById(ctx context.Context, id int64) (GetStudentLearningTokenByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getStudentLearningTokenById, id)
	var i GetStudentLearningTokenByIdRow
	err := row.Scan(
		&i.StudentLearningTokenID,
		&i.Quota,
		&i.QuotaBonus,
		&i.CourseFeeValue,
		&i.TransportFeeValue,
		&i.LastUpdatedAt,
		&i.StudentEnrollmentID,
		&i.StudentID,
		&i.StudentUsername,
		&i.StudentDetail,
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
	)
	return i, err
}

const getStudentLearningTokens = `-- name: GetStudentLearningTokens :many
SELECT slt.id AS student_learning_token_id, quota, quota_bonus, course_fee_value, transport_fee_value, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
ORDER BY slt.id
LIMIT ? OFFSET ?
`

type GetStudentLearningTokensParams struct {
	Limit  int32
	Offset int32
}

type GetStudentLearningTokensRow struct {
	StudentLearningTokenID int64
	Quota                  int32
	QuotaBonus             int32
	CourseFeeValue         int32
	TransportFeeValue      int32
	LastUpdatedAt          time.Time
	StudentEnrollmentID    int64
	StudentID              int64
	StudentUsername        string
	StudentDetail          json.RawMessage
	Class                  Class
	Course                 Course
	Instrument             Instrument
	Grade                  Grade
}

func (q *Queries) GetStudentLearningTokens(ctx context.Context, arg GetStudentLearningTokensParams) ([]GetStudentLearningTokensRow, error) {
	rows, err := q.db.QueryContext(ctx, getStudentLearningTokens, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetStudentLearningTokensRow
	for rows.Next() {
		var i GetStudentLearningTokensRow
		if err := rows.Scan(
			&i.StudentLearningTokenID,
			&i.Quota,
			&i.QuotaBonus,
			&i.CourseFeeValue,
			&i.TransportFeeValue,
			&i.LastUpdatedAt,
			&i.StudentEnrollmentID,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
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

const getStudentLearningTokensByEnrollmentId = `-- name: GetStudentLearningTokensByEnrollmentId :many
SELECT slt.id AS student_learning_token_id, quota, quota_bonus, course_fee_value, transport_fee_value, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE slt.enrollment_id = ?
`

type GetStudentLearningTokensByEnrollmentIdRow struct {
	StudentLearningTokenID int64
	Quota                  int32
	QuotaBonus             int32
	CourseFeeValue         int32
	TransportFeeValue      int32
	LastUpdatedAt          time.Time
	StudentEnrollmentID    int64
	StudentID              int64
	StudentUsername        string
	StudentDetail          json.RawMessage
	Class                  Class
	Course                 Course
	Instrument             Instrument
	Grade                  Grade
}

func (q *Queries) GetStudentLearningTokensByEnrollmentId(ctx context.Context, enrollmentID int64) ([]GetStudentLearningTokensByEnrollmentIdRow, error) {
	rows, err := q.db.QueryContext(ctx, getStudentLearningTokensByEnrollmentId, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetStudentLearningTokensByEnrollmentIdRow
	for rows.Next() {
		var i GetStudentLearningTokensByEnrollmentIdRow
		if err := rows.Scan(
			&i.StudentLearningTokenID,
			&i.Quota,
			&i.QuotaBonus,
			&i.CourseFeeValue,
			&i.TransportFeeValue,
			&i.LastUpdatedAt,
			&i.StudentEnrollmentID,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
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

const getStudentLearningTokensByIds = `-- name: GetStudentLearningTokensByIds :many
SELECT slt.id AS student_learning_token_id, quota, quota_bonus, course_fee_value, transport_fee_value, last_updated_at, slt.enrollment_id AS student_enrollment_id,
    se.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name
FROM student_learning_token AS slt
    JOIN student_enrollment AS se ON slt.enrollment_id = se.id

    JOIN user AS user_student ON se.student_id = user_student.id
    
    JOIN class on se.class_id = class.id
    JOIN course ON class.course_id = course.id
    JOIN instrument ON course.instrument_id = instrument.id
    JOIN grade ON course.grade_id = grade.id
WHERE slt.id IN (/*SLICE:ids*/?)
`

type GetStudentLearningTokensByIdsRow struct {
	StudentLearningTokenID int64
	Quota                  int32
	QuotaBonus             int32
	CourseFeeValue         int32
	TransportFeeValue      int32
	LastUpdatedAt          time.Time
	StudentEnrollmentID    int64
	StudentID              int64
	StudentUsername        string
	StudentDetail          json.RawMessage
	Class                  Class
	Course                 Course
	Instrument             Instrument
	Grade                  Grade
}

func (q *Queries) GetStudentLearningTokensByIds(ctx context.Context, ids []int64) ([]GetStudentLearningTokensByIdsRow, error) {
	sql := getStudentLearningTokensByIds
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
	var items []GetStudentLearningTokensByIdsRow
	for rows.Next() {
		var i GetStudentLearningTokensByIdsRow
		if err := rows.Scan(
			&i.StudentLearningTokenID,
			&i.Quota,
			&i.QuotaBonus,
			&i.CourseFeeValue,
			&i.TransportFeeValue,
			&i.LastUpdatedAt,
			&i.StudentEnrollmentID,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
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

const getTeacherSalaries = `-- name: GetTeacherSalaries :many
SELECT ts.id AS teacher_salary_id, profit_sharing_percentage, added_at,
    presence.id AS presence_id, date, used_student_token_quota, duration,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    class.id AS class_id, course_id, instrument.id, instrument.name, grade.id, grade.name,
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
ORDER BY ts.id
`

type GetTeacherSalariesRow struct {
	TeacherSalaryID         int64
	ProfitSharingPercentage float64
	AddedAt                 time.Time
	PresenceID              int64
	Date                    time.Time
	UsedStudentTokenQuota   float64
	Duration                int32
	TeacherID               sql.NullInt64
	TeacherUsername         sql.NullString
	TeacherDetail           []byte
	ClassID                 sql.NullInt64
	CourseID                sql.NullInt64
	Instrument              Instrument
	Grade                   Grade
	StudentID               sql.NullInt64
	StudentUsername         sql.NullString
	StudentDetail           []byte
}

func (q *Queries) GetTeacherSalaries(ctx context.Context) ([]GetTeacherSalariesRow, error) {
	rows, err := q.db.QueryContext(ctx, getTeacherSalaries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTeacherSalariesRow
	for rows.Next() {
		var i GetTeacherSalariesRow
		if err := rows.Scan(
			&i.TeacherSalaryID,
			&i.ProfitSharingPercentage,
			&i.AddedAt,
			&i.PresenceID,
			&i.Date,
			&i.UsedStudentTokenQuota,
			&i.Duration,
			&i.TeacherID,
			&i.TeacherUsername,
			&i.TeacherDetail,
			&i.ClassID,
			&i.CourseID,
			&i.Instrument.ID,
			&i.Instrument.Name,
			&i.Grade.ID,
			&i.Grade.Name,
			&i.StudentID,
			&i.StudentUsername,
			&i.StudentDetail,
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

const getTeacherSalaryById = `-- name: GetTeacherSalaryById :one
SELECT id, presence_id, profit_sharing_percentage, added_at FROM teacher_salary
WHERE id = ? LIMIT 1
`

// ============================== TEACHER_SALARY ==============================
func (q *Queries) GetTeacherSalaryById(ctx context.Context, id int64) (TeacherSalary, error) {
	row := q.db.QueryRowContext(ctx, getTeacherSalaryById, id)
	var i TeacherSalary
	err := row.Scan(
		&i.ID,
		&i.PresenceID,
		&i.ProfitSharingPercentage,
		&i.AddedAt,
	)
	return i, err
}

const insertEnrollmentPayment = `-- name: InsertEnrollmentPayment :execlastid
INSERT INTO enrollment_payment (
    payment_date, balance_top_up, value, value_penalty, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?
)
`

type InsertEnrollmentPaymentParams struct {
	PaymentDate  time.Time
	BalanceTopUp int32
	Value        int32
	ValuePenalty int32
	EnrollmentID sql.NullInt64
}

func (q *Queries) InsertEnrollmentPayment(ctx context.Context, arg InsertEnrollmentPaymentParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertEnrollmentPayment,
		arg.PaymentDate,
		arg.BalanceTopUp,
		arg.Value,
		arg.ValuePenalty,
		arg.EnrollmentID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const insertStudentLearningToken = `-- name: InsertStudentLearningToken :execlastid
INSERT INTO student_learning_token (
    quota, quota_bonus, course_fee_value, transport_fee_value, enrollment_id
) VALUES (
    ?, ?, ?, ?, ?
)
`

type InsertStudentLearningTokenParams struct {
	Quota             int32
	QuotaBonus        int32
	CourseFeeValue    int32
	TransportFeeValue int32
	EnrollmentID      int64
}

func (q *Queries) InsertStudentLearningToken(ctx context.Context, arg InsertStudentLearningTokenParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertStudentLearningToken,
		arg.Quota,
		arg.QuotaBonus,
		arg.CourseFeeValue,
		arg.TransportFeeValue,
		arg.EnrollmentID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const insertTeacherSalary = `-- name: InsertTeacherSalary :execlastid
INSERT INTO teacher_salary (
    presence_id, profit_sharing_percentage, added_at
) VALUES (
    ?, ?, ?
)
`

type InsertTeacherSalaryParams struct {
	PresenceID              int64
	ProfitSharingPercentage float64
	AddedAt                 time.Time
}

func (q *Queries) InsertTeacherSalary(ctx context.Context, arg InsertTeacherSalaryParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertTeacherSalary, arg.PresenceID, arg.ProfitSharingPercentage, arg.AddedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const updateEnrollmentPayment = `-- name: UpdateEnrollmentPayment :exec
UPDATE enrollment_payment SET payment_date = ?, balance_top_up = ?, value = ?, value_penalty = ?
WHERE id = ?
`

type UpdateEnrollmentPaymentParams struct {
	PaymentDate  time.Time
	BalanceTopUp int32
	Value        int32
	ValuePenalty int32
	ID           int64
}

func (q *Queries) UpdateEnrollmentPayment(ctx context.Context, arg UpdateEnrollmentPaymentParams) error {
	_, err := q.db.ExecContext(ctx, updateEnrollmentPayment,
		arg.PaymentDate,
		arg.BalanceTopUp,
		arg.Value,
		arg.ValuePenalty,
		arg.ID,
	)
	return err
}

const updateStudentLearningToken = `-- name: UpdateStudentLearningToken :exec
UPDATE student_learning_token SET quota = ?, quota_bonus = ?, course_fee_value = ?, transport_fee_value = ?
WHERE id = ?
`

type UpdateStudentLearningTokenParams struct {
	Quota             int32
	QuotaBonus        int32
	CourseFeeValue    int32
	TransportFeeValue int32
	ID                int64
}

func (q *Queries) UpdateStudentLearningToken(ctx context.Context, arg UpdateStudentLearningTokenParams) error {
	_, err := q.db.ExecContext(ctx, updateStudentLearningToken,
		arg.Quota,
		arg.QuotaBonus,
		arg.CourseFeeValue,
		arg.TransportFeeValue,
		arg.ID,
	)
	return err
}

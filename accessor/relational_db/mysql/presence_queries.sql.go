// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: presence_queries.sql

package mysql

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

const countPresences = `-- name: CountPresences :one
SELECT Count(id) AS total FROM presence
WHERE
    (date >= ? AND date <= ?)
    AND (class_id = ? OR ? = false)
    AND (student_id = ? OR ? = false)
`

type CountPresencesParams struct {
	StartDate        time.Time
	EndDate          time.Time
	ClassID          sql.NullInt64
	UseClassFilter   interface{}
	StudentID        sql.NullInt64
	UseStudentFilter interface{}
}

func (q *Queries) CountPresences(ctx context.Context, arg CountPresencesParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countPresences,
		arg.StartDate,
		arg.EndDate,
		arg.ClassID,
		arg.UseClassFilter,
		arg.StudentID,
		arg.UseStudentFilter,
	)
	var total int64
	err := row.Scan(&total)
	return total, err
}

const countPresencesByIds = `-- name: CountPresencesByIds :one
SELECT Count(id) AS total FROM presence
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) CountPresencesByIds(ctx context.Context, ids []int64) (int64, error) {
	sql := countPresencesByIds
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

const deletePresenceById = `-- name: DeletePresenceById :exec
DELETE FROM presence
WHERE id = ?
`

func (q *Queries) DeletePresenceById(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deletePresenceById, id)
	return err
}

const deletePresencesByIds = `-- name: DeletePresencesByIds :exec
DELETE FROM presence
WHERE id IN (/*SLICE:ids*/?)
`

func (q *Queries) DeletePresencesByIds(ctx context.Context, ids []int64) error {
	sql := deletePresencesByIds
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

const editPresences = `-- name: EditPresences :exec
UPDATE presence
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, teacher_id = ?
WHERE id IN (/*SLICE:ids*/?)
`

type EditPresencesParams struct {
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	TeacherID             sql.NullInt64
	Ids                   []int64
}

func (q *Queries) EditPresences(ctx context.Context, arg EditPresencesParams) error {
	sql := editPresences
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

const getPresenceById = `-- name: GetPresenceById :one
SELECT presence.id AS presence_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    slt.id, slt.quota, slt.course_fee_value, slt.transport_fee_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE presence.id = ? LIMIT 1
`

type GetPresenceByIdRow struct {
	PresenceID            int64
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

func (q *Queries) GetPresenceById(ctx context.Context, id int64) (GetPresenceByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getPresenceById, id)
	var i GetPresenceByIdRow
	err := row.Scan(
		&i.PresenceID,
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

const getPresenceIdsOfSameClassAndDate = `-- name: GetPresenceIdsOfSameClassAndDate :many
WITH ref_presence AS (
    SELECT id, date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id
    FROM presence
    WHERE presence.id = ?
)
SELECT presence.id AS id, presence.is_paid AS is_paid
FROM presence
    JOIN ref_presence ON presence.class_id = ref_presence.class_id AND presence.date = ref_presence.date
ORDER by presence.id
`

type GetPresenceIdsOfSameClassAndDateRow struct {
	ID     int64
	IsPaid int32
}

// ============================== PRESENCE ==============================
func (q *Queries) GetPresenceIdsOfSameClassAndDate(ctx context.Context, id int64) ([]GetPresenceIdsOfSameClassAndDateRow, error) {
	rows, err := q.db.QueryContext(ctx, getPresenceIdsOfSameClassAndDate, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPresenceIdsOfSameClassAndDateRow
	for rows.Next() {
		var i GetPresenceIdsOfSameClassAndDateRow
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

const getPresences = `-- name: GetPresences :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    slt.id, slt.quota, slt.course_fee_value, slt.transport_fee_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE
    (presence.date >= ? AND presence.date <= ?)
    AND (class_id = ? OR ? = false)
    AND (student_id = ? OR ? = false)
ORDER BY class.id
LIMIT ? OFFSET ?
`

type GetPresencesParams struct {
	StartDate        time.Time
	EndDate          time.Time
	ClassID          sql.NullInt64
	UseClassFilter   interface{}
	StudentID        sql.NullInt64
	UseStudentFilter interface{}
	Limit            int32
	Offset           int32
}

type GetPresencesRow struct {
	PresenceID            int64
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

func (q *Queries) GetPresences(ctx context.Context, arg GetPresencesParams) ([]GetPresencesRow, error) {
	rows, err := q.db.QueryContext(ctx, getPresences,
		arg.StartDate,
		arg.EndDate,
		arg.ClassID,
		arg.UseClassFilter,
		arg.StudentID,
		arg.UseStudentFilter,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPresencesRow
	for rows.Next() {
		var i GetPresencesRow
		if err := rows.Scan(
			&i.PresenceID,
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

const getPresencesByIds = `-- name: GetPresencesByIds :many
SELECT presence.id AS presence_id, date, used_student_token_quota, duration, note, is_paid,
    class.id, class.transport_fee, class.teacher_id, class.course_id, class.is_deactivated, course.id, course.default_fee, course.default_duration_minute, course.instrument_id, course.grade_id, instrument.id, instrument.name, grade.id, grade.name,
    presence.teacher_id AS teacher_id, user_teacher.username AS teacher_username, user_teacher.user_detail AS teacher_detail,
    presence.student_id AS student_id, user_student.username AS student_username, user_student.user_detail AS student_detail,
    class.teacher_id AS class_teacher_id, user_class_teacher.username AS class_teacher_username, user_class_teacher.user_detail AS class_teacher_detail,
    slt.id, slt.quota, slt.course_fee_value, slt.transport_fee_value, slt.created_at, slt.last_updated_at, slt.enrollment_id
FROM presence
    LEFT JOIN teacher ON presence.teacher_id = teacher.id
    LEFT JOIN user AS user_teacher ON teacher.user_id = user_teacher.id
    LEFT JOIN user AS user_student ON presence.student_id = user_student.id

    LEFT JOIN class on presence.class_id = class.id
    LEFT JOIN course ON course_id = course.id
    LEFT JOIN instrument ON course.instrument_id = instrument.id
    LEFT JOIN grade ON course.grade_id = grade.id
    
    LEFT JOIN teacher AS class_teacher ON class.teacher_id = class_teacher.id
    LEFT JOIN user AS user_class_teacher ON class_teacher.user_id = user_class_teacher.id

    JOIN student_learning_token as slt ON presence.token_id = slt.id
WHERE presence.id IN (/*SLICE:ids*/?)
`

type GetPresencesByIdsRow struct {
	PresenceID            int64
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

func (q *Queries) GetPresencesByIds(ctx context.Context, ids []int64) ([]GetPresencesByIdsRow, error) {
	sql := getPresencesByIds
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
	var items []GetPresencesByIdsRow
	for rows.Next() {
		var i GetPresencesByIdsRow
		if err := rows.Scan(
			&i.PresenceID,
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

const insertPresence = `-- name: InsertPresence :execlastid
INSERT INTO presence (
    date, used_student_token_quota, duration, note, is_paid, class_id, teacher_id, student_id, token_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`

type InsertPresenceParams struct {
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

func (q *Queries) InsertPresence(ctx context.Context, arg InsertPresenceParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertPresence,
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

const updatePresence = `-- name: UpdatePresence :exec
UPDATE presence
SET date = ?, used_student_token_quota = ?, duration = ?, note = ?, is_paid = ?, class_id = ?, teacher_id = ?, student_id = ?, token_id = ?
WHERE id = ?
`

type UpdatePresenceParams struct {
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

func (q *Queries) UpdatePresence(ctx context.Context, arg UpdatePresenceParams) error {
	_, err := q.db.ExecContext(ctx, updatePresence,
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

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package mysql

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Class struct {
	ID            int64
	TransportFee  int32
	TeacherID     sql.NullInt64
	CourseID      int64
	IsDeactivated int32
}

type Course struct {
	ID                    int64
	DefaultFee            int32
	DefaultDurationMinute int32
	InstrumentID          int64
	GradeID               int64
}

type EnrollmentPayment struct {
	ID                int64
	PaymentDate       time.Time
	BalanceTopUp      int32
	CourseFeeValue    int32
	TransportFeeValue int32
	PenaltyFeeValue   int32
	EnrollmentID      sql.NullInt64
}

type Grade struct {
	ID   int64
	Name string
}

type Instrument struct {
	ID   int64
	Name string
}

type Presence struct {
	ID                    int64
	Date                  time.Time
	UsedStudentTokenQuota float64
	Duration              int32
	Note                  string
	ClassID               sql.NullInt64
	TeacherID             sql.NullInt64
	StudentID             sql.NullInt64
	TokenID               int64
}

type Student struct {
	ID     int64
	UserID int64
}

type StudentEnrollment struct {
	ID        int64
	StudentID int64
	ClassID   int64
	IsDeleted int32
}

type StudentLearningToken struct {
	ID                int64
	Quota             float64
	CourseFeeValue    int32
	TransportFeeValue int32
	CreatedAt         time.Time
	LastUpdatedAt     time.Time
	EnrollmentID      int64
}

type Teacher struct {
	ID     int64
	UserID int64
}

type TeacherSalary struct {
	ID                      int64
	PresenceID              int64
	ProfitSharingPercentage float64
	AddedAt                 time.Time
}

type TeacherSpecialFee struct {
	ID        int64
	Fee       int32
	TeacherID int64
	CourseID  int64
}

type User struct {
	ID            int64
	Username      string
	Email         sql.NullString
	UserDetail    json.RawMessage
	PrivilegeType int32
	IsDeactivated int32
	CreatedAt     sql.NullTime
}

type UserCredential struct {
	UserID   int64
	Username string
	Email    sql.NullString
	Password string
}

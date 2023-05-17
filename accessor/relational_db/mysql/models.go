// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2

package mysql

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Class struct {
	ID                  int64
	DefaultTransportFee int64
	TeacherID           sql.NullInt64
	CourseID            int64
	IsDeactivated       int32
}

type Course struct {
	ID                    int64
	DefaultFee            int64
	DefaultDurationMinute int32
	InstrumentID          int64
	GradeID               int64
}

type EnrollmentPayment struct {
	ID           int64
	PaymentDate  time.Time
	BalanceTopUp int32
	Value        int32
	ValuePenalty int32
	EnrollmentID sql.NullInt64
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
	ClassID               sql.NullInt64
	TeacherID             sql.NullInt64
	TokenID               int64
}

type Student struct {
	ID     int64
	UserID int64
}

type StudentAttend struct {
	StudentID  int64
	PresenceID int64
}

type StudentEnrollment struct {
	ID        int64
	StudentID int64
	ClassID   int64
}

type StudentLearningToken struct {
	ID                int64
	Quota             int32
	QuotaBonus        int32
	CourseFeeValue    int32
	TransportFeeValue int32
	LastUpdatedAt     time.Time
	EnrollmentID      sql.NullInt64
}

type Teacher struct {
	ID     int64
	UserID int64
}

type TeacherSalary struct {
	ID                      int64
	PresenceID              int64
	CourseFeeValue          int32
	TransportFeeValue       int32
	ProfitSharingPercentage float64
	AddedAt                 time.Time
}

type TeacherSpecialFee struct {
	ID        int64
	Fee       int64
	TeacherID int64
	CourseID  int64
}

type User struct {
	ID            int64
	Username      string
	Email         string
	UserDetail    json.RawMessage
	PrivilegeType int32
	IsDeactivated int32
	CreatedAt     sql.NullTime
}

type UserCredential struct {
	UserID   int64
	Email    string
	Password string
}

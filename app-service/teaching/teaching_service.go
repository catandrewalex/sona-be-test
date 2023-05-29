package teaching

import (
	"context"

	"sonamusica-backend/app-service/identity"
	"sonamusica-backend/app-service/util"
)

type Teacher struct {
	TeacherID TeacherID     `json:"teacherId"`
	User      identity.User `json:"user"`
}

type Student struct {
	StudentID StudentID     `json:"studentId"`
	User      identity.User `json:"user"`
}

type Instrument struct {
	ID   InstrumentID `json:"id"`
	Name string       `json:"name"`
}

type Grade struct {
	ID   GradeID `json:"id"`
	Name string  `json:"name"`
}

type TeacherID int64
type StudentID int64
type InstrumentID int64
type GradeID int64
type CourseID int64
type ClassID int64
type StudentEnrollmentID int64
type StudentLearningTokenID int64
type TeacherSpecialFeeID int64
type PresenceID int64

const (
	TeacherID_None TeacherID = iota
)
const (
	StudentID_None StudentID = iota
)
const (
	InstrumentID_None InstrumentID = iota
)
const (
	GradeID_None GradeID = iota
)
const (
	CourseID_None CourseID = iota
)
const (
	ClassID_None ClassID = iota
)
const (
	StudentEnrollmentID_None StudentEnrollmentID = iota
)
const (
	StudentLearningTokenID_None StudentLearningTokenID = iota
)
const (
	TeacherSpecialFeeID_None TeacherSpecialFeeID = iota
)
const (
	PresenceID_None PresenceID = iota
)

type TeachingService interface {
	GetTeachers(ctx context.Context, pagination util.PaginationSpec) (GetTeachersResult, error)
	GetTeacherById(ctx context.Context, id TeacherID) (Teacher, error)
	GetTeachersByIds(ctx context.Context, ids []TeacherID) ([]Teacher, error)
	InsertTeachers(ctx context.Context, userIDs []identity.UserID) ([]TeacherID, error)
	InsertTeachersWithNewUsers(ctx context.Context, specs []identity.InsertUserSpec) ([]TeacherID, error)

	GetStudents(ctx context.Context, pagination util.PaginationSpec) (GetStudentsResult, error)
	GetStudentById(ctx context.Context, id StudentID) (Student, error)
	GetStudentsByIds(ctx context.Context, ids []StudentID) ([]Student, error)
}

type GetTeachersResult struct {
	Teachers         []Teacher
	PaginationResult util.PaginationResult
}

type GetStudentsResult struct {
	Students         []Student
	PaginationResult util.PaginationResult
}

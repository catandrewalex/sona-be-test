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

type TeachingService interface {
	GetTeachers(ctx context.Context, pagination util.PaginationSpec) (GetTeachersResult, error)
	GetTeacherByUserID(ctx context.Context, userID identity.UserID) (Teacher, error)

	GetStudents(ctx context.Context, pagination util.PaginationSpec) (GetStudentsResult, error)
	GetStudentByUserID(ctx context.Context, userID identity.UserID) (Student, error)
}

type GetTeachersResult struct {
	Teachers         []Teacher
	PaginationResult util.PaginationResult
}

type GetStudentsResult struct {
	Students         []Student
	PaginationResult util.PaginationResult
}
